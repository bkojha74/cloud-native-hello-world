# GCP Terraform Lab 01 - Creating a Google Cloud Storage Bucket

## Objective

Setup a new GCP project and provision the first cloud resource using Terraform.

---

# Prerequisites

## GCP Account

- Google Cloud account created
- Free trial activated
- Billing account configured
- Project created

Project:

```text
terraform-gcp-lab-bipin
```

## Cloud Shell

Launch Cloud Shell from the top-right terminal icon in GCP Console.

Verify project:

```bash
gcloud config get-value project
```

Expected:

```text
terraform-gcp-lab-bipin
```

---

# Verify GCloud Configuration

Check configuration:

```bash
gcloud config list
```

Expected:

```text
[core]
account = bkobaby2004@gmail.com
project = terraform-gcp-lab-bipin
```

---

# Enable Required APIs

Enable Compute Engine, Storage and IAM APIs.

```bash
gcloud services enable \
    compute.googleapis.com \
    storage.googleapis.com \
    iam.googleapis.com
```

Verify:

```bash
gcloud services list --enabled | grep -E "compute|storage|iam"
```

Expected services:

```text
compute.googleapis.com
iam.googleapis.com
storage.googleapis.com
```

---

# Create Terraform Workspace

Create working directory:

```bash
mkdir ~/terraform-lab
cd ~/terraform-lab
```

Verify:

```bash
pwd
```

Expected:

```text
/home/<user>/terraform-lab
```

---

# Install Terraform

Add HashiCorp repository:

```bash
wget -O- https://apt.releases.hashicorp.com/gpg | \
gpg --dearmor | \
sudo tee /usr/share/keyrings/hashicorp-archive-keyring.gpg > /dev/null
```

Configure repository:

```bash
echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | \
sudo tee /etc/apt/sources.list.d/hashicorp.list
```

Update packages:

```bash
sudo apt update
```

Install Terraform:

```bash
sudo apt install -y terraform
```

Verify installation:

```bash
terraform version
```

Expected:

```text
Terraform v1.17.2
```

---

# Create Terraform Configuration

Create file:

```bash
vi main.tf
```

Content:

```hcl
terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 7.0"
    }
  }
}

provider "google" {
  project = "terraform-gcp-lab-bipin"
  region  = "asia-south1"
}

resource "google_storage_bucket" "demo" {
  name     = "bkobaby2004-terraform-lab-2026-001"
  location = "ASIA-SOUTH1"
}
```

Verify file:

```bash
cat main.tf
```

---

# Initialize Terraform

Initialize providers:

```bash
terraform init
```

Expected:

```text
Terraform has been successfully initialized!
```

Files created:

```text
.terraform/
.terraform.lock.hcl
```

---

# Validate Configuration

```bash
terraform validate
```

Expected:

```text
Success! The configuration is valid.
```

---

# Generate Execution Plan

```bash
terraform plan
```

Expected:

```text
Plan: 1 to add, 0 to change, 0 to destroy.
```

Terraform Resource:

```text
google_storage_bucket.demo
```

---

# Apply Changes

Execute:

```bash
terraform apply
```

Approve:

```text
yes
```

Expected:

```text
Apply complete!
Resources: 1 added, 0 changed, 0 destroyed.
```

---

# Verify Bucket Creation

List buckets:

```bash
gcloud storage buckets list
```

Expected bucket:

```text
gs://bkobaby2004-terraform-lab-2026-001/
```

---

# Inspect Terraform State

List resources managed by Terraform:

```bash
terraform state list
```

Expected:

```text
google_storage_bucket.demo
```

Inspect state:

```bash
terraform state show google_storage_bucket.demo
```

---

# Verify Idempotency

Run plan again:

```bash
terraform plan
```

Expected:

```text
No changes.
Your infrastructure matches the configuration.
```

This confirms Terraform state and cloud resources are synchronized.

---

# Cleanup

Destroy resources:

```bash
terraform destroy
```

Approve:

```text
yes
```

Verify bucket removal:

```bash
gcloud storage buckets list
```

---

# Key Concepts Learned

## Google Cloud

- Billing Account
- Project
- APIs
- Cloud Storage
- Cloud Shell

## Terraform

- Provider
- Resource
- State
- Plan
- Apply
- Destroy
- Idempotency

## Commands Practiced

```bash
gcloud config list
gcloud config get-value project
gcloud services enable
terraform init
terraform validate
terraform plan
terraform apply
terraform destroy
terraform state list
terraform state show
```

---

# Lab Outcome

Successfully provisioned and managed a Google Cloud Storage Bucket using Terraform.

Resource:

```text
google_storage_bucket.demo
```

Bucket:

```text
bkobaby2004-terraform-lab-2026-001
```

Project:

```text
terraform-gcp-lab-bipin
```

Status:

✅ GCP Setup Complete

✅ Terraform Setup Complete

✅ First Infrastructure-as-Code Deployment Complete