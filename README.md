# go-svc

Go library providing reusable building blocks for HTTP services.

```
go get github.com/yandzee/go-svc
```

## Packages

| Package | Description |
|---|---|
| `crypto` | Secure random bytes, hashing (SHA1, SHA256), hex-encoded random values |
| `data/jsoner` | JSON encoding/decoding helpers |
| `data/page` | Pagination types and utilities |
| `flow` | Control flow types (Continue/Break) for pipeline processing |
| `identity` | Authentication, credential validation, JWT token pairs, user registry |
| `lifecycle` | Service lifecycle event emission and state management |
| `log` | Structured logging utilities wrapping `slog` |
| `pipeline` | Generic stage-based pipeline with flow control |
| `router` | HTTP routing abstractions with compression support |
| `router/std` | `net/http` stdlib-based router implementation |
| `server` | HTTP/HTTP2 server with graceful shutdown |
| `service` | High-level service orchestration and lifecycle management |
| `utils/fs` | Filesystem utilities (directory scanning) |
| `utils/http` | HTTP helpers |
| `utils/jwt` | JWT utilities |
| `utils/writers` | Custom writer implementations |

## Running tests

```sh
go test ./tests/...
```
