# Parse command line options
POSTGRES_USER="postgres"
LOG_LEVEL="ERROR"  # Default log level
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
    --log-level|-l)
      LOG_LEVEL="$2"
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

# Determine migrations path
if [ -d "./liquibase_migrations" ]; then
  MIGRATIONS_PATH="./liquibase_migrations"
else
  MIGRATIONS_PATH="../liquibase_migrations"
fi

show_help() {
  echo "Usage: $0 -p port -h host -P password [database_name|--help]"
  echo "   or: $0 --port port --host host --password password [database_name|--help]"
  echo "  -p, --port port         : PostgreSQL port"
  echo "  -h, --host host         : PostgreSQL host"
  echo "  -P, --password password : PostgreSQL password"
  echo "  -l, --log-level level   : Liquibase log level (default: INFO)"
  echo "  database_name           : The name of the database to migrate (optional)."
  echo "                            If not provided, both postgres_auth and postgres_auth_dev will be migrated."
  echo "  --help                  : Display this help message."
  exit 0
}

migrate_db() {
  local db_name=$1
  echo "Migrating database: $db_name"
  liquibase update --log-level=$LOG_LEVEL --changelog-file=$MIGRATIONS_PATH/changelog-public.sql --url=jdbc:postgresql://$POSTGRES_HOST:$POSTGRES_PORT/$db_name --username=$POSTGRES_USER --password=$POSTGRES_PASSWORD
  liquibase update --log-level=$LOG_LEVEL --changelog-file=$MIGRATIONS_PATH/changelog-auth.sql --url=jdbc:postgresql://$POSTGRES_HOST:$POSTGRES_PORT/$db_name --username=$POSTGRES_USER --password=$POSTGRES_PASSWORD
  liquibase update --log-level=$LOG_LEVEL --changelog-file=$MIGRATIONS_PATH/changelog-rls-auth.sql --url=jdbc:postgresql://$POSTGRES_HOST:$POSTGRES_PORT/$db_name --username=$POSTGRES_USER --password=$POSTGRES_PASSWORD
  liquibase update --log-level=$LOG_LEVEL --changelog-file=$MIGRATIONS_PATH/changelog-index-auth.sql --url=jdbc:postgresql://$POSTGRES_HOST:$POSTGRES_PORT/$db_name --username=$POSTGRES_USER --password=$POSTGRES_PASSWORD
  liquibase update --log-level=$LOG_LEVEL --changelog-file=$MIGRATIONS_PATH/changelog-grants-auth.sql --url=jdbc:postgresql://$POSTGRES_HOST:$POSTGRES_PORT/$db_name --username=$POSTGRES_USER --password=$POSTGRES_PASSWORD
  liquibase update --log-level=$LOG_LEVEL --changelog-file=$MIGRATIONS_PATH/changelog-procedures-auth.xml --url=jdbc:postgresql://$POSTGRES_HOST:$POSTGRES_PORT/$db_name --username=$POSTGRES_USER --password=$POSTGRES_PASSWORD
}

# Determine which databases to migrate
if [ -z "$DATABASE_NAME" ]; then
  # If no database name is provided, migrate the default databases
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
echo "LOG_LEVEL: $LOG_LEVEL"
echo "DATABASES: ${databases[@]}"
echo "MIGRATIONS_PATH: $MIGRATIONS_PATH"

# Run migrations for each database
for db in "${databases[@]}"; do
  migrate_db "$db"
done