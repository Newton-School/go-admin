# Agent Guide

This file is for coding agents working on this repository. It is not product documentation.

## Project Summary

`go-admin` is a Go SDK for building Django-style admin panels. It exposes a standard `net/http` handler, registers apps and model-like resources, renders simple server-side HTML, and exposes matching JSON admin APIs. Authentication, sessions, roles, and permissions are intentionally owned by the host application.

## Repository Layout

- `admin.go`: public package facade for SDK users.
- `internal/core/`: implementation for sites, apps, resources, fields, handlers, templates, queries, actions, and memory repositories.
- `tests/`: package-level behavior tests for the public SDK.
- `examples/memory/`: full local example with multiple apps and useful models.
- `examples/nethttp/`: minimal `net/http` auth wrapper example.
- `docs/`: the only product documentation site, built with Astro Starlight.
- `.codex/environments/local.toml`: Codex run actions for the example server and docs server.
- `.agent/rules/`: repository rules that agents must follow.

## Required Rules

Read and follow these rules before making related changes:

- `.agent/rules/documentation-location.mdc`: all product documentation must live in the top-level `docs/` website. Do not create new docs locations anywhere else.
- `.agent/rules/documentation-maintenance.mdc`: maintain docs as a buildable Starlight site and run the docs build after docs changes.
- `.agent/rules/root-readme.mdc`: keep the root `README.md` minimal. It should contain only basic install/use information and links to the full docs.
- `.agent/rules/release-management.mdc`: only when explicitly asked to create a release, update install tags, generate release notes, create the Git tag, and publish the GitHub Release.

When rules conflict with a direct user instruction, follow the user instruction and keep the change tightly scoped.

## Development Workflow

- Prefer existing public APIs and `internal/core` patterns over new abstractions.
- Keep UI simple and Django-like for the admin panel; do not introduce fancy admin UI styling.
- Keep changes focused. Avoid unrelated refactors and generated-file churn.
- Use small commits for small feature or documentation slices. Do not push unless explicitly asked.
- Do not add authentication or permissions to the SDK core unless explicitly requested.
- Keep product docs out of root files, examples, packages, and ad hoc markdown files.

## Common Commands

Run Go checks from the repository root:

```bash
go test ./...
go vet ./...
```

Run the memory example:

```bash
go run ./examples/memory
```

Run or build docs with the local Node version:

```bash
cd docs
nvm use
npm install
npm run dev
npm run build
```

## Verification Expectations

- For Go SDK or handler changes, run `go test ./...`.
- For broader Go changes, also run `go vet ./...`.
- For documentation-site changes, run `cd docs && nvm use && npm run build`.
- For frontend-visible admin UI changes, run the relevant server and inspect the result in a browser when practical.
