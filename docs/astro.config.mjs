import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
  site: 'https://newton-school.github.io',
  base: '/go-admin',
  integrations: [
    starlight({
      title: 'go-admin',
      description: 'Django-style admin panels for Go services.',
      customCss: ['./src/styles/custom.css'],
      favicon: '/favicon.svg',
      logo: {
        src: './src/assets/logo.svg',
        alt: 'go-admin',
      },
      social: [
        {
          icon: 'github',
          label: 'GitHub',
          href: 'https://github.com/Newton-School/go-admin',
        },
      ],
      sidebar: [
        {
          label: 'Start',
          items: [
            { label: 'Overview', slug: 'index' },
            { label: 'Installation', slug: 'getting-started/installation' },
            { label: 'Quick Start', slug: 'getting-started/quick-start' },
            { label: 'Secure The Admin', slug: 'getting-started/secure-the-admin' },
            { label: 'Example App', slug: 'getting-started/example-app' },
          ],
        },
        {
          label: 'Guides',
          items: [
            { label: 'Core Concepts', slug: 'guide/core-concepts' },
            { label: 'Apps And Resources', slug: 'guide/apps-and-resources' },
            { label: 'Repositories', slug: 'guide/repositories' },
            { label: 'Fields And Forms', slug: 'guide/fields-and-forms' },
            { label: 'List Pages', slug: 'guide/list-pages' },
            { label: 'Actions', slug: 'guide/actions' },
            { label: 'JSON API', slug: 'guide/json-api' },
            { label: 'Testing', slug: 'guide/testing' },
          ],
        },
        {
          label: 'Reference',
          items: [
            { label: 'Public Go API', slug: 'reference/public-go-api' },
            { label: 'Configuration', slug: 'reference/configuration' },
            { label: 'API Routes', slug: 'reference/api-routes' },
          ],
        },
        {
          label: 'Deployment',
          items: [
            { label: 'Admin Service', slug: 'deployment/admin-service' },
            { label: 'Documentation Site', slug: 'deployment/documentation-site' },
          ],
        },
      ],
    }),
  ],
});
