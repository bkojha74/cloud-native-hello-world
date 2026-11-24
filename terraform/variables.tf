variable "project_id" {
  type        = string
  description = "The GCP Project ID where resources will be provisioned."
  default     = "terraform-gcp-lab-bipin"
}

variable "region" {
  type        = string
  description = "The GCP region to deploy resources into."
  default     = "asia-south1"
}

variable "mongo_uri" {
  type        = string
  description = "The MongoDB Atlas connection URI."
  sensitive   = true
}

variable "mongo_db" {
  type        = string
  description = "The MongoDB database name."
  default     = "hello_world"
}

variable "backend_image_tag" {
  type        = string
  description = "Tag for the backend image."
  default     = "latest"
}

variable "frontend_image_tag" {
  type        = string
  description = "Tag for the frontend image."
  default     = "latest"
}
