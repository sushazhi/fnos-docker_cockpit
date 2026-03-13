// Container types
export interface Container {
  Id: string
  Names: string[]
  Image: string
  ImageID: string
  Command: string
  Created: number
  State: 'running' | 'paused' | 'exited' | 'created' | 'restarting' | 'removing' | 'dead'
  Status: string
  Ports: Port[]
  Labels: Record<string, string>
  Mounts: Mount[]
  NetworkSettings?: NetworkSettings
}

export interface Port {
  IP?: string
  PrivatePort: number
  PublicPort?: number
  Type: string
}

export interface Mount {
  Type: string
  Name?: string
  Source: string
  Destination: string
  Driver?: string
  Mode: string
  RW: boolean
  Propagation: string
}

export interface NetworkSettings {
  Networks: Record<string, Network>
}

export interface Network {
  IPAMConfig?: unknown
  Links?: string[]
  Aliases?: string[]
  NetworkID: string
  EndpointID: string
  Gateway: string
  IPAddress: string
  IPPrefixLen: number
  IPv6Gateway: string
  GlobalIPv6Address: string
  GlobalIPv6PrefixLen: number
  MacAddress: string
}

// Image types
export interface Image {
  Id: string
  ParentId: string
  RepoTags: string[]
  RepoDigests: string[]
  Created: number
  Size: number
  VirtualSize: number
  SharedSize: number
  Labels: Record<string, string>
  Containers: number
}

// Volume types
export interface Volume {
  Name: string
  Driver: string
  Mountpoint: string
  CreatedAt: string
  Labels: Record<string, string>
  Scope: string
  Options: Record<string, string>
}

// Batch operation types
export interface BatchOperationRequest {
  ids: string[]
  operation: 'start' | 'stop' | 'restart' | 'pause' | 'unpause' | 'remove'
  force?: boolean
  timeout?: number
}

export interface BatchOperationResult {
  success: string[]
  failed: Array<{
    id: string
    error: string
  }>
}
