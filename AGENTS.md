## Agent Instructions

This document provides guidelines for AI agents working in this Go/HTMX/Tailwind/PostgreSQL e-commerce repository.

### Tech Stack & Setup

- **Backend:** Go 1.24.2 with PostgreSQL database
- **Frontend:** HTMX + Tailwind CSS
- **Config:** TOML files in `configs/` directory
- **Database:** Set up PostgreSQL and update `configs/config.toml`

### Build, Lint, and Test

- **Build frontend:** `tailwind -i ./web/static/css/style.css -o ./web/static/css/dist/style.css`
- **Build backend:** `go build -o lojagtec cmd/server/main.go`
- **Run backend:** `./lojagtec`
- **Test:** `go test ./...`
- **Run a single test:** `go test -run ^TestName$ path/to/package`
- **Lint:** `go fmt ./...` and `go vet ./...`

### Code Style

- **Imports:** Group into three blocks: standard library, third-party, internal packages
- **Formatting:** Use `gofmt` for all Go code
- **Types:** Use structs for config/data structures, type inference (`:=`) where appropriate
- **Naming:** PascalCase for exported identifiers, camelCase for local variables
- **Error Handling:** Handle errors explicitly, use `log.Fatalf` for startup failures, avoid panics
- **Comments:** Explain complex logic only, avoid restating obvious code
