# Config HTTP Package

The `config/http` package provides HTTP handlers for configuration management, enabling runtime configuration inspection and modification through a RESTful HTTP interface.

## Features

- **RESTful API**: Resource-centric design with hierarchical URL paths mapping to configuration keys
- **Read/Write Operations**: GET, PUT, DELETE support based on config capabilities
- **Embeddable**: Works with any `http.ServeMux`, chi, gorilla/mux, or other routers
- **Security Options**: Read-only mode, key filtering, middleware support
- **Meta Endpoints**: List keys, get handler info
- **Admin Endpoints**: Configuration reload support
- **Interface Detection**: Automatically detects and uses `MutableConfig`, `MarshableConfig`, and other optional interfaces
- **High Test Coverage**: >97% test coverage

## Installation

```bash
go get github.com/grinps/go-utils/config/http
```

## Quick Start

```go
package main

import (
    "context"
    "net/http"

    "github.com/grinps/go-utils/config"
    confighttp "github.com/grinps/go-utils/config/http"
)

func main() {
    ctx := context.Background()
    data := map[string]any{
        "server": map[string]any{
            "port": 8080,
            "host": "localhost",
        },
    }
    
    cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
    handler := confighttp.NewHandler(cfg)
    
    http.ListenAndServe(":8080", handler)
}
```

## API Endpoints

### Configuration Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/` | Get entire configuration |
| `GET` | `/{key...}` | Get value at key path |
| `HEAD` | `/{key...}` | Check if key exists |
| `PUT` | `/{key...}` | Set value at key path |
| `DELETE` | `/{key...}` | Delete key |

### Meta Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/_meta/info` | Get handler information |
| `GET` | `/_meta/keys` | List all configuration keys |
| `GET` | `/_meta/keys/{prefix}` | List keys with prefix |

### Admin Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/_admin/reload` | Reload configuration |

## URL to Key Mapping

URL paths are converted to dot-notation configuration keys:

```
/server/port       → server.port
/database/host     → database.host
/app/settings/debug → app.settings.debug
```

## Response Format

### Success Response (GET)

```json
{
  "key": "server.port",
  "value": 8080
}
```

### Keys Response

```json
{
  "keys": ["server.port", "server.host", "database.host"],
  "count": 3,
  "prefix": ""
}
```

### Info Response

```json
{
  "provider": "KoanfConfig",
  "mutable": true,
  "marshable": true,
  "has_keys": true,
  "has_delete": true,
  "has_reload": true,
  "read_only": false,
  "base_path": "/api/config",
  "key_filtered": false
}
```

### Error Response

```json
{
  "error": {
    "code": "CONFIG_KEY_NOT_FOUND",
    "message": "key not found",
    "details": {"key": "server.missing"}
  }
}
```

## Configuration Options

### WithBasePath

Mount the handler at a specific path prefix:

```go
handler := confighttp.NewHandler(cfg, confighttp.WithBasePath("/api/config"))
// Requests to /api/config/server/port will look up "server.port"
```

### WithReadOnly

Disable write operations:

```go
handler := confighttp.NewHandler(cfg, confighttp.WithReadOnly(true))
// PUT and DELETE return 403 Forbidden
```

### WithKeyFilter

Filter accessible keys:

```go
handler := confighttp.NewHandler(cfg, confighttp.WithKeyFilter(func(key string) bool {
    return !strings.HasPrefix(key, "secrets.")
}))
// Keys starting with "secrets." are hidden and inaccessible
```

### WithReloadHandler

Enable configuration reload:

```go
handler := confighttp.NewHandler(cfg, confighttp.WithReloadHandler(func(ctx context.Context) error {
    return kcfg.Load(ctx, file.Provider("config.json"), json.Parser())
}))
// POST /_admin/reload will trigger the reload function
```

### WithMiddleware

Add middleware for authentication, logging, etc:

```go
handler := confighttp.NewHandler(cfg, confighttp.WithMiddleware(
    loggingMiddleware,
    authMiddleware,
))
```

### WithDelimiter

Use a custom key delimiter:

```go
handler := confighttp.NewHandler(cfg, confighttp.WithDelimiter("/"))
// /server/port becomes key "server/port" instead of "server.port"
```

### WithMetaEndpoints / WithAdminEndpoints

Disable meta or admin endpoints:

```go
handler := confighttp.NewHandler(cfg,
    confighttp.WithMetaEndpoints(false),   // Disable /_meta/*
    confighttp.WithAdminEndpoints(false),  // Disable /_admin/*
)
```

## Mounting on Existing Servers

### Standard http.ServeMux

```go
mux := http.NewServeMux()
mux.HandleFunc("/api/users", usersHandler)

configHandler := confighttp.NewHandler(cfg, confighttp.WithBasePath("/api/config"))
mux.Handle("/api/config/", configHandler)

http.ListenAndServe(":8080", mux)
```

### Chi Router

```go
import "github.com/go-chi/chi/v5"

r := chi.NewRouter()
r.Mount("/config", confighttp.NewHandler(cfg))
```

### Gorilla Mux

```go
import "github.com/gorilla/mux"

r := mux.NewRouter()
r.PathPrefix("/config").Handler(confighttp.NewHandler(cfg, confighttp.WithBasePath("/config")))
```

## Interface Support

The handler automatically detects and uses optional interfaces from the `config` package:

| Interface | Methods | Enables |
|-----------|---------|---------|
| `config.Config` | `GetValue`, `GetConfig`, `Name` | GET operations (required) |
| `config.MutableConfig` | `SetValue` | PUT operations |
| `config.AllKeysProvider` | `Keys(prefix)` | `/_meta/keys` endpoint |
| `config.Deleter` | `Delete(key)` | DELETE operations |
| `config.AllGetter` | `All(ctx)` | Full config retrieval |

## Usage Examples

### Read-Only Production Handler

```go
handler := confighttp.NewHandler(cfg,
    confighttp.WithReadOnly(true),
    confighttp.WithKeyFilter(func(key string) bool {
        // Hide sensitive keys
        sensitivePatterns := []string{"password", "secret", "key", "token"}
        for _, pattern := range sensitivePatterns {
            if strings.Contains(strings.ToLower(key), pattern) {
                return false
            }
        }
        return true
    }),
    confighttp.WithMetaEndpoints(true),
    confighttp.WithAdminEndpoints(false),
)
```

### Development Handler with Reload

```go
handler := confighttp.NewHandler(cfg,
    confighttp.WithReloadHandler(func(ctx context.Context) error {
        log.Println("Reloading configuration...")
        return kcfg.Load(ctx, file.Provider("config.yaml"), yaml.Parser())
    }),
)
```

### Authenticated Handler

```go
authMiddleware := func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if !isValidToken(token) {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}

handler := confighttp.NewHandler(cfg, confighttp.WithMiddleware(authMiddleware))
```

## cURL Examples

```bash
# Get all configuration
curl http://localhost:8080/

# Get specific value
curl http://localhost:8080/server/port

# Check if key exists
curl -I http://localhost:8080/server/port

# Set a value (requires MutableConfig)
curl -X PUT http://localhost:8080/server/port \
  -H "Content-Type: application/json" \
  -d '{"value": 9090}'

# Delete a key (requires Deleter)
curl -X DELETE http://localhost:8080/server/debug

# List all keys
curl http://localhost:8080/_meta/keys

# List keys with prefix
curl http://localhost:8080/_meta/keys/server

# Get handler info
curl http://localhost:8080/_meta/info

# Reload configuration
curl -X POST http://localhost:8080/_admin/reload
```

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `CONFIG_KEY_NOT_FOUND` | 404 | Key does not exist |
| `CONFIG_INVALID_KEY` | 400 | Invalid key format |
| `CONFIG_READ_ONLY` | 403 | Write operation in read-only mode |
| `CONFIG_NOT_MUTABLE` | 501 | Config doesn't support mutations |
| `CONFIG_KEY_FILTERED` | 403 | Key access denied by filter |
| `CONFIG_INVALID_BODY` | 400 | Invalid request body |
| `CONFIG_DELETE_NOT_SUPPORTED` | 501 | Config doesn't support delete |
| `CONFIG_RELOAD_NOT_CONFIGURED` | 501 | Reload handler not configured |
| `CONFIG_RELOAD_FAILED` | 500 | Reload operation failed |
| `METHOD_NOT_ALLOWED` | 405 | HTTP method not supported |
| `INTERNAL_ERROR` | 500 | Internal server error |

## Thread Safety

The Handler is safe for concurrent use. Thread safety of configuration modifications depends on the underlying Config implementation.

## Testing

```bash
cd config/http
go test -v -cover ./...
```

## License

Part of the grinps/go-utils project.

## See Also

- [config package](../README.md) - Parent config package
- [config/koanf package](../koanf/README.md) - Koanf-based config with Keys and Delete support
- [config/ext package](../ext/README.md) - Config extensions
