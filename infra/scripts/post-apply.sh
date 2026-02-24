#!/usr/bin/env bash
# Passos manuais após `terraform apply` (prod): kubeconfig, Helm LB controller, Argo CD, Applications.
# Rodar da raiz do repositório: ./infra/scripts/post-apply.sh
# Pré-requisitos: terraform apply já rodado em infra/terraform/environments/prod; aws, kubectl, helm no PATH.
set -e

REPO_ROOT="${REPO_ROOT:-$(cd "$(dirname "$0")/../.." && pwd)}"
TF_DIR="$REPO_ROOT/infra/terraform/environments/prod"
AWS_REGION="${AWS_REGION:-us-east-1}"

echo "==> Usando REPO_ROOT=$REPO_ROOT TF_DIR=$TF_DIR AWS_REGION=$AWS_REGION"

echo "==> Obtendo outputs do Terraform..."
cluster_name=$(terraform -chdir="$TF_DIR" output -raw cluster_id)
lb_role_arn=$(terraform -chdir="$TF_DIR" output -raw lb_controller_role_arn)
vpc_id=$(terraform -chdir="$TF_DIR" output -raw vpc_id)

echo "==> 1. Atualizando kubeconfig (cluster=$cluster_name)..."
aws eks update-kubeconfig --region "$AWS_REGION" --name "$cluster_name"

echo "==> 2. Verificando acesso ao cluster..."
kubectl get nodes --request-timeout=10s

echo "==> 3. Instalando AWS Load Balancer Controller (Helm)..."
helm repo add eks https://aws.github.io/eks-charts 2>/dev/null || true
helm repo update
helm upgrade --install aws-load-balancer-controller eks/aws-load-balancer-controller \
  -n kube-system \
  --set clusterName="$cluster_name" \
  --set serviceAccount.create=true \
  --set "serviceAccount.annotations.eks\.amazonaws\.com/role-arn=$lb_role_arn" \
  --set region="$AWS_REGION" \
  --set vpcId="$vpc_id" \
  --wait --timeout 3m

echo "==> 4. Instalando Argo CD (Helm)..."
kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -
helm repo add argo https://argoproj.github.io/argo-helm 2>/dev/null || true
helm repo update
helm upgrade --install argocd argo/argo-cd -n argocd --wait --timeout 3m

echo "==> 5. Aguardando pods do Argo CD..."
kubectl wait --for=condition=Ready pod -l app.kubernetes.io/name=argocd-server -n argocd --timeout=120s 2>/dev/null || true
kubectl get pods -n argocd

echo "==> 6. Aplicando Applications do Argo CD..."
kubectl apply -f "$REPO_ROOT/infra/k8s/argocd/application-prod.yaml"
kubectl apply -f "$REPO_ROOT/infra/k8s/argocd/application-monitoring.yaml"
kubectl apply -f "$REPO_ROOT/infra/k8s/argocd/application-loki.yaml"
kubectl apply -f "$REPO_ROOT/infra/k8s/argocd/application-promtail.yaml"

echo "==> Concluído. Argo CD fará o sync automático (prod, monitoring, loki, promtail)."
echo "    Pods: kubectl get pods -n video-system && kubectl get pods -n monitoring"
echo "    Ingress: kubectl get ingress -n video-system"
