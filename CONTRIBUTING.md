# Contributing

Thanks for contributing to `go-admin`.

## Development Setup

Use Go `1.26` or newer.

Run SDK checks from the repository root:

```bash
go test ./...
go vet ./...
```

Run the example admin service:

```bash
go run ./examples/memory
```

Run the documentation site:

```bash
cd docs
nvm use
npm install
npm run dev
```

## Project Boundaries

- Keep the admin UI simple, server-rendered, and Django-like.
- Do not add authentication, sessions, roles, or permissions to the SDK core. Host applications own that layer.
- Keep product documentation in the Starlight site under `docs/`.
- Keep the root `README.md` limited to basic install and usage information.

## Pull Request Checklist

- Format Go files with `gofmt`.
- Run `go test ./...`.
- Run `go vet ./...` for broader Go changes.
- Run `cd docs && nvm use && npm run build` when documentation changes.
- Keep changes focused on one behavior, fix, or documentation topic.
