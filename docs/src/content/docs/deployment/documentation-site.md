---
title: Documentation Site
description: Build and host the Starlight documentation site separately from the Go SDK.
---

The documentation website is a standalone Astro Starlight app in `docs/`.

## Local Development

Use the local `.nvmrc` so the docs run on Node 22:

```bash
cd docs
nvm use
npm install
npm run dev
```

The dev server prints the local URL.

## Production Build

```bash
cd docs
nvm use
npm ci
npm run build
```

The static output is written to:

```text
docs/dist/
```

`docs/dist/` is ignored by git. Host that directory with any static hosting provider.

## Preview A Build

```bash
cd docs
nvm use
npm run build
npm run preview
```

Use preview to inspect the production output before publishing.

## GitHub Pages

The docs are configured for:

```text
https://newton-school.github.io/go-admin
```

A GitHub Pages workflow can build and publish `docs/dist/`:

```yaml
name: Deploy docs

on:
  push:
    branches: [master]

permissions:
  contents: read
  pages: write
  id-token: write

concurrency:
  group: pages
  cancel-in-progress: true

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
      - uses: actions/setup-node@v6
        with:
          node-version-file: docs/.nvmrc
          cache: npm
          cache-dependency-path: docs/package-lock.json
      - run: npm ci
        working-directory: docs
      - run: npm run build
        working-directory: docs
      - uses: actions/upload-pages-artifact@v4
        with:
          path: docs/dist

  deploy:
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    runs-on: ubuntu-latest
    needs: build
    steps:
      - id: deployment
        uses: actions/deploy-pages@v5
```

If you host the site at a custom domain, update the `site` value in `docs/astro.config.mjs`.

## Static Hosts

For Netlify, Vercel, Cloudflare Pages, S3, or any static host:

| Setting | Value |
| --- | --- |
| Base directory | `docs` |
| Install command | `npm ci` |
| Build command | `npm run build` |
| Output directory | `docs/dist` or `dist` when the base directory is already `docs` |
| Node version | From `docs/.nvmrc` |

