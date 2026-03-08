param(
    [string]$Version,
    [switch]$ForceDownload,
    [switch]$SkipVueBuild
)

$ErrorActionPreference = "Stop"

# Add Go to PATH if not already present
$goPath = "C:\Program Files\Go\bin"
if (Test-Path $goPath) {
    $env:PATH = "$goPath;$env:PATH"
}

# Set Go module proxy for China
$env:GOPROXY = "https://goproxy.cn,direct"

$PROJECT_DIR = if ($PSScriptRoot) { $PSScriptRoot } else { (Get-Location).Path }
$MANIFEST_FILE = Join-Path $PROJECT_DIR "manifest"

if ($Version) {
    $APP_VERSION = $Version.Trim()
    Write-Host "Using version: $APP_VERSION" -ForegroundColor Cyan
} else {
    $Version = ""
    $lines = Get-Content $MANIFEST_FILE
    foreach ($line in $lines) {
        if ($line -match "^version\s*=\s*(\S+)") {
            $Version = $matches[1].Trim()
            break
        }
    }
    if (-not $Version) {
        Write-Host "Error: Cannot read version from manifest" -ForegroundColor Red
        exit 1
    }
    $APP_VERSION = $Version
    Write-Host "Using manifest version: $APP_VERSION" -ForegroundColor Cyan
}

$BUILD_DIR = Join-Path $PROJECT_DIR ".local-build"
$VERSION_FILE = Join-Path $BUILD_DIR "versions.json"
$FNPACK_URL = "https://static2.fnnas.com/fnpack/fnpack-1.2.1-windows-amd64"

function Get-VersionInfo {
    if (Test-Path $VERSION_FILE) {
        try { return Get-Content $VERSION_FILE -Raw | ConvertFrom-Json } catch { return @{} }
    }
    return @{}
}

function Save-VersionInfo {
    param($Component, $Version)
    $versions = Get-VersionInfo
    if ($versions -is [System.Management.Automation.PSCustomObject]) {
        $hash = @{}
        $versions.PSObject.Properties | ForEach-Object { $hash[$_.Name] = $_.Value }
        $versions = $hash
    }
    $versions[$Component] = $Version
    $versions | ConvertTo-Json -Depth 10 | Set-Content $VERSION_FILE -Force
}

function Test-VersionMatch {
    param($Component, $ExpectedVersion)
    $versions = Get-VersionInfo
    return ($versions.$Component -eq $ExpectedVersion)
}

function Get-FileDirect {
    param($Url, $OutFile, $Description, $Component, $Version)
    if ((-not $ForceDownload) -and (Test-Path $OutFile) -and ((Get-Item $OutFile).Length -gt 0)) {
        if (Test-VersionMatch -Component $Component -ExpectedVersion $Version) {
            Write-Host "  Using cached $Description (version $Version)" -ForegroundColor Green
            return $true
        }
    }
    Write-Host "  Downloading $Description..." -ForegroundColor Yellow
    try {
        $ProgressPreference = 'SilentlyContinue'
        Invoke-WebRequest -Uri $Url -OutFile $OutFile -UseBasicParsing
        if ((Test-Path $OutFile) -and (Get-Item $OutFile).Length -gt 0) {
            Write-Host "  Download $Description success" -ForegroundColor Green
            Save-VersionInfo -Component $Component -Version $Version
            return $true
        }
    } catch {
        Write-Host "  Error: Download $Description failed - $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
    return $false
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Dockpit - Build" -ForegroundColor Cyan
Write-Host "  Version: $APP_VERSION" -ForegroundColor Gray
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "[1/5] Setup build directory..." -ForegroundColor Yellow
$dirsToClean = @("app\ui\assets", "app\ui\images", "app\server", "cmd", "config", "wizard")
foreach ($dir in $dirsToClean) {
    $cleanPath = Join-Path $BUILD_DIR $dir
    if (Test-Path $cleanPath) {
        Remove-Item -Path "$cleanPath\*" -Recurse -Force -ErrorAction SilentlyContinue
    }
}
@("app\server", "app\ui", "cmd", "config", "wizard") | ForEach-Object {
    New-Item -ItemType Directory -Force -Path (Join-Path $BUILD_DIR $_) | Out-Null
}
Write-Host "  Build directory ready" -ForegroundColor Green

Write-Host "[2/5] Build Vue frontend..." -ForegroundColor Yellow
$UI_DIR = Join-Path $PROJECT_DIR "app\ui"
if (-not $SkipVueBuild) {
    if (Test-Path "$UI_DIR\package.json") {
        Push-Location $UI_DIR
        if (-not (Test-Path "node_modules")) {
            Write-Host "  Installing npm dependencies..." -ForegroundColor Yellow
            $npmInstall = npm install 2>&1
            if ($LASTEXITCODE -ne 0) {
                Write-Host "  Error: npm install failed" -ForegroundColor Red
                Write-Host $npmInstall
                Pop-Location
                exit 1
            }
        }
        Write-Host "  Building Vue app..." -ForegroundColor Yellow
        $npmBuild = npm run build 2>&1
        $buildExitCode = $LASTEXITCODE
        Pop-Location
        if ($buildExitCode -ne 0) {
            Write-Host "  Error: Vue build failed" -ForegroundColor Red
            Write-Host $npmBuild
            exit 1
        }
        if (Test-Path "$UI_DIR\dist") {
            Write-Host "  Vue build complete" -ForegroundColor Green
        } else {
            Write-Host "  Error: Vue build output not found" -ForegroundColor Red
            exit 1
        }
    }
} else {
    Write-Host "  Skipping Vue build" -ForegroundColor Yellow
}

Write-Host "[3/5] Copy project files..." -ForegroundColor Yellow
Copy-Item "$PROJECT_DIR\cmd\*" "$BUILD_DIR\cmd\" -Recurse -Force
Copy-Item "$PROJECT_DIR\config\*" "$BUILD_DIR\config\" -Recurse -Force
Copy-Item "$PROJECT_DIR\wizard\*" "$BUILD_DIR\wizard\" -Recurse -Force

$manifestContent = Get-Content $MANIFEST_FILE -Raw -Encoding UTF8
$manifestContent = $manifestContent -replace "(?m)^version\s*=.*", "version = $APP_VERSION"
[System.IO.File]::WriteAllText("$BUILD_DIR\manifest", $manifestContent, [System.Text.Encoding]::UTF8)

@("LICENSE", "ICON.PNG", "ICON_256.PNG") | ForEach-Object {
    if (Test-Path "$PROJECT_DIR\$_") { Copy-Item "$PROJECT_DIR\$_" "$BUILD_DIR\" -Force }
}

$UI_DIST = Join-Path $UI_DIR "dist"
if (Test-Path $UI_DIST) {
    Copy-Item "$UI_DIST\*" "$BUILD_DIR\app\ui\" -Recurse -Force
    Write-Host "  Vue dist files copied" -ForegroundColor Green
}

if (Test-Path "$UI_DIR\config") { 
    Copy-Item "$UI_DIR\config" "$BUILD_DIR\app\ui\" -Force
    Write-Host "  UI config copied" -ForegroundColor Green
}
if (Test-Path "$UI_DIR\images") { 
    Copy-Item "$UI_DIR\images" "$BUILD_DIR\app\ui\" -Recurse -Force
    Write-Host "  UI images copied" -ForegroundColor Green
}

Write-Host "[4/5] Build and prepare server files..." -ForegroundColor Yellow
$serverDir = Join-Path $BUILD_DIR "app\server"
New-Item -ItemType Directory -Force -Path $serverDir | Out-Null

$SERVER_SRC = Join-Path $PROJECT_DIR "app\server"
if (Test-Path "$SERVER_SRC\main.go") {
    Push-Location $SERVER_SRC
    Write-Host "  Running go mod tidy..." -ForegroundColor Yellow
    go mod tidy
    $tidyExitCode = $LASTEXITCODE
    if ($tidyExitCode -ne 0) {
        Write-Host "  Warning: go mod tidy failed with exit code $tidyExitCode" -ForegroundColor Yellow
    } else {
        Write-Host "  go mod tidy completed" -ForegroundColor Green
    }
    
    Write-Host "  Building Go server..." -ForegroundColor Yellow
    
    $env:GOOS = "linux"
    $env:GOARCH = "arm64"
    go build -ldflags "-X dockpit/internal/handler.appVersion=$APP_VERSION" -o dockpit .
    $buildExitCode = $LASTEXITCODE
    Pop-Location
    
    if ($buildExitCode -ne 0) {
        Write-Host "  Error: Go build failed with exit code $buildExitCode" -ForegroundColor Red
        exit 1
    }
    
    if (Test-Path "$SERVER_SRC\dockpit") {
        Copy-Item "$SERVER_SRC\dockpit" "$serverDir\dockpit" -Force
        Write-Host "  Server executable built and copied (linux/arm64)" -ForegroundColor Green
    } else {
        Write-Host "  Error: Server executable not found" -ForegroundColor Red
        exit 1
    }
} else {
    Write-Host "  Error: Server source not found (main.go)" -ForegroundColor Red
    exit 1
}

Write-Host "[5/5] Build package..." -ForegroundColor Yellow
$FNPACK_VER = "1.2.1"
$FNPACK_FILE = $FNPACK_URL.Substring($FNPACK_URL.LastIndexOf('/') + 1)
$fnpackPath = Join-Path $BUILD_DIR $FNPACK_FILE
if ((-not $ForceDownload) -and (Test-Path $fnpackPath) -and (Test-VersionMatch -Component "fnpack" -ExpectedVersion $FNPACK_VER)) {
    Write-Host "  Using cached fnpack $FNPACK_VER" -ForegroundColor Green
} else {
    if (-not (Get-FileDirect -Url $FNPACK_URL -OutFile $fnpackPath -Description "fnpack" -Component "fnpack" -Version $FNPACK_VER)) { exit 1 }
}

Remove-Item "$BUILD_DIR\dockpit.fpk" -Force -ErrorAction SilentlyContinue
Push-Location $BUILD_DIR
Start-Process -FilePath $fnpackPath -ArgumentList "build" -Wait -NoNewWindow
$ok = Test-Path "dockpit.fpk"
Pop-Location

if ($ok) {
    Move-Item "$BUILD_DIR\dockpit.fpk" "$PROJECT_DIR\dockpit-$APP_VERSION.fpk" -Force
    Write-Host "  Build success!" -ForegroundColor Green
} else {
    Write-Host "  Error: Build failed" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  Build complete!" -ForegroundColor Green
Write-Host "  Output: dockpit-$APP_VERSION.fpk" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
