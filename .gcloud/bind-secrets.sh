#!/bin/bash

# Set your variables
PROJECT_NUMBER=$GCP_PROJECT_NUMBER
PROJECT_ID=$GCP_PROJECT_ID
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
