#!/bin/bash
set -e

echo "Copy composer dependencies..."
composer install
composer dumpautoload

until psql -h $DB_HOST -U $DB_USERNAME -d app_db -c 'SELECT 1' > /dev/null; do sleep 1; done;
psql -h $DB_HOST -U $DB_USERNAME -lqt | cut -d \| -f 1 | grep -qw app_db  &> /dev/null

if [[ ! -f .env ]]; then
  echo "Copy .env file..."
  cp .env.example .env
fi

php artisan key:generate

echo "Npm install..."
npm install --no-cache
npm run dev

echo "Database migrations..."
php artisan migrate --force --seed

/usr/local/sbin/php-fpm

$@
