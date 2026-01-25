# Parse command line options
POSTGRES_USER="auth_admin"
while [[ $# -gt 0 ]]; do
  case $1 in
    --port|-p)
      POSTGRES_PORT="$2"
      shift 2
      ;;
    --host|-h)
      POSTGRES_HOST="$2"
      shift 2
      ;;
    --password|-P)
      POSTGRES_PASSWORD="$2"
      shift 2
      ;;
    --help)
      show_help
      ;;
    -*)
      echo "Unknown option: $1"
      show_help
      ;;
    *)
      if [ -z "$DATABASE_NAME" ]; then
        DATABASE_NAME="$1"
      else
        echo "Multiple database names not supported"
        show_help
      fi
      shift
      ;;
  esac
done

# Check if all required options are provided
if [ -z "$POSTGRES_PORT" ] || [ -z "$POSTGRES_HOST" ] || [ -z "$POSTGRES_PASSWORD" ]; then
  echo "Error: Missing required options."
  show_help
fi

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
  echo "Usage: $0 -p port -h host -P password [database_name|--help]"
  echo "   or: $0 --port port --host host --password password [database_name|--help]"
  echo "  -p, --port port         : PostgreSQL port"
  echo "  -h, --host host         : PostgreSQL host"
  echo "  -P, --password password : PostgreSQL password"
  echo "  database_name           : The name of the database to migrate (optional)."
  echo "                            If not provided, both postgres_auth and postgres_auth_dev will be migrated."
  echo "  --help                  : Display this help message."
  exit 0
}

# Function to URL encode a string
urlencode() {
    local string="${1}"
    local strlen=${#string}
    local encoded=""
    local pos c o
    for (( pos=0 ; pos<strlen ; pos++ )); do
        c=${string:$pos:1}
        case "$c" in
            [-_.~a-zA-Z0-9] ) o="${c}" ;;
            * )               printf -v o '%%%02x' "'$c"
        esac
        encoded+="${o}"
    done
    echo "${encoded}"
}

# Function to run migrations for a given database
migrate_db() {
  local db_name=$1
  local encoded_password=$(urlencode "$POSTGRES_PASSWORD")
  export GOTRUE_DB_DATABASE_URL="postgres://$POSTGRES_USER:$encoded_password@$POSTGRES_HOST:$POSTGRES_PORT/$db_name?sslmode=require"
  export GOTRUE_DB_NAMESPACE="auth"
  echo "Migrating database: $db_name"
  echo "GOTRUE_DB_DATABASE_URL: $GOTRUE_DB_DATABASE_URL"
  # Drop the existing migration table if it exists to allow re-running migrations
  psql "$GOTRUE_DB_DATABASE_URL" -c "DROP TABLE IF EXISTS auth.schema_migrations;" || echo "Warning: Could not drop table, psql may not be available"
  ./auth migrate -c ""
}

# Determine which databases to migrate
if [ -z "$DATABASE_NAME" ]; then
  # If no argument is provided, migrate both databases
  databases=("postgres_auth" "postgres_auth_dev")
else
  # Otherwise, migrate the specified database
  databases=("$DATABASE_NAME")
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