---
title: go-admin
description: Django-style admin panels for Go services.
template: splash
hero:
  title: go-admin
  tagline: Build Django-style admin panels in Go with a small SDK, server-rendered HTML, and your own auth middleware.
  image:
    file: ../../assets/logo.svg
  actions:
    - text: Start building
      link: getting-started/installation/
      icon: right-arrow
    - text: GitHub
      link: https://github.com/Newton-School/go-admin
      variant: secondary
---

## What It Gives You

- Apps and resources, like Django admin.
- List, create, edit, detail, delete, filter, search, sort, pagination, and bulk action screens.
- A matching JSON API under `/admin/api/v1`.
- An ORM-agnostic repository interface.
- A standard `net/http` handler that you secure with your own middleware.

## What You Own

`go-admin` does not include login, sessions, users, roles, or permissions. Mount it behind your existing auth and authorization layer.

## Install

```bash
go get github.com/Newton-School/go-admin@v0.1.0
```

## Documentation Map

| Section | Use it for |
| --- | --- |
| [Installation](getting-started/installation/) | Add the module, create a site, and mount the handler. |
| [Quick Start](getting-started/quick-start/) | Register your first app and resource. |
| [Repositories](guide/repositories/) | Connect resources to a database, ORM, or service layer. |
| [Fields And Forms](guide/fields-and-forms/) | Configure widgets, validation, choices, and fieldsets. |
| [List Pages](guide/list-pages/) | Configure columns, search, filters, sorting, and pagination. |
| [Actions](guide/actions/) | Add bulk operations to list pages and API clients. |
| [JSON API](guide/json-api/) | Use the built-in admin API. |
| [Public Go API](reference/public-go-api/) | Look up exported functions and types. |
| [Documentation Site](deployment/documentation-site/) | Build and host this Starlight site separately. |
