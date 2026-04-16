FROM trailbase/trailbase:latest

ENV ADDRESS=0.0.0.0:4000

# Bake migrations into the image at a non-volume path.
# docker-entrypoint.sh copies them to /app/traildepot/migrations/ on startup.
COPY migrations/ /app/seed-migrations/
COPY --chmod=755 docker-entrypoint.sh /app/docker-entrypoint.sh

ENTRYPOINT ["/app/docker-entrypoint.sh"]
