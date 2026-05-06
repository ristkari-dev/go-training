# Deploying go-training slides

The slides site is built into a Docker image, pushed to Google Artifact
Registry, and served by Cloud Run. A push to `main` triggers
`.github/workflows/deploy.yml` which does the build → push → deploy.

The custom domain `https://golang.ristkari.dev/` points at the Cloud Run
service via a Cloudflare CNAME (DNS-only).

This file documents the **one-time setup** you do once per project (or after
an `iam` cleanup), not what runs on every push.

## Prerequisites

- `gcloud` CLI authenticated as an account with Owner or sufficient roles
  on the `ristkari-dev` project.
- `gh` CLI authenticated against `ristkari-dev/go-training` for setting
  repo secrets (or you can set them in the GitHub UI).
- Cloudflare access for the `ristkari.dev` zone.

## Step 1 — bootstrap GCP

```bash
./deploy/setup.sh
```

The script is idempotent — re-runnable. It:

1. Enables the required APIs (`artifactregistry`, `run`, `iamcredentials`, `sts`).
2. Creates Artifact Registry repo `go-training` in `europe-north1`.
3. Creates service account `github-deploy@ristkari-dev.iam.gserviceaccount.com`.
4. Grants it `roles/artifactregistry.writer`, `roles/run.admin`,
   `roles/iam.serviceAccountUser`.
5. Creates Workload Identity Federation pool `github-actions` and OIDC
   provider `github`, scoped to this repo only.
6. Binds the SA so GitHub Actions on this repo can impersonate it via WIF.
7. Prints the three values you need as GitHub repo secrets.

## Step 2 — set GitHub repo secrets

Take the three values printed by `setup.sh` and set them as repo secrets:

```bash
gh secret set GCP_PROJECT_ID                 -b "ristkari-dev"
gh secret set GCP_WORKLOAD_IDENTITY_PROVIDER -b "<value-from-setup-output>"
gh secret set GCP_SERVICE_ACCOUNT_EMAIL      -b "github-deploy@ristkari-dev.iam.gserviceaccount.com"
```

Or set them in the GitHub UI: **Settings → Secrets and variables → Actions**.

## Step 3 — first-time service materialisation

The deploy workflow uses `gcloud run services replace deploy/cloudrun.yaml`.
That works only if a service named `go-training-slides` already exists. Create
it once with a placeholder image:

```bash
gcloud run deploy go-training-slides \
    --image=gcr.io/cloudrun/hello \
    --region=europe-north1 \
    --project=ristkari-dev \
    --platform=managed \
    --allow-unauthenticated \
    --port=8080
```

After this, the GitHub Actions deploy will replace it with the real image on
every push to `main`.

## Step 4 — map the custom domain

```bash
gcloud beta run domain-mappings create \
    --service=go-training-slides \
    --domain=golang.ristkari.dev \
    --region=europe-north1 \
    --project=ristkari-dev
```

Output includes a CNAME target like `ghs.googlehosted.com.`.

## Step 5 — Cloudflare DNS

Add a CNAME record in the Cloudflare dashboard for `ristkari.dev`:

| Type  | Name   | Target                      | Proxy status |
|-------|--------|-----------------------------|--------------|
| CNAME | golang | `ghs.googlehosted.com`      | **DNS only** (gray cloud) |

Cloudflare proxying (orange cloud) breaks Cloud Run's domain mapping because
Cloud Run handles its own TLS at the mapped hostname. Leave it gray.

Propagation typically takes <5 minutes. Verify with:

```bash
dig golang.ristkari.dev CNAME +short
gcloud beta run domain-mappings describe \
    --domain=golang.ristkari.dev \
    --region=europe-north1 \
    --project=ristkari-dev
```

The `domain-mappings describe` output shows `READY=True` and serves on HTTPS
once Google has provisioned a managed certificate (a few minutes after the
DNS propagates).

## Verifying a deploy

After a push to `main`:

```bash
# Watch the deploy workflow
gh run watch

# When green, hit the service:
curl -sS -I https://golang.ristkari.dev/
```

Expected: `HTTP/2 200`, `content-type: text/html`.

The Cloud Run service URL (without the custom domain) is:

```bash
gcloud run services describe go-training-slides \
    --region=europe-north1 --project=ristkari-dev \
    --format='value(status.url)'
```

## Rolling back

```bash
gcloud run services list --project=ristkari-dev
gcloud run revisions list --service=go-training-slides \
    --region=europe-north1 --project=ristkari-dev
gcloud run services update-traffic go-training-slides \
    --to-revisions=go-training-slides-<previous-revision>=100 \
    --region=europe-north1 --project=ristkari-dev
```
