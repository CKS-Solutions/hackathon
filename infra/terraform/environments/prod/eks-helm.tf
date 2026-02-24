# Providers Kubernetes e Helm para o cluster EKS (auth via aws eks get-token).
# Dependem do data.aws_eks_cluster.cluster (definido em main.tf).

provider "kubernetes" {
  host                   = data.aws_eks_cluster.cluster.endpoint
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.cluster.certificate_authority[0].data)

  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    args        = ["eks", "get-token", "--cluster-name", data.aws_eks_cluster.cluster.name]
  }
}

provider "helm" {
  kubernetes {
    host                   = data.aws_eks_cluster.cluster.endpoint
    cluster_ca_certificate = base64decode(data.aws_eks_cluster.cluster.certificate_authority[0].data)

    exec {
      api_version = "client.authentication.k8s.io/v1beta1"
      command     = "aws"
      args        = ["eks", "get-token", "--cluster-name", data.aws_eks_cluster.cluster.name]
    }
  }
}

# Namespace para o Argo CD (controle explícito antes do Helm release).
resource "kubernetes_namespace" "argocd" {
  metadata {
    name = "argocd"
  }
  depends_on = [data.aws_eks_cluster.cluster]
}

# AWS Load Balancer Controller (Entrega 6).
resource "helm_release" "aws_load_balancer_controller" {
  name       = "aws-load-balancer-controller"
  repository = "https://aws.github.io/eks-charts"
  chart      = "aws-load-balancer-controller"
  namespace  = "kube-system"
  version    = "1.8.1"

  set {
    name  = "clusterName"
    value = module.eks.cluster_id
  }
  set {
    name  = "serviceAccount.create"
    value = "true"
  }
  set {
    name  = "serviceAccount.annotations.eks\\.amazonaws\\.com/role-arn"
    value = aws_iam_role.lb_controller.arn
  }
  set {
    name  = "region"
    value = var.aws_region
  }
  set {
    name  = "vpcId"
    value = module.vpc.vpc_id
  }

  depends_on = [data.aws_eks_cluster.cluster]
}

# Argo CD.
resource "helm_release" "argocd" {
  name       = "argocd"
  repository = "https://argoproj.github.io/argo-helm"
  chart      = "argo-cd"
  namespace  = kubernetes_namespace.argocd.metadata[0].name
  version    = "7.7.10"

  depends_on = [kubernetes_namespace.argocd]
}

# Argo CD Application: overlay prod (video-system).
resource "kubernetes_manifest" "argocd_app_prod" {
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "video-system-prod"
      namespace = "argocd"
    }
    spec = {
      project = "default"
      source = {
        repoURL        = "https://github.com/CKS-Solutions/hackathon"
        path           = "infra/k8s/overlays/prod"
        targetRevision = "master"
      }
      destination = {
        server    = "https://kubernetes.default.svc"
        namespace = "video-system"
      }
      syncPolicy = {
        automated = {
          prune   = true
          selfHeal = true
        }
      }
    }
  }
  depends_on = [helm_release.argocd]
}

# Argo CD Application: monitoring stack (Prometheus + Grafana).
resource "kubernetes_manifest" "argocd_app_monitoring" {
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "monitoring-stack"
      namespace = "argocd"
    }
    spec = {
      project = "default"
      source = {
        repoURL        = "https://prometheus-community.github.io/helm-charts"
        chart          = "kube-prometheus-stack"
        targetRevision = "58.0.0"
        helm = {
          values = <<-EOT
            alertmanager:
              enabled: false
            nodeExporter:
              enabled: false
            prometheus:
              prometheusSpec:
                ignoreNamespaceSelectors: true
                serviceMonitorSelectorNilUsesHelmValues: false
                serviceMonitorSelector: {}
                retention: 7d
                resources:
                  requests:
                    memory: 400Mi
                    cpu: 100m
                  limits:
                    memory: 1Gi
            grafana:
              adminPassword: "admin"
              ingress:
                enabled: false
              resources:
                requests:
                  memory: 128Mi
                  cpu: 50m
          EOT
        }
      }
      destination = {
        server    = "https://kubernetes.default.svc"
        namespace = "monitoring"
      }
      syncPolicy = {
        automated = {
          prune   = true
          selfHeal = true
        }
        syncOptions = ["CreateNamespace=true", "ServerSideApply=true"]
      }
    }
  }
  depends_on = [helm_release.argocd]
}
