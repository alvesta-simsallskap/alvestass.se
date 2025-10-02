# Alvesta simsällskap homepage

## Dependencies

- Bulma. Source git repo in `styles/bulma-main` for sass compilation
- Sass (installed using homebrew to avoid npm)
- Alpine js for minor javascript

## Compile styles (Sass)

https://helloyes.dev/blog/2021/minimal-sass-and-eleventy-setup/

```
sass --load-path=styles/bulma-main styles/site.scss site.css
```

## Shared calendars

Calendars can be shared through the website.

Uses Cloudflare page rules:
https://developers.cloudflare.com/rules/page-rules/how-to/url-forwarding/

### Tävling - A och B-gruppen
https://www.alvestass.se/kalender/tavling-ab
https://outlook.office365.com/owa/calendar/52e09640f0ba45dc8466407dc44a91b9@alvestass.se/9e3292d4e0bb43e18d93e51f0c5fd57d13799513555319312600/calendar.ics