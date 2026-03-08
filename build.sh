#!/bin/bash

# Parse arguments
VERSION=""
FORCE_DOWNLOAD=false
SKIP_VUE_BUILD=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -Version)
            VERSION="$2"
            shift 2
            ;;
        -ForceDownload)
            FORCE_DOWNLOAD=true
            shift
            ;;
        -SkipVueBuild)
            SKIP_VUE_BUILD=true
            shift
            ;;
        *)
            shift
            ;;
    esac
done

set -e

# Set Go module proxy for China
export GOPROXY="https://goproxy.cn,direct"

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
MANIFEST_FILE="$PROJECT_DIR/manifest"

# Get version
if [ -n "$VERSION" ]; then
    APP_VERSION="$VERSION"
    echo "Using version: $APP_VERSION"
else
    APP_VERSION=$(grep "^version" "$MANIFEST_FILE" | cut -d'=' -f2 | tr -d ' ')
    echo "Using manifest version: $APP_VERSION"
fi

if [ -z "$APP_VERSION" ]; then
    echo "Error: Cannot read version from manifest"
    exit 1
fi

BUILD_DIR="$PROJECT_DIR/.local-build"
VERSION_FILE="$BUILD_DIR/versions.json"
FNPACK_VER="1.2.1"
FNPACK_URL="https://static2.fnnas.com/fnpack/fnpack-${FNPACK_VER}-linux-amd64"

# Helper functions
get_version_info() {
    if [ -f "$VERSION_FILE" ]; then
        cat "$VERSION_FILE"
    else
        echo "{}"
    fi
}

save_version_info() {
    local component="$1"
    local version="$2"
    local versions
    if [ -f "$VERSION_FILE" ]; then
        versions=$(cat "$VERSION_FILE")
    else
        versions="{}"
    fi
    echo "$versions" | python3 -c "import json,sys; d=json.load(sys.stdin); d['$component']='$version'; print(json.dumps(d))" > "$VERSION_FILE" 2>/dev/null || echo "{\"$component\": \"$version\"}" > "$VERSION_FILE"
}

test_version_match() {
    local component="$1"
    local expected="$2"
    local versions
    versions=$(get_version_info)
    local current
    current=$(echo "$versions" | python3 -c "import json,sys; print(json.load(sys.stdin).get('$component', ''))" 2>/dev/null || echo "")
    [ "$current" = "$expected" ]
}

get_file_direct() {
    local url="$1"
    local outfile="$2"
    local description="$3"
    local component="$4"
    local version="$5"

    if [ "$FORCE_DOWNLOAD" = false ] && [ -f "$outfile" ] && [ -s "$outfile" ]; then
        if test_version_match "$component" "$version"; then
            echo "  Using cached $description (version $version)"
            return 0
        fi
    fi

    echo "  Downloading $description..."
    if curl -fsSL "$url" -o "$outfile"; then
        if [ -f "$outfile" ] && [ -s "$outfile" ]; then
            echo "  Download $description success"
            save_version_info "$component" "$version"
            return 0
        fi
    fi
    echo "  Error: Download $description failed"
    return 1
}

echo "========================================"
echo "  Dockpit - Build"
echo "  Version: $APP_VERSION"
echo "========================================"

echo "[1/5] Setup build directory..."
# Clean specific directories
dirs_to_clean=("app/ui/assets" "app/ui/images" "app/server" "cmd" "config" "wizard")
for dir in "${dirs_to_clean[@]}"; do
    clean_path="$BUILD_DIR/$dir"
    if [ -d "$clean_path" ]; then
        rm -rf "$clean_path"/*
    fi
done

# Create directories
mkdir -p "$BUILD_DIR/app/server"
mkdir -p "$BUILD_DIR/app/ui"
mkdir -p "$BUILD_DIR/cmd"
mkdir -p "$BUILD_DIR/config"
mkdir -p "$BUILD_DIR/wizard"
echo "  Build directory ready"

echo "[2/5] Build Vue frontend..."
UI_DIR="$PROJECT_DIR/app/ui"
if [ "$SKIP_VUE_BUILD" = false ]; then
    if [ -f "$UI_DIR/package.json" ]; then
        cd "$UI_DIR"
        if [ ! -d "node_modules" ]; then
            echo "  Installing npm dependencies..."
            npm install
        fi
        echo "  Building Vue app..."
        npm run build
        if [ ! -d "$UI_DIR/dist" ]; then
            echo "  Error: Vue build output not found"
            exit 1
        fi
        echo "  Vue build complete"
    fi
else
    echo "  Skipping Vue build"
fi

echo "[3/5] Copy project files..."
cp -r "$PROJECT_DIR/cmd"/* "$BUILD_DIR/cmd/"
cp -r "$PROJECT_DIR/config"/* "$BUILD_DIR/config/"
cp -r "$PROJECT_DIR/wizard"/* "$BUILD_DIR/wizard/"

# Update manifest with version
sed "s/^version\s*=.*/version = $APP_VERSION/" "$MANIFEST_FILE" > "$BUILD_DIR/manifest"

# Copy additional files
for file in "LICENSE" "ICON.PNG" "ICON_256.PNG"; do
    if [ -f "$PROJECT_DIR/$file" ]; then
        cp "$PROJECT_DIR/$file" "$BUILD_DIR/"
    fi
done

# Copy Vue dist files
UI_DIST="$UI_DIR/dist"
if [ -d "$UI_DIST" ]; then
    cp -r "$UI_DIST"/* "$BUILD_DIR/app/ui/"
    echo "  Vue dist files copied"
fi

# Copy UI config and images
if [ -d "$UI_DIR/config" ]; then
    cp -r "$UI_DIR/config" "$BUILD_DIR/app/ui/"
    echo "  UI config copied"
fi
if [ -d "$UI_DIR/images" ]; then
    cp -r "$UI_DIR/images" "$BUILD_DIR/app/ui/"
    echo "  UI images copied"
fi

echo "[4/5] Build and prepare server files..."
SERVER_DIR="$BUILD_DIR/app/server"
mkdir -p "$SERVER_DIR"

SERVER_SRC="$PROJECT_DIR/app/server"
if [ -f "$SERVER_SRC/main.go" ]; then
    cd "$SERVER_SRC"
    echo "  Running go mod tidy..."
    go mod tidy || echo "  Warning: go mod tidy failed"

    echo "  Building Go server..."
    GOOS=linux GOARCH=arm64 go build -ldflags "-X dockpit/internal/handler.appVersion=$APP_VERSION" -o dockpit .
    if [ $? -ne 0 ]; then
        echo "  Error: Go build failed"
        exit 1
    fi

    if [ -f "$SERVER_SRC/dockpit" ]; then
        cp "$SERVER_SRC/dockpit" "$SERVER_DIR/"
        echo "  Server executable built and copied (linux/arm64)"
    else
        echo "  Error: Server executable not found"
        exit 1
    fi
else
    echo "  Error: Server source not found (main.go)"
    exit 1
fi

echo "[5/5] Build package..."
FNPACK_FILE="fnpack-${FNPACK_VER}-linux-amd64"
FNPACk_PATH="$BUILD_DIR/$FNPACK_FILE"

if [ "$FORCE_DOWNLOAD" = false ] && [ -f "$FNPACk_PATH" ] && test_version_match "fnpack" "$FNPACK_VER"; then
    echo "  Using cached fnpack $FNPACK_VER"
else
    if ! get_file_direct "$FNPACK_URL" "$FNPACk_PATH" "fnpack" "fnpack" "$FNPACK_VER"; then
        exit 1
    fi
fi

chmod +x "$FNPACk_PATH"

rm -f "$BUILD_DIR/dockpit.fpk"
cd "$BUILD_DIR"
"$FNPACk_PATH" build

if [ -f "dockpit.fpk" ]; then
    mv dockpit.fpk "$PROJECT_DIR/dockpit-$APP_VERSION.fpk"
    echo "  Build success!"
else
    echo "  Error: Build failed"
    exit 1
fi

echo ""
echo "========================================"
echo "  Build complete!"
echo "  Output: dockpit-$APP_VERSION.fpk"
echo "========================================"
