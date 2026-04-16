#!/bin/sh
set -e

# Seed migration files from the image into the traildepot volume.
# Only copies files that don't already exist so applied migrations are never overwritten.
mkdir -p /app/traildepot/migrations
for f in /app/seed-migrations/*.sql; do
  [ -f "$f" ] || continue
  dest="/app/traildepot/migrations/$(basename "$f")"
  if [ ! -f "$dest" ]; then
    cp "$f" "$dest"
    echo "Seeded migration: $(basename "$f")"
  fi
done

# Start Trailbase (replaces this shell process)
exec /app/trail run
