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

The checked-in workflow at `.github/workflows/deploy-docs.yml` builds and publishes `docs/dist/` on every push to `master`. It can also be run manually from the Actions tab.

GitHub Pages must use GitHub Actions as the Pages source in the repository settings.

The workflow uses this deployment shape:

```yaml
name: Deploy docs

on:
  push:
    branches: [master]
  workflow_dispatch:

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
      - name: Checkout
        uses: actions/checkout@v6

      - name: Setup Pages
        uses: actions/configure-pages@v5

      - name: Setup Node
        uses: actions/setup-node@v6
        with:
          node-version-file: docs/.nvmrc
          cache: npm
          cache-dependency-path: docs/package-lock.json

      - name: Install dependencies
        run: npm ci
        working-directory: docs

      - name: Build docs
        run: npm run build
        working-directory: docs

      - name: Upload Pages artifact
        uses: actions/upload-pages-artifact@v4
        with:
          path: docs/dist

  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    steps:
      - name: Deploy to GitHub Pages
        id: deployment
        uses: actions/deploy-pages@v4
```

For the GitHub Pages project URL, `docs/astro.config.mjs` sets:

```js
site: 'https://newton-school.github.io',
base: '/go-admin',
```

If you host the site at a custom domain, update `site` and remove or change `base` for that final URL.

## Static Hosts

For Netlify, Vercel, Cloudflare Pages, S3, or any static host:

| Setting | Value |
| --- | --- |
| Base directory | `docs` |
| Install command | `npm ci` |
| Build command | `npm run build` |
| Output directory | `docs/dist` or `dist` when the base directory is already `docs` |
| Node version | From `docs/.nvmrc` |
