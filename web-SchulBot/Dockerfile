# ── Stage 1: Build app + assets ───────────────────────────────────────────────
FROM php:8.4-apache AS build

RUN apt-get update && apt-get install -y --no-install-recommends \
        curl \
        git \
        unzip \
        libsqlite3-dev \
        libzip-dev \
    && docker-php-ext-install pdo pdo_sqlite zip opcache \
    && a2enmod rewrite headers \
    && rm -rf /var/lib/apt/lists/*

COPY --from=composer:2 /usr/bin/composer /usr/bin/composer

RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get update \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY composer.json composer.lock ./
RUN composer install --no-dev --prefer-dist --optimize-autoloader --no-interaction

COPY package.json package-lock.json ./
RUN npm ci

COPY . .

RUN php artisan wayfinder:generate --with-form || true
RUN npm run build


# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM php:8.4-apache

RUN apt-get update && apt-get install -y --no-install-recommends \
        libsqlite3-dev \
        libzip-dev \
        unzip \
    && docker-php-ext-install pdo pdo_sqlite zip opcache \
    && a2enmod rewrite headers \
    && rm -rf /var/lib/apt/lists/*

RUN sed -i 's|/var/www/html|/var/www/html/public|g' /etc/apache2/sites-available/000-default.conf \
    && echo 'ServerName localhost' >> /etc/apache2/apache2.conf

WORKDIR /var/www/html

COPY --from=build /app /var/www/html

RUN mkdir -p storage/framework/sessions storage/framework/views storage/framework/cache bootstrap/cache database \
    && chown -R www-data:www-data storage bootstrap/cache database \
    && chmod -R 775 storage bootstrap/cache

COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

EXPOSE 80

ENTRYPOINT ["docker-entrypoint.sh"]