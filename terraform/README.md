# Terraform Infrastructure for Cloud-Native Hello World (GCP Cloud Run)

This directory provides Terraform (Infrastructure as Code) scripts to automatically provision:
* **GCP Services & APIs** (`run`, `artifactregistry`, `cloudbuild`, `iam`)
* **Google Artifact Registry Repository** (`helloworld-repo`)
* **Backend Cloud Run Service** (`helloworld-backend`) with MongoDB Atlas environment variables and health probes
* **Frontend Cloud Run Service** (`helloworld-frontend`) with Nginx reverse proxy automatically linked to the backend URL
* **IAM Policies** granting public HTTPS access (`allUsers` ➔ `roles/run.invoker`)

---

## Prerequisites

1. **MongoDB Atlas Network Access**:
   * Add `0.0.0.0/0` in MongoDB Atlas **Network Access** so Cloud Run dynamic IPs can connect.
   * Verify database `hello_world` and collection `messages` has `{ _id: "1", message: "..." }`.
2. **Cloud Shell Environment**:
   * Install Terraform in Cloud Shell if not present:
     ```bash
     wget -O - https://apt.releases.hashicorp.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg
     echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(grep -oP '(?<=UBUNTU_CODENAME=).*' /etc/os-release || lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/hashicorp.list
     sudo apt update && sudo apt install -y terraform
     ```
   * Enable Bootstrap GCP APIs:
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

## Deployment Workflow (via Cloud Shell)

### 1. Navigate to the terraform directory
```bash
cd cloud-native-hello-world/terraform
```

### 2. Configure variables
```bash
cp terraform.tfvars.example terraform.tfvars
# Verify project_id, region (e.g. asia-south1), and mongo_uri in terraform.tfvars
```

### 3. Initialize Terraform
```bash
terraform init
```

### 4. Provision Artifact Registry first
```bash
terraform apply -target=google_artifact_registry_repository.docker_repo -auto-approve
```

### 5. Build & push container images using Cloud Build
```bash
# Build and push backend image
gcloud builds submit ../backend --tag $(terraform output -raw backend_image_target)

# Build and push frontend image
gcloud builds submit ../frontend --tag $(terraform output -raw frontend_image_target)
```

### 6. Apply full Terraform plan
```bash
terraform apply -auto-approve
```

### 7. Access your application
Terraform will output the live URLs:
```text
Outputs:

backend_url  = "https://helloworld-backend-xxxxxx-el.a.run.app"
frontend_url = "https://helloworld-frontend-xxxxxx-el.a.run.app"
```

Open `frontend_url` in your browser to view the application!

---

## Monitoring Billing & Resource Usage

### 1. View Costs & Invoices in GCP Console
* Open the **[Google Cloud Billing Console](https://console.cloud.google.com/billing)**.
* Click **Reports** in the left menu, filter by project `terraform-gcp-lab-bipin`, and group by **Service** to see charges for Cloud Run, Artifact Registry, and Cloud Build.
* (Recommended) Click **Budgets & alerts** to configure a small monthly threshold alert (e.g. ₹500 / $5).

### 2. View Service Usage & Metrics
* In Cloud Console ➔ **Cloud Run** ➔ click `helloworld-backend` or `helloworld-frontend` ➔ **Metrics** to monitor:
  * Request counts and latency percentiles (P50/P95/P99)
  * Billable container instance time
  * Active container instances (scales between 0 and 5)
* Via Cloud Shell CLI:
  ```bash
  # Check service status and active configuration
  gcloud run services describe helloworld-backend --region asia-south1
  gcloud run services describe helloworld-frontend --region asia-south1
  ```

---

## Clean Up / Destroy Resources

To tear down all resources and avoid ongoing costs:

```bash
# 1. Destroy Terraform-managed resources (Cloud Run & Artifact Registry)
terraform destroy -auto-approve

# 2. (Optional) Remove MongoDB Atlas 0.0.0.0/0 IP whitelist rule in Atlas console
```
