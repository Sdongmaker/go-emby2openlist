# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

go-emby2openlist is a reverse proxy for Emby that enables direct cloud storage streaming by intercepting Emby API requests and redirecting media streams to cloud storage direct links (primarily Openlist/Aliyun). This eliminates server bandwidth consumption by creating client-to-cloud direct connections.

**Current Version**: v2.3.2
**Language**: Go 1.24.2
**Framework**: Gin

## Build & Run Commands

### Local Development
```bash
# Run with default config
go run main.go

# Run with custom data root
go run main.go -dr /path/to/config

# Run with custom ports
go run main.go -p 8080 -ps 8443

# Check version
go run main.go -version
```

### Building
```bash
# Build all platforms (creates binaries in dist/)
bash build.sh

# Manual build for current platform
CGO_ENABLED=0 go build -ldflags="-X main.ginMode=release" -o ge2o
```

### Testing
```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/service/m3u8/
go test ./internal/util/jsons/

# Run with verbose output
go test -v ./internal/service/openlist/localtree/

# Run single test
go test -run TestFunctionName ./internal/service/path/
```

### Docker
```bash
# Build and run with compose
docker-compose up -d --build

# View logs
docker logs -f go-emby2openlist -n 1000

# Restart after config changes
docker-compose restart

# Update to new version
docker-compose down
git pull
docker-compose up -d --build
docker image prune -f
```

## Architecture Essentials

### Request Flow

The system uses **regex-based routing with ordered pattern matching**:

```
Client Request
  ↓
globalDftHandler() (web/handler.go)
  ↓
[Iterate 26 regex rules in order - web/route.go]
  ↓
Middleware Chain: ApiKeyChecker → CacheableRouteMarker → RequestCacher
  ↓
Matched Handler (emby/*, m3u8/*, etc.)
  ↓
Response (often 302 redirect to cloud storage)
```

**Critical**: Route order matters. Rules are checked sequentially until first match. Modify `web/route.go` rules carefully.

### Core Service Modules

#### Emby Service (`internal/service/emby/`)
- **Purpose**: Intercepts Emby API calls, modifies responses to enable direct streaming
- **Key Files**:
  - `playbackinfo.go` - Intercepts PlaybackInfo API, adds transcoded versions, prevents unwanted client transcoding
  - `redirect.go` - Core redirect logic: `Redirect2OpenlistLink()` maps Emby paths → Openlist → cloud URLs
  - `media.go` - Extracts media source info, handles MediaSourceId encoding (uses `[[_]]` segment separator)
  - `auth.go` - Validates API keys against Emby server, caches valid keys in sync.Map

#### Openlist Service (`internal/service/openlist/`)
- **Purpose**: Fetches cloud storage direct links via Openlist API
- **Key Functions**:
  - `FetchResource()` - Main entry point, handles transcode fallback to raw
  - `FetchFsGet()` - Gets original quality link
  - `FetchFsOther()` - Gets transcoded link with format selection
- **Important**: All API calls go through `Fetch()` which adds Authorization token from config

#### M3U8 Service (`internal/service/m3u8/`)
- **Purpose**: Maintains HLS transcoding playlists in memory
- **Architecture**: Single goroutine (`loopMaintainPlaylist`) manages max 10 playlists with LRU eviction
- **Pre-caching**: Buffered channel (1000 size) queues async playlist generation requests
- **Key Exports**:
  - `GetPlaylist()` - Returns m3u8 text for video
  - `GetTsLink()` - Returns direct link for TS segment
  - `GetSubtitleLink()` - Returns subtitle URL

#### Path Service (`internal/service/path/`)
- **Purpose**: Maps Emby file paths → Openlist paths
- **Transformation**:
  1. URL decode
  2. Convert Windows backslashes → forward slashes
  3. Remove Emby mount-path prefix
  4. Apply custom path mappings from config
  5. If uncertain, try all Openlist root directories
- **Key Function**: `Emby2Openlist(embyPath)` returns `PathConvertResult` with `Success`, `Path`, and `Range()` method

#### Local Tree Generation (`internal/service/openlist/localtree/`)
- **Purpose**: Syncs Openlist directory structure to local filesystem for Emby scanning
- **Three Modes**:
  1. **STRM files** - Lightweight text files with URLs
  2. **Virtual files** - Empty files with embedded duration metadata
  3. **Music virtual files** - Full ID3 tag extraction via FFmpeg
- **Components**:
  - `Synchronizer` - Syncs remote changes to local
  - `Snapshot` - Tracks file state between scans
  - Single goroutine with configurable refresh interval

### Caching Architecture (`internal/web/cache/`)

**Cache Key**: `MD5(method + uriNoArgs + sortedParams + body + selectedHeaders)`

**Ignored Params**: 70+ variants including PlaySessionId, X-Forwarded-For, connection headers

**Memory Management**:
- Max 100MB total (`MaxCacheSize`)
- Max 8092 entries (`MaxCacheNum`)
- Pre-cache channel (1000) with FIFO eviction
- Single maintenance goroutine, cleanup every 10s

**Cache Control via Response Headers**:
- `cache.HeaderKeyExpired` header sets expiration (milliseconds since epoch)
- Value `-1` = skip caching

**Durations**:
- PlaybackInfo: 12 hours
- Video subtitles: 30 days (fixed)
- Direct links: 10 minutes
- Random items: configurable

### Configuration System (`internal/config/`)

**Loading**: YAML → struct unmarshaling → reflection-based auto-initialization → cascade Init()

**All config structs implement**:
```go
type Initializer interface {
    Init() error  // Called automatically after YAML load
}
```

**Key Config Sections**:
- `emby` - Source Emby server, mount paths, proxy strategies
- `openlist` - Openlist API connection, local tree generation settings
- `video-preview` - Transcoding container filters, ignored template IDs
- `path.emby2openlist` - Path prefix mappings
- `cache` - Enable/disable, expiration duration
- `ssl` - Certificate files, single-port mode

**Path Mapping Format**: `- /emby/prefix:/openlist/prefix`

## Important Design Patterns

### Single Goroutine Maintenance
Used for M3U8 playlists, cache, and local tree. Pre-cache channel buffers requests, single goroutine processes serially. Avoids race conditions.

### Sync.Map for Lock-Free Caching
API key validation cache and response cache use `sync.Map` for high concurrency without explicit locks.

### Regex Rule Chain
Routes defined as ordered slice of `[regex, handler]` pairs. First match wins. Add new routes in appropriate priority order.

### MediaSourceId Encoding
Custom IDs use `[[_]]` as segment separator to encode transcoding metadata:
```
{originId}[[_]]{templateId}[[_]]{format}[[_]]{openlistPath}
```

### Context Markers
Middleware chain stores matched route and submatch groups in Gin context:
```go
c.Set(constant.MatchRouteKey, regex.String())
c.Set(constant.RouteSubMatchGinKey, submatchGroups)
```

## Adding New Features

### Adding a New Route Handler
1. Add regex pattern to `internal/constant/constant.go`
2. Add handler function to appropriate service (emby/openlist/m3u8)
3. Add `[pattern, handler]` pair to `rules` slice in `web/route.go` at correct priority
4. If route needs caching, it must return proper cache headers

### Adding New Configuration
1. Add struct to appropriate file in `internal/config/`
2. Add field to `Config` struct in `config.go`
3. Implement `Init() error` method for validation
4. Add YAML example to `config-example.yml`
5. Configuration automatically loaded via reflection

### Modifying Cache Behavior
- Adjust durations in handler code using `c.Header(cache.HeaderKeyExpired, cache.Duration(time.Duration))`
- Modify ignored parameters in `cache/cache.go` `shouldIgnoreParam()`
- Change memory limits in `cache/cache.go` constants

### Extending Path Mapping
Path conversions happen in `internal/service/path/path.go`. The `Emby2Openlist()` function:
1. Loads mappings from `config.C.Path.Emby2Openlist`
2. Tries each mapping prefix in order
3. Returns first successful match or all possible paths via `Range()`

Modify config YAML to add new mappings, or extend logic in `path.go` for complex transformations.

## Common Pitfalls

### Route Handler Return
Handlers MUST call one of: `c.Redirect()`, `c.String()`, `c.JSON()`, or `ProxyOrigin(c)`. Not returning a response causes client timeout.

### MediaSourceId Parsing
When adding transcode versions, ensure MediaSourceId format matches `resolveMediaSourceId()` expectations in `media.go`. Use `MediaSourceIdSegment` constant.

### Path Conversion Edge Cases
Emby may send URL-encoded paths or Windows-style backslashes. Always use `path.Emby2Openlist()` rather than manual string replacement.

### Cache Key Collisions
Adding new query parameters? Verify they're not in the "ignored" list in `cache/cache.go`, or cache keys may collide unexpectedly.

### Goroutine Safety
M3U8 playlist map and cache map use single-goroutine maintenance. Don't add direct map access elsewhere without proper locking.

## Debugging

### Enable Verbose Logging
Set `GIN_MODE=debug` environment variable before running.

### Profiling
pprof server runs on port 60360 by default:
```bash
go tool pprof http://localhost:60360/debug/pprof/profile
```

### Trace Route Matching
Check logs for "正在匹配路由" (matching route) messages to see which regex matched.

### Inspect Cache
Cache statistics logged every 10 seconds when `GIN_MODE=debug`.

### Test Path Conversion
Use `path_test.go` to verify Emby → Openlist path transformations.

## Project Structure Highlights

```
internal/
  config/        - YAML configuration with reflection-based init
  constant/      - Regex patterns, route keys
  model/         - HTTP response wrappers
  service/
    emby/        - Emby API interception and modification
    openlist/    - Cloud storage API calls
      localtree/ - Local filesystem generation
    m3u8/        - HLS playlist management
    music/       - MP3 metadata writing
    path/        - Path transformation
  util/          - Reusable utilities (JSON, HTTP, crypto, etc.)
  web/
    cache/       - Response caching system
    handler.go   - Route matching dispatcher
    route.go     - Route rule definitions
    web.go       - Gin server initialization
main.go          - Entry point
build.sh         - Multi-platform build script
```

## External Dependencies

Key third-party packages:
- `gin-gonic/gin` - HTTP framework
- `gorilla/websocket` - WebSocket proxy
- `gopkg.in/yaml.v3` - Config parsing
- `github.com/abema/go-mp4` - MP4 metadata writing
- Standard library extensively used (crypto/md5, net/http/httputil, reflect, sync)

## References

- Main README: Detailed user documentation, deployment guides
- Issue #108: Core configuration guide
- Docker Hub: `ambitiousjun/go-emby2openlist:v2.3.2`
- Repository: https://github.com/AmbitiousJun/go-emby2openlist
