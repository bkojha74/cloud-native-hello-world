# GCP Cloud Run Deployment and Cleanup Guide

This document provides the complete, battle-tested, step-by-step procedure for deploying the **Cloud-Native Hello World** stack to Google Cloud Platform (GCP) using Terraform, Google Cloud Build, Google Artifact Registry, and Google Cloud Run, as well as the teardown/cleanup process to avoid unnecessary cloud costs.

---

## 1. Architecture Overview

```mermaid
flowchart TD
    subgraph Users ["Client Layer"]
        Browser["User Browser"]
    end

    subgraph GCP ["Google Cloud Platform (Region: asia-south1)"]
        subgraph CloudRun ["Cloud Run Services"]
            FE["helloworld-frontend<br/>(Nginx Reverse Proxy + React SPA)"]
            BE["helloworld-backend<br/>(Go 1.23 + Gin HTTP Server)"]
        end

        AR["Artifact Registry<br/>(helloworld-repo)"]
        CB["Cloud Build<br/>(Container Builder)"]
    end

    subgraph External ["Managed Database"]
        Atlas[("MongoDB Atlas Cluster<br/>(taskmanagercluster)")]
    end

    Browser -->|HTTPS :443| FE
    FE -->|/api/* (Reverse Proxy)| BE
    BE -->|Query messages collection| Atlas
    CB -->|Pushes backend & frontend images| AR
    AR -->|Supplies container images| CloudRun
```

---

## 2. Prerequisites

### 2.1 MongoDB Atlas Preparation
1. **Network Access**:
   - Go to [MongoDB Atlas](https://cloud.mongodb.com/) ➔ **Security** ➔ **Network Access**.
   - Click **Add IP Address** ➔ Select **ALLOW ACCESS FROM ANYWHERE (`0.0.0.0/0`)** ➔ **Confirm**.
   *(Required because Cloud Run instances use dynamic public IPs from Google's pool).*
2. **Database & Collection**:
   - Database name: `hello_world`
   - Collection name: `messages`
   - Seed Document:
     ```json
     {
       "_id": "1",
       "message": "Hello Bipin, how are you doing today?"
     }
     ```
3. **Connection String**:
   - Ensure you have your connection string ready, for example:
     `mongodb+srv://<username>:<password>@cluster.mongodb.net/hello_world?retryWrites=true&w=majority`

---

## 3. Cloud Shell Environment Setup

Open **Google Cloud Shell** in your browser and verify your active project:

```bash
gcloud config get-value project
# Expected output: terraform-gcp-lab-bipin (or your project ID)
```

### 3.1 Install Terraform (if not already installed)
```bash
wget -O - https://apt.releases.hashicorp.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(grep -oP '(?<=UBUNTU_CODENAME=).*' /etc/os-release || lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/hashicorp.list
sudo apt update && sudo apt install -y terraform
```

Verify installation:
```bash
terraform version
```

### 3.2 Enable Bootstrap GCP APIs
Terraform requires **Cloud Resource Manager** and **Service Usage** APIs to be active so it can manage other services:

```bash
gcloud services enable \
  cloudresourcemanager.googleapis.com \
  serviceusage.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com \
  run.googleapis.com \
  iam.googleapis.com
```

---

## 4. End-to-End Deployment Procedure

### Step 4.1: Clone Repository and Enter Terraform Directory
```bash
git clone https://github.com/bkojha74/cloud-native-hello-world.git
cd cloud-native-hello-world/terraform
```
*(If already cloned: `cd cloud-native-hello-world && git pull && cd terraform`)*

### Step 4.2: Configure Variables (`terraform.tfvars`)
Create your variable definition file from the example:
```bash
cp terraform.tfvars.example terraform.tfvars
```

Ensure `terraform.tfvars` contains:
```hcl
project_id = "terraform-gcp-lab-bipin"
region     = "asia-south1"

mongo_uri  = "mongodb+srv://<user>:<password>@taskmanagercluster.jwtpwgc.mongodb.net/hello_world?retryWrites=true&w=majority"
mongo_db   = "hello_world"

backend_image_tag  = "latest"
frontend_image_tag = "latest"
```

### Step 4.3: Initialize Terraform
```bash
terraform init
```

### Step 4.4: Provision Artifact Registry
Create the Docker repository first so Cloud Build has a target for pushing images:
```bash
terraform apply -target=google_artifact_registry_repository.docker_repo -auto-approve
```

### Step 4.5: Build and Push Docker Containers via Cloud Build
Build both the backend and frontend multi-stage containers directly in the cloud:

```bash
# 1. Build and push the Go Backend image
gcloud builds submit ../backend --tag $(terraform output -raw backend_image_target)

# 2. Build and push the React/Nginx Frontend image
gcloud builds submit ../frontend --tag $(terraform output -raw frontend_image_target)
```

Confirm that both builds finish with `STATUS: SUCCESS`.

### Step 4.6: Deploy Services to Cloud Run
Apply the full Terraform plan to spin up the Cloud Run services and grant public access:
```bash
terraform apply -auto-approve
```

Terraform outputs the live URLs:
```text
Outputs:

artifact_registry_repo = "helloworld-repo"
backend_image_target   = "asia-south1-docker.pkg.dev/terraform-gcp-lab-bipin/helloworld-repo/backend:latest"
backend_url            = "https://helloworld-backend-xxxxxx-el.a.run.app"
frontend_image_target  = "asia-south1-docker.pkg.dev/terraform-gcp-lab-bipin/helloworld-repo/frontend:latest"
frontend_url           = "https://helloworld-frontend-xxxxxx-el.a.run.app"
```

---

## 5. Verification and Smoke Testing

### 5.1 Backend Health Check
```bash
curl -i $(terraform output -raw backend_url)/health
```
*Expected response:* HTTP 200 with `{"status":"UP"}`.

### 5.2 Backend API Query
```bash
curl -i "$(terraform output -raw backend_url)/api/message?fromdb=1&cache=0&sql=0"
```
*Expected response:* HTTP 200 with JSON payload and `"source":"mongodb"`.

### 5.3 Frontend Reverse Proxy Test
```bash
curl -i "$(terraform output -raw frontend_url)/api/message?fromdb=1&cache=0&sql=0"
```
*Expected response:* HTTP 200 matching the backend response.

### 5.4 Browser Access
Open the `frontend_url` directly in your browser:
* Visual card displaying the dynamic message from MongoDB.
* Badge showing `SOURCE: MONGODB` and execution latency in milliseconds.
* Interactive **Fetch Message Again** button.

---

## 6. Cleanup & Teardown Procedure

To completely tear down the deployed resources and prevent ongoing Google Cloud charges:

### Step 6.1: Destroy Terraform-Managed Resources
In Cloud Shell from `cloud-native-hello-world/terraform`:
```bash
terraform destroy -auto-approve
```
This safely removes:
* `helloworld-frontend` Cloud Run service
* `helloworld-backend` Cloud Run service
* IAM public invoker roles (`allUsers` binding)
* `helloworld-repo` Artifact Registry repository and stored container images

### Step 6.2: Clean Up Cloud Build Artifact Storage Buckets (Optional)
Cloud Build creates temporary source tarball buckets in Google Cloud Storage:
```bash
gcloud storage rm --recursive gs://${PROJECT_ID}_cloudbuild/
```

### Step 6.3: Revoke MongoDB Atlas Network Access (Recommended for Security)
1. Go to [MongoDB Atlas Console](https://cloud.mongodb.com/) ➔ **Security** ➔ **Network Access**.
2. Find the `0.0.0.0/0` IP entry.
3. Click **Delete** to close public firewall access to your cluster.

---

## 7. Monitoring Billing, Costs & Resource Usage

### 7.1 Understanding Free Tier Coverage for this Stack
This deployment is heavily covered under Google Cloud's **Always Free Tier**:
* **Cloud Run**: First **2 million requests/month**, 360,000 vCPU-seconds, and 180,000 GiB-seconds are **100% Free** each month.
* **Scale-to-Zero**: Both `helloworld-frontend` and `helloworld-backend` are configured with `min_instance_count = 0`. When no requests are arriving, **0 container instances run and costs are $0.00**.
* **Cloud Build**: First **120 build-minutes per day** are free.
* **Artifact Registry**: Includes **0.5 GB/month** free storage.

### 7.2 How to Check Costs in Google Cloud Console
1. **Billing Overview**:
   - Go to [Google Cloud Billing Console](https://console.cloud.google.com/billing).
   - Select your linked billing account.
2. **Cost Breakdown & Reports**:
   - In the left sidebar, click **Reports**.
   - Under **Filters** on the right side:
     - **Projects**: Select `terraform-gcp-lab-bipin`.
     - **Time range**: Select `This month` or `Invoice month`.
     - **Group by**: Select `Service` to see costs broken down by Cloud Run, Artifact Registry, Cloud Build, and Cloud Storage.
3. **Cost Table**:
   - Click **Cost table** in the sidebar to view line-item charges and SKU details.
4. **Setting up a Budget Alert (Recommended)**:
   - In Billing, click **Budgets & alerts** ➔ **Create Budget**.
   - Set the scope to `terraform-gcp-lab-bipin`.
   - Set a target amount (e.g., ₹500 or $5.00).
   - Configure threshold notifications (e.g., email alerts at 50%, 90%, and 100%).

### 7.3 How to Monitor Usage & Live Metrics
#### Cloud Run Request Volume and Latency:
In Google Cloud Console:
* Navigate to **Cloud Run** ➔ click **helloworld-backend** (or **helloworld-frontend**).
* Click the **Metrics** tab to view:
  * **Request count** (requests/sec)
  * **Request latency** (P50, P95, P99)
  * **Billable container instance time**
  * **Container instance count** (shows autoscaling between 0 and 5 instances)
  * **CPU and Memory utilization**

Via Cloud Shell CLI:
```bash
# Check service status and active configuration
gcloud run services describe helloworld-backend --region asia-south1
gcloud run services describe helloworld-frontend --region asia-south1

# Check Artifact Registry storage usage
gcloud artifacts repositories describe helloworld-repo --location asia-south1
```

---

## 8. Troubleshooting & Lessons Learned

| Issue / Error | Root Cause | Solution |
|---|---|---|
| `403 Cloud Resource Manager API disabled` | Brand new GCP project lacks the bootstrap resource management API. | Run `gcloud services enable cloudresourcemanager.googleapis.com serviceusage.googleapis.com`. |
| `Reserved env names provided: PORT` | Cloud Run automatically injects `PORT`; explicit declaration is rejected. | Do not declare `PORT` in container `env` blocks; rely on `ports { container_port = 80 }`. |
| `Total memory < 512 Mi is not supported with cpu always allocated` | Cloud Run v2 requires at least 512Mi memory when allocating 1 CPU (`1000m`). | Set `limits = { cpu = "1000m", memory = "512Mi" }`. |
| `cannot destroy service without setting deletion_protection=false` | Cloud Run v2 enables accidental-deletion protection by default. | Set `deletion_protection = false` in both services and run `terraform apply` before `terraform destroy`. |
| `Unable to load message` (Frontend red error box) | MongoDB Atlas firewall blocking Cloud Run dynamic IP. | Add `0.0.0.0/0` in MongoDB Atlas Network Access. |
