#!/bin/sh
set -e

# Create the SQLite database file if it doesn't exist
if [ ! -f "${DB_DATABASE:-/var/www/html/database/database.sqlite}" ]; then
    touch "${DB_DATABASE:-/var/www/html/database/database.sqlite}"
    chown www-data:www-data "${DB_DATABASE:-/var/www/html/database/database.sqlite}"
fi

# Run migrations on every start (safe: only applies pending ones)
php artisan migrate --force

# Clear and warm caches
php artisan config:cache
php artisan route:cache
php artisan view:cache

exec apache2-foreground