import { defineConfig } from 'astro/config';

import cloudflare from '@astrojs/cloudflare';

import alpinejs from '@astrojs/alpinejs';

// https://astro.build/config
export default defineConfig({
  adapter: cloudflare({
    imageService: 'cloudflare', // Use Cloudflare image service
  }),

  // @astrojs/cloudflare v13 enables Cloudflare KV sessions by default.
  // This site has no session needs, so use in-memory to suppress that behaviour.
  session: {
    driver: { entrypoint: 'unstorage/drivers/memory' },
  },

  integrations: [alpinejs()]
});