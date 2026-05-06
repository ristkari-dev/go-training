#!/usr/bin/env bash
set -euo pipefail

# Idempotent GCP bootstrap for the go-training slides deployment.
# Re-runnable: each step either creates the resource or no-ops if it already exists.

PROJECT_ID="${PROJECT_ID:-ristkari-dev}"
REGION="${REGION:-europe-north1}"
AR_REPO="${AR_REPO:-go-training}"
SERVICE_NAME="${SERVICE_NAME:-go-training-slides}"
SA_NAME="${SA_NAME:-github-deploy}"
SA_EMAIL="${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"
WIF_POOL="${WIF_POOL:-github-actions}"
WIF_PROVIDER="${WIF_PROVIDER:-github}"
GITHUB_REPO="${GITHUB_REPO:-ristkari-dev/go-training}"

bold()  { printf '\033[1m%s\033[0m\n' "$*"; }
note()  { printf '  → %s\n' "$*"; }

bold "Project:        $PROJECT_ID"
bold "Region:         $REGION"
bold "AR repo:        $AR_REPO"
bold "Service:        $SERVICE_NAME"
bold "Service acct:   $SA_EMAIL"
bold "WIF pool:       $WIF_POOL"
bold "WIF provider:   $WIF_PROVIDER"
bold "GitHub repo:    $GITHUB_REPO"
echo

bold "1. Enabling required APIs"
gcloud services enable \
    artifactregistry.googleapis.com \
    iamcredentials.googleapis.com \
    run.googleapis.com \
    sts.googleapis.com \
    --project="$PROJECT_ID"

bold "2. Creating Artifact Registry repo (if missing)"
if gcloud artifacts repositories describe "$AR_REPO" \
        --location="$REGION" --project="$PROJECT_ID" >/dev/null 2>&1; then
    note "repo $AR_REPO already exists, skipping"
else
    gcloud artifacts repositories create "$AR_REPO" \
        --repository-format=docker \
        --location="$REGION" \
        --description="go-training container images" \
        --project="$PROJECT_ID"
fi

bold "3. Creating service account (if missing)"
if gcloud iam service-accounts describe "$SA_EMAIL" --project="$PROJECT_ID" >/dev/null 2>&1; then
    note "service account $SA_EMAIL already exists, skipping"
else
    gcloud iam service-accounts create "$SA_NAME" \
        --display-name="GitHub Actions deploy for go-training" \
        --project="$PROJECT_ID"
fi

bold "4. Granting roles to the service account"
for role in \
    roles/artifactregistry.writer \
    roles/run.admin \
    roles/iam.serviceAccountUser \
; do
    note "binding $role"
    gcloud projects add-iam-policy-binding "$PROJECT_ID" \
        --member="serviceAccount:$SA_EMAIL" \
        --role="$role" \
        --condition=None \
        --quiet >/dev/null
done

bold "5. Creating Workload Identity Federation pool (if missing)"
if gcloud iam workload-identity-pools describe "$WIF_POOL" \
        --location=global --project="$PROJECT_ID" >/dev/null 2>&1; then
    note "pool $WIF_POOL already exists, skipping"
else
    gcloud iam workload-identity-pools create "$WIF_POOL" \
        --location=global \
        --display-name="GitHub Actions" \
        --project="$PROJECT_ID"
fi

POOL_NAME=$(gcloud iam workload-identity-pools describe "$WIF_POOL" \
    --location=global --project="$PROJECT_ID" --format='value(name)')

bold "6. Creating WIF OIDC provider for GitHub (if missing)"
if gcloud iam workload-identity-pools providers describe "$WIF_PROVIDER" \
        --location=global --workload-identity-pool="$WIF_POOL" \
        --project="$PROJECT_ID" >/dev/null 2>&1; then
    note "provider $WIF_PROVIDER already exists, skipping"
else
    gcloud iam workload-identity-pools providers create-oidc "$WIF_PROVIDER" \
        --location=global \
        --workload-identity-pool="$WIF_POOL" \
        --display-name="GitHub OIDC" \
        --issuer-uri="https://token.actions.githubusercontent.com" \
        --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository,attribute.ref=assertion.ref" \
        --attribute-condition="assertion.repository == '${GITHUB_REPO}'" \
        --project="$PROJECT_ID"
fi

PROVIDER_NAME=$(gcloud iam workload-identity-pools providers describe "$WIF_PROVIDER" \
    --location=global --workload-identity-pool="$WIF_POOL" \
    --project="$PROJECT_ID" --format='value(name)')

bold "7. Allowing the GitHub repo to impersonate the SA"
gcloud iam service-accounts add-iam-policy-binding "$SA_EMAIL" \
    --role=roles/iam.workloadIdentityUser \
    --member="principalSet://iam.googleapis.com/${POOL_NAME}/attribute.repository/${GITHUB_REPO}" \
    --project="$PROJECT_ID" \
    --condition=None \
    --quiet >/dev/null

echo
bold "Done. Add these as GitHub repository secrets:"
echo
echo "  GCP_PROJECT_ID                = $PROJECT_ID"
echo "  GCP_WORKLOAD_IDENTITY_PROVIDER = $PROVIDER_NAME"
echo "  GCP_SERVICE_ACCOUNT_EMAIL     = $SA_EMAIL"
echo
bold "Then create the Cloud Run service for the first time (the deploy workflow"
bold "expects the service to exist). One of:"
echo
echo "  # Option A (recommended): apply the checked-in spec with a placeholder image:"
echo "  gcloud run services replace deploy/cloudrun.yaml \\"
echo "      --region=$REGION --project=$PROJECT_ID"
echo "  # (You'll then need to push to main once to deploy a real image — until then"
echo "  #  the service has 'invalid image' status, which is fine.)"
echo
echo "  # Option B: deploy a placeholder image now to materialise the service:"
echo "  gcloud run deploy $SERVICE_NAME \\"
echo "      --image=gcr.io/cloudrun/hello \\"
echo "      --region=$REGION --project=$PROJECT_ID \\"
echo "      --allow-unauthenticated"
echo
bold "Then map the custom domain (Cloudflare DNS will need a CNAME — see deploy/README.md):"
echo "  gcloud beta run domain-mappings create \\"
echo "      --service=$SERVICE_NAME \\"
echo "      --domain=golang.ristkari.dev \\"
echo "      --region=$REGION --project=$PROJECT_ID"
