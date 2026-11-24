# 1. Enable Required GCP APIs
resource "google_project_service" "apis" {
  for_each = toset([
    "run.googleapis.com",
    "artifactregistry.googleapis.com",
    "cloudbuild.googleapis.com",
    "iam.googleapis.com"
  ])

  service            = each.key
  disable_on_destroy = false
}

# 2. Artifact Registry for Docker Images
resource "google_artifact_registry_repository" "docker_repo" {
  depends_on    = [google_project_service.apis]
  location      = var.region
  repository_id = "helloworld-repo"
  description   = "Docker repository for Cloud-Native Hello World application"
  format        = "DOCKER"
}

locals {
  repo_base       = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.docker_repo.name}"
  backend_image  = "${local.repo_base}/backend:${var.backend_image_tag}"
  frontend_image = "${local.repo_base}/frontend:${var.frontend_image_tag}"
}

# 3. Backend Cloud Run Service
resource "google_cloud_run_v2_service" "backend" {
  depends_on = [google_artifact_registry_repository.docker_repo]
  name       = "helloworld-backend"
  location   = var.region
  ingress    = "INGRESS_TRAFFIC_ALL"

  template {
    scaling {
      min_instance_count = 0
      max_instance_count = 5
    }

    containers {
      image = local.backend_image

      ports {
        container_port = 8080
      }

      env {
        name  = "HTTP_ADDR"
        value = ":8080"
      }
      env {
        name  = "MONGO_URI"
        value = var.mongo_uri
      }
      env {
        name  = "MONGO_DB"
        value = var.mongo_db
      }
      env {
        name  = "LOG_LEVEL"
        value = "info"
      }

      resources {
        limits = {
          cpu    = "1000m"
          memory = "512Mi"
        }
      }

      startup_probe {
        http_get {
          path = "/health"
          port = 8080
        }
        initial_delay_seconds = 2
        period_seconds        = 5
        failure_threshold     = 3
      }
    }
  }
}

# Allow public unauthenticated access to Backend
resource "google_cloud_run_v2_service_iam_member" "backend_public" {
  project  = var.project_id
  location = google_cloud_run_v2_service.backend.location
  name     = google_cloud_run_v2_service.backend.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

# 4. Frontend Cloud Run Service
resource "google_cloud_run_v2_service" "frontend" {
  depends_on = [google_cloud_run_v2_service.backend]
  name       = "helloworld-frontend"
  location   = var.region
  ingress    = "INGRESS_TRAFFIC_ALL"

  template {
    scaling {
      min_instance_count = 0
      max_instance_count = 5
    }

    containers {
      image = local.frontend_image

      ports {
        container_port = 80
      }

      # Automatically wires backend Cloud Run URL into frontend Nginx reverse proxy!
      env {
        name  = "BACKEND_URL"
        value = google_cloud_run_v2_service.backend.uri
      }
      env {
        name  = "PORT"
        value = "80"
      }

      resources {
        limits = {
          cpu    = "1000m"
          memory = "256Mi"
        }
      }

      startup_probe {
        http_get {
          path = "/"
          port = 80
        }
        initial_delay_seconds = 2
        period_seconds        = 5
        failure_threshold     = 3
      }
    }
  }
}

# Allow public unauthenticated access to Frontend
resource "google_cloud_run_v2_service_iam_member" "frontend_public" {
  project  = var.project_id
  location = google_cloud_run_v2_service.frontend.location
  name     = google_cloud_run_v2_service.frontend.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
