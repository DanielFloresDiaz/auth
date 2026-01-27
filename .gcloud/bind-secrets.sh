#!/bin/bash

# Set your variables
PROJECT_NUMBER=$GCP_PROJECT_NUMBER
PROJECT_ID=$GCP_PROJECT_ID

# Flags: short -n/-i and long --number/--id override PROJECT_NUMBER and PROJECT_ID
while [ "$#" -gt 0 ]; do
  case "$1" in
    -n|--number)
      if [ -z "$2" ] || [ "${2#-}" != "$2" ]; then
        echo "Error: $1 requires a non-flag argument" >&2
        exit 1
      fi
      PROJECT_NUMBER=$2
      shift 2
      ;;
    -i|--id)
      if [ -z "$2" ] || [ "${2#-}" != "$2" ]; then
        echo "Error: $1 requires a non-flag argument" >&2
        exit 1
      fi
      PROJECT_ID=$2
      shift 2
      ;;
    *)
      echo "Unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

NAMESPACE=solomon
KSA_NAME=auth-service-account

# Define the secrets from your YAML file
SECRETS=(
  "AUTH_SECRET"
  "AUTH_DB_ADMIN_PASSWORD"
  "GOTRUE_EXTERNAL_GOOGLE_SECRET"
  "GOTRUE_EXTERNAL_GOOGLE_CLIENT_ID"
)

echo "Binding IAM roles to the Kubernetes service account: $KSA_NAME in namespace: $NAMESPACE"
echo "Project ID: $PROJECT_ID"
echo "Project Number: $PROJECT_NUMBER"
# Loop through the secrets and apply the IAM binding
for SECRET_NAME in "${SECRETS[@]}"; do
  echo "Binding secret: $SECRET_NAME"
  gcloud secrets add-iam-policy-binding $SECRET_NAME \
    --project=$PROJECT_ID \
    --role=roles/secretmanager.secretAccessor \
    --member="principal://iam.googleapis.com/projects/$PROJECT_NUMBER/locations/global/workloadIdentityPools/$PROJECT_ID.svc.id.goog/subject/ns/$NAMESPACE/sa/$KSA_NAME"
done

# Grant cloudsql.client role to the service account
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --role=roles/cloudsql.client \
  --member="principal://iam.googleapis.com/projects/$PROJECT_NUMBER/locations/global/workloadIdentityPools/$PROJECT_ID.svc.id.goog/subject/ns/$NAMESPACE/sa/$KSA_NAME"
