set -e

echo "DEBUG: DB_SOURCE is '$DB_SOURCE'"

echo "run db migration"
/app/migrate -path /app/migrations -database "$DB_SOURCE" -verbose up

echo "start the app"
exec "$@"