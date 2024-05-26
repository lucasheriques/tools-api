provider "google" {
  project     = "lucasfaria-tools-api"
  region      = "us-central1"
  credentials = file("terraform-sa-key.json")
}

resource "google_container_cluster" "autopilot" {
  name             = "lucasfaria-tools-cluster"
  location         = "us-central1"
  enable_autopilot = true
}

data "google_client_config" "default" {}

data "google_container_cluster" "autopilot_cluster" {
  name     = google_container_cluster.autopilot.name
  location = google_container_cluster.autopilot.location
}

resource "google_compute_global_address" "default" {
  name = "my-global-address"
}

resource "kubernetes_ingress_v1" "go_rest_api_ingress" {
  metadata {
    name      = "go-rest-api-ingress"
    namespace = "default"
    annotations = {
      "kubernetes.io/ingress.class" : "gce"
      "kubernetes.io/ingress.global-static-ip-name" = google_compute_global_address.default.name
    }
  }

  spec {
    default_backend {
      service {
        name = kubernetes_service.go_rest_api.metadata[0].name
        port {
          number = kubernetes_service.go_rest_api.spec[0].port[0].port
        }
      }
    }

    rule {
      host = "tools.lucasfaria.dev"
      http {
        path {
          path      = "/"
          path_type = "Prefix"
          backend {
            service {
              name = kubernetes_service.go_rest_api.metadata[0].name
              port {
                number = kubernetes_service.go_rest_api.spec[0].port[0].port
              }
            }
          }
        }
      }
    }
  }

  depends_on = [
    kubernetes_service.go_rest_api,
    google_compute_global_address.default,
  ]
}

output "kubeconfig" {
  value = google_container_cluster.autopilot.endpoint
}
