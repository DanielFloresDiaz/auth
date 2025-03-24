# Set default values for PostgreSQL connection
: ${POSTGRES_PORT:=5432}
: ${POSTGRES_DB:=postgres_auth}
: ${POSTGRES_USER:=auth_admin}

MIGRATIONS_PATH="../liquibase_migrations"

show_help() {
  echo "Usage: $0 [database_name|--help]"
  echo "  database_name: The name of the database to migrate (optional)."
  echo "                 If not provided, both postgres_auth and postgres_auth_dev will be migrated."
  echo "  --help         : Display this help message."
  exit 0
}

migrate_db() {
  local db_name=$1
  echo "Migrating database: $db_name"
  liquibase update --changelog-file=$MIGRATIONS_PATH/changelog-public.sql --url=jdbc:postgresql://$POSTGRES_HOST:$POSTGRES_PORT/$db_name --username=$POSTGRES_USER --password=$POSTGRES_PASSWORD
  liquibase update --changelog-file=$MIGRATIONS_PATH/changelog-auth.sql --url=jdbc:postgresql://$POSTGRES_HOST:$POSTGRES_PORT/$db_name --username=$POSTGRES_USER --password=$POSTGRES_PASSWORD
  liquibase update --changelog-file=$MIGRATIONS_PATH/changelog-procedures-auth.xml --url=jdbc:postgresql://$POSTGRES_HOST:$POSTGRES_PORT/$db_name --username=$POSTGRES_USER --password=$POSTGRES_PASSWORD
}

# Determine which databases to migrate
if [ "$1" == "--help" ]; then
  show_help
elif [ -z "$1" ]; then
  # If no argument is provided, migrate the default database
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

# Run migrations for each database
for db in "${databases[@]}"; do
  migrate_db "$db"
done

# EXAMPLE:
# export POSTGRES_HOST=
# export POSTGRES_PORT=
# export POSTGRES_USER=
# export POSTGRES_PASSWORD=