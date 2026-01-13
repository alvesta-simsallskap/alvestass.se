# Alvesta simsällskap homepage

## Development

```
pnpm dev
```

Wrangler and KV (key, value) in Cloudflare

```
npx wrangler kv namespace create "TIME"
npx wrangler kv key put <KEY> <VALUE> --binding=<TIME> --local
```

## Website

Pages CMS.

## Dependencies

- Bulma (pnpm installed)
- Alpine js for minor javascript

## References

https://css-tricks.com/using-pages-cms-for-static-site-content-management/

## Calendars to subscribe to

Added as static .ics files in "public" folder, with _headers file controlling Content-Type header.
