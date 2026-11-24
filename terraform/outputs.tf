output "artifact_registry_repo" {
  description = "The created Artifact Registry Docker repository."
  value       = google_artifact_registry_repository.docker_repo.name
}

output "backend_image_target" {
  description = "Target Docker tag for backend image."
  value       = local.backend_image
}

output "frontend_image_target" {
  description = "Target Docker tag for frontend image."
  value       = local.frontend_image
}

output "backend_url" {
  description = "The public HTTPS URL of the Backend Cloud Run service."
  value       = google_cloud_run_v2_service.backend.uri
}

output "frontend_url" {
  description = "The public HTTPS URL of the Frontend Cloud Run service."
  value       = google_cloud_run_v2_service.frontend.uri
}
