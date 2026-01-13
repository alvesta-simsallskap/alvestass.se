// @ts-check
import { defineConfig } from 'astro/config';

import cloudflare from '@astrojs/cloudflare';

import alpinejs from '@astrojs/alpinejs';

// https://astro.build/config
export default defineConfig({
  adapter: cloudflare({
    imageService: 'cloudflare', // Use Cloudflare image service
    platformProxy: {
      enabled: true,
    },
  }),

  integrations: [alpinejs()]
});