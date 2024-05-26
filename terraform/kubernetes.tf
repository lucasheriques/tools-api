provider "kubernetes" {
  host                   = "https://${data.google_container_cluster.autopilot_cluster.endpoint}"
  token                  = data.google_client_config.default.access_token
  cluster_ca_certificate = base64decode(data.google_container_cluster.autopilot_cluster.master_auth.0.cluster_ca_certificate)
}

resource "kubernetes_deployment" "go_rest_api" {
  metadata {
    name      = "go-rest-api"
    namespace = "default"
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "go-rest-api"
      }
    }

    template {
      metadata {
        labels = {
          app = "go-rest-api"
        }
      }

      spec {
        container {
          name  = "go-rest-api"
          image = "us-central1-docker.pkg.dev/lucasfaria-tools-api/my-repo/tools-lucasfaria-dev:latest"
          port {
            container_port = 4000
          }

          readiness_probe {
            http_get {
              path = "/v1/healthcheck"
              port = 4000
            }
            initial_delay_seconds = 5
            period_seconds        = 10
          }
          liveness_probe {
            http_get {
              path = "/v1/healthcheck"
              port = 4000
            }
            initial_delay_seconds = 15
            period_seconds        = 20
          }
        }
      }
    }
  }
}

resource "kubernetes_deployment" "gotenberg" {
  metadata {
    name      = "gotenberg"
    namespace = "default"
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "gotenberg"
      }
    }

    template {
      metadata {
        labels = {
          app = "gotenberg"
        }
      }

      spec {
        container {
          name  = "gotenberg"
          image = "gotenberg/gotenberg:8"
          port {
            container_port = 3000
          }
          security_context {
            privileged  = false
            run_as_user = 1001
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "go_rest_api" {
  metadata {
    name      = "go-rest-api"
    namespace = "default"
  }

  spec {
    selector = {
      app = kubernetes_deployment.go_rest_api.spec[0].template[0].metadata[0].labels["app"]
    }

    port {
      port        = 80
      target_port = 4000
    }

    type = "NodePort"
  }
}

resource "kubernetes_service" "gotenberg" {
  metadata {
    name      = "gotenberg"
    namespace = "default"
  }

  spec {
    selector = {
      app = kubernetes_deployment.gotenberg.spec[0].template[0].metadata[0].labels["app"]
    }

    port {
      port        = 80
      target_port = 3000
    }

    type = "ClusterIP"
  }
}
