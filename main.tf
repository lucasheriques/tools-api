provider "google" {
  project     = "lucasfaria-tools-api"
  region      = "us-central1"
  credentials = file("terraform-sa-key.json")
}

resource "google_container_cluster" "autopilot" {
  name     = "my-autopilot-cluster"
  location = "us-central1"

  enable_autopilot = true
}

data "google_client_config" "default" {}

data "google_container_cluster" "autopilot_cluster" {
  name     = google_container_cluster.autopilot.name
  location = google_container_cluster.autopilot.location
}

output "kubeconfig" {
  value = google_container_cluster.autopilot.endpoint
}
