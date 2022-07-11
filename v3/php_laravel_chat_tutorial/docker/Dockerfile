FROM composer:2.1.5 AS composer

FROM php:8.0.9-fpm-alpine3.13 as php_builds

RUN apk --update add --no-cache \
      autoconf \
      postgresql-dev \
      libpng-dev \
      freetype-dev \
      libjpeg-turbo-dev \
      libzip-dev \
      zip \
      && docker-php-ext-install opcache pdo_pgsql pgsql gd zip

FROM php:8.0.9-fpm-alpine3.13

ENV PATH=/root/.composer/vendor/bin:/usr/local/bin/pear:$PATH
ENV COMPOSER_ALLOW_SUPERUSER=1
ENV PHP_EXT_DIR=/usr/local/lib/php/extensions/no-debug-non-zts-20200930

COPY --from=composer /usr/bin/composer /usr/bin/composer
COPY --from=php_builds $PHP_EXT_DIR/opcache.so $PHP_EXT_DIR/opcache.so
COPY --from=php_builds $PHP_EXT_DIR/pdo_pgsql.so $PHP_EXT_DIR/pdo_pgsql.so
COPY --from=php_builds $PHP_EXT_DIR/pgsql.so $PHP_EXT_DIR/pgsql.so
COPY --from=php_builds $PHP_EXT_DIR/gd.so $PHP_EXT_DIR/gd.so
COPY --from=php_builds $PHP_EXT_DIR/zip.so $PHP_EXT_DIR/zip.so

RUN apk --update add --no-cache \
    bash \
    curl \
    git \
    libpng-dev \
    freetype-dev \
    libjpeg-turbo-dev \
    libzip-dev \
    zip \
    nodejs \
    npm \
    postgresql-client \
    python2 \
    rsync \
    nano \
  && mkdir /usr/share/man/man1 /usr/share/man/man7 \
  && docker-php-ext-enable opcache pdo_pgsql pgsql gd zip

WORKDIR /app

COPY . /app/