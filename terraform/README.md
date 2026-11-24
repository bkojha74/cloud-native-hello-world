# Terraform Infrastructure for Cloud-Native Hello World (GCP Cloud Run)

This directory provides Terraform (Infrastructure as Code) scripts to automatically provision:
* **GCP Services & APIs** (`run`, `artifactregistry`, `cloudbuild`, `iam`)
* **Google Artifact Registry Repository** (`helloworld-repo`)
* **Backend Cloud Run Service** (`helloworld-backend`) with MongoDB Atlas environment variables and health probes
* **Frontend Cloud Run Service** (`helloworld-frontend`) with Nginx reverse proxy automatically linked to the backend URL
* **IAM Policies** granting public HTTPS access (`allUsers` ➔ `roles/run.invoker`)

---

## Deployment Workflow (via Cloud Shell)

### 1. Navigate to the terraform directory
```bash
cd cloud-native-hello-world/terraform
```

### 2. Configure variables
```bash
cp terraform.tfvars.example terraform.tfvars
# Verify project_id and mongo_uri in terraform.tfvars
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

backend_url = "https://helloworld-backend-xxxxxx-el.a.run.app"
frontend_url = "https://helloworld-frontend-xxxxxx-el.a.run.app"
```

Open `frontend_url` in your browser to view the application!

---

## Clean Up / Destroy Resources

To tear down all resources and avoid ongoing costs:
```bash
terraform destroy -auto-approve
```
