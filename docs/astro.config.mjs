import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
  site: 'https://newton-school.github.io/go-admin',
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
      ],
    }),
  ],
});
