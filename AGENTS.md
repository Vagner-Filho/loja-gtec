## Agent Instructions

This document provides guidelines for AI agents working in this repository.

### Build, Lint, and Test

- **Build frontend:** `tailwind -i ./web/static/css/style.css -o ./web/static/css/dist/style.css`
- **Build backend:** `go build -o lojagtec cmd/server/main.go`
- **Run backend:** `./lojagtec`
- **Test:** `go test ./...`
- **Run a single test:** `go test -run ^TestName$ path/to/package`
- **Lint:** `go fmt ./...` and `go vet ./...`

### Code Style

- **Imports:** Group imports into three blocks: standard library, third-party, and internal packages.
- **Formatting:** Use `gofmt` for all Go code.
- **Types:** Use structs for configuration and data structures. Use type inference (`:=`) where appropriate.
- **Naming:** Follow standard Go naming conventions (`PascalCase` for exported, `camelCase` for local).
- **Error Handling:** Handle errors explicitly. Use `log.Fatalf` for critical errors on startup. Don't use panics.
- **Comments:** Add comments to explain complex logic, not to restate what the code does.
