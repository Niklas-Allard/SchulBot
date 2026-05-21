# ── Stage 1: Build Vue/Inertia assets ─────────────────────────────────────────
FROM node:22-alpine AS node-builder

WORKDIR /app

RUN npm install -g npm

COPY package.json npm-workspace.yaml ./
RUN npm install --frozen-lockfile

COPY . .
RUN npm run build

# ── Stage 2: PHP dependencies ─────────────────────────────────────────────────
FROM composer:2 AS composer-builder

WORKDIR /app

COPY composer.json composer.lock ./
RUN composer install --no-dev --no-scripts --prefer-dist --optimize-autoloader

COPY . .
RUN composer run-script post-autoload-dump 2>/dev/null || true

# ── Stage 3: Runtime (PHP 8.4 + Apache) ───────────────────────────────────────
FROM php:8.4-apache

# PHP extensions for Laravel
RUN apt-get update && apt-get install -y --no-install-recommends \
        libsqlite3-dev \
        libzip-dev \
        unzip \
    && docker-php-ext-install pdo pdo_sqlite zip opcache \
    && a2enmod rewrite headers \
    && rm -rf /var/lib/apt/lists/*

# Apache config: serve from public/
RUN sed -i 's|/var/www/html|/var/www/html/public|g' /etc/apache2/sites-available/000-default.conf \
    && echo 'ServerName localhost' >> /etc/apache2/apache2.conf

WORKDIR /var/www/html

# Copy PHP vendor from composer stage
COPY --from=composer-builder /app/vendor ./vendor

# Copy built frontend assets
COPY --from=node-builder /app/public/build ./public/build

# Copy app source
COPY . .

# Runtime directories
RUN mkdir -p storage/framework/{sessions,views,cache} bootstrap/cache database \
    && chown -R www-data:www-data storage bootstrap/cache database \
    && chmod -R 775 storage bootstrap/cache

COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

EXPOSE 80

ENTRYPOINT ["docker-entrypoint.sh"]