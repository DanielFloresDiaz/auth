# Set default values for PostgreSQL connection
: ${POSTGRES_PORT:=5432}
: ${POSTGRES_DB:=postgres_auth}
: ${POSTGRES_USER:=auth_admin}

MIGRATIONS_PATH="../liquibase_migrations"

liquibase update --changelog-file=$MIGRATIONS_PATH/changelog-public.sql --url=jdbc:postgresql://$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_DB --username=$POSTGRES_USER --password=$POSTGRES_PASSWORD
liquibase update --changelog-file=$MIGRATIONS_PATH/changelog-auth.sql --url=jdbc:postgresql://$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_DB --username=$POSTGRES_USER --password=$POSTGRES_PASSWORD
liquibase update --changelog-file=$MIGRATIONS_PATH/changelog-procedures-auth.xml --url=jdbc:postgresql://$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_DB --username=$POSTGRES_USER --password=$POSTGRES_PASSWORD

# EXAMPLE:
# export POSTGRES_HOST=
# export POSTGRES_PORT=
# export POSTGRES_DB=
# export POSTGRES_USER=
# export POSTGRES_PASSWORD=
# ./migrations/gcloud_liquibase_migrations.sh