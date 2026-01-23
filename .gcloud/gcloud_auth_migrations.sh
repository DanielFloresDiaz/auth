# Set default values for PostgreSQL connection
: ${POSTGRES_PORT:=5432}
: ${POSTGRES_USER:=postgres}
: ${POSTGRES_HOST:=localhost}
: ${POSTGRES_PASSWORD:=root}

MIGRATIONS_PATH="./migrations"

export GOTRUE_DB_DRIVER="postgres"
#export DATABASE_URL="postgres://$POSTGRES_USER:"$POSTGRES_PASSWORD"@$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_DB"
export GOTRUE_DB_MIGRATIONS_PATH=$MIGRATIONS_PATH

# Set fake environment variables to hack the migrations
export API_EXTERNAL_URL="http://localhost:9999"
export GOTRUE_SITE_URL="http://localhost:3000"
export GOTRUE_JWT_SECRET="mysecret"

# Function to display help message
show_help() {
  echo "Usage: $0 [database_name|--help]"
  echo "  database_name: The name of the database to migrate (optional)."
  echo "                 If not provided, both postgres_auth and postgres_auth_dev will be migrated."
  echo "  --help         : Display this help message."
  exit 0
}

# Function to run migrations for a given database
migrate_db() {
  local db_name=$1
  export DATABASE_URL="postgres://$POSTGRES_USER:"$POSTGRES_PASSWORD"@$POSTGRES_HOST:$POSTGRES_PORT/$db_name"
  echo "Migrating database: $db_name"
  echo "DATABASE_URL: $DATABASE_URL"
  ./auth migrate -c ""
}

# Determine which databases to migrate
if [ "$1" == "--help" ]; then
  show_help
elif [ -z "$1" ]; then
  # If no argument is provided, migrate both databases
  databases=("postgres_auth" "postgres_auth_dev")
else
  # Otherwise, migrate the specified database
  databases=("$1")
fi

# Print the values
echo "POSTGRES_PORT: $POSTGRES_PORT"
echo "POSTGRES_USER: $POSTGRES_USER"
echo "POSTGRES_HOST: $POSTGRES_HOST"
echo "POSTGRES_PASSWORD: $POSTGRES_PASSWORD"
echo "DATABASES: ${databases[@]}"
echo "MIGRATIONS_PATH: $MIGRATIONS_PATH"

# Build the project using make build
make build

# Run migrations for each database
for db in "${databases[@]}"; do
  migrate_db "$db"
done