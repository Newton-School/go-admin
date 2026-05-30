---
title: Installation
description: Install go-admin and prepare a Go service for an admin panel.
---

`go-admin` is a Go module. It mounts as a standard `net/http` handler, so it works with the standard library and with routers that can serve an `http.Handler`.

## Requirements

- Go 1.26 or newer.
- A Go module for your application.
- A repository implementation for each model-like resource you register.
- Your own auth middleware around the mounted admin handler.

## Install The SDK

Run this from your application module:

```bash
go get github.com/Newton-School/go-admin@v0.1.1
```

Use an explicit tag so your application keeps the same SDK version until you choose to upgrade.

Import the package:

```go
import admin "github.com/Newton-School/go-admin"
```

## Create A Site

Create a site with a title and mount path:

```go
site := admin.New(admin.SiteConfig{
    Title:    "Acme Admin",
    BasePath: "/admin",
})
```

`BasePath` is normalized. An empty value and `/` both become `/admin`.

## Mount The Handler

Register both the slash and non-slash paths when using `net/http` directly:

```go
http.Handle("/admin/", requireAdmin(site.Handler()))
http.Handle("/admin", requireAdmin(site.Handler()))
```

`site.Handler()` serves the HTML admin and JSON API. Put auth, logging, rate limits, and request context middleware outside this handler.

## Build The Documentation Site

The documentation site lives in `docs/` and is built separately from the Go module.

```bash
cd docs
nvm use
npm install
npm run build
```

The built static site is written to `docs/dist/`.
