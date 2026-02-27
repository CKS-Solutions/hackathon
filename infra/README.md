# Infraestrutura

Provisionamento (Terraform), workloads no Kubernetes (Kustomize), GitOps (Argo CD) e pipelines de CI/CD (GitHub Actions).

---

## 1. Arquitetura

O sistema roda na AWS. O cluster **EKS** orquestra os microsserviços. Processamento de vídeo é assíncrono: upload e fila via mensageria; o usuário é notificado quando o processamento termina.

```
                    ┌─────────────────────────────────────────────────────────────────┐
                    │                         AWS                                       │
                    │  ┌───────────────────────────────────────────────────────────┐  │
                    │  │                    Amazon EKS (Kubernetes)                    │  │
                    │  │   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐   │  │
                    │  │   │  ms-auth     │    │  ms-video    │    │  ms-notify  │   │  │
                    │  │   │  (JWT)       │    │  upload/     │    │  (email)    │   │  │
                    │  │   │              │    │  list/download│   │             │   │  │
                    │  │   └──────┬───────┘    └──────┬──────┘    └──────┬──────┘   │  │
                    │  │          │                   │                   │          │  │
                    │  │          │              ┌────▼────┐              │          │  │
                    │  │          │              │ RabbitMQ│◄─────────────┘          │  │
                    │  │          │              └────┬────┘                          │  │
                    │  │   ┌───────▼───────┐    ┌─────▼─────┐                        │  │
                    │  │   │  PostgreSQL   │    │ PostgreSQL│                          │  │
                    │  │   │  (auth)       │    │ (video)   │                          │  │
                    │  │   └───────────────┘    └───────────┘                        │  │
                    │  │   Ingress / ALB  ◄─── tráfego HTTP(S)                         │  │
                    │  └───────────────────────────────────────────────────────────┘  │  │
                    │  Terraform: VPC, EKS, ECR, RDS, S3, IAM                          │
                    └─────────────────────────────────────────────────────────────────┘
                                         ▲
                    ┌────────────────────┴────────────────────┐
                    │  GitHub Actions: build/push ECR; Argo CD aplica Kustomize       │
                    └─────────────────────────────────────────┘
```

---

## 2. Componentes

| Serviço      | Responsabilidade | API principal | Banco |
|-------------|------------------|---------------|--------|
| **ms-auth** | Registro, login, JWT | `POST /register`, `POST /login` | PostgreSQL |
| **ms-video** | Upload, listagem, download; publica na fila | `POST /video`, `GET /video`, `GET /video/:id/download` | PostgreSQL |
| **ms-notify** | Notificações (e-mail) | `POST /notification` | — |

- **Mensageria:** RabbitMQ — filas para eventos (ex.: `video.uploaded`, `video.processed`).
- **Kubernetes:** namespace `video-system` para os apps; Ingress path-based (`/auth`, `/video`, `/notify`) com ALB (AWS Load Balancer Controller).
- **CI/CD:** GitHub Actions faz build e push para ECR; Argo CD aplica o overlay prod a partir do Git.
- **Secrets:** Kubernetes Secrets (ConfigMaps/Secrets no cluster).

---

## 3. Fluxo de dados

1. Usuário faz login no ms-auth e recebe JWT.
2. Usuário envia vídeo para ms-video (com JWT).
3. ms-video valida JWT, persiste metadados e publica mensagem na fila.
4. Worker consome a fila, processa o vídeo e publica "processado" ou chama ms-notify.
5. ms-notify envia e-mail ao usuário.
6. Usuário lista vídeos e faz download quando pronto.

---

## 4. Stack técnica

| Área | Tecnologia |
|------|------------|
| Cloud | AWS |
| Orquestração | Kubernetes (EKS) |
| IaC | Terraform (backend S3) |
| Manifests K8s | Kustomize (base + overlay prod; overlay dev para cluster local) |
| CD | Argo CD (GitOps) |
| CI | GitHub Actions (build, Trivy, push ECR) |
| Mensageria | RabbitMQ |
| Banco | PostgreSQL por serviço |
| Imagens | Docker → ECR |

---

## 5. Estrutura de pastas

```
infra/
├── README.md
├── terraform/
│   ├── modules/          # vpc, eks, ecr, etc.
│   └── environments/
│       ├── dev/
│       └── prod/
├── k8s/
│   ├── base/             # ms-auth, ms-video, ms-notify, rabbitmq
│   ├── overlays/
│   │   ├── dev/
│   │   └── prod/
│   ├── argocd/           # Applications (prod, monitoring)
│   └── monitoring/
├── docker-compose.yml
└── scripts/
```

Workflows do GitHub em `.github/workflows/` na raiz do repositório.

---

## 6. O que está implementado

VPC, EKS, ECR, IAM (OIDC para ECR; IAM user para pipelines Terraform). AWS Load Balancer Controller (Helm). Argo CD (Helm). Applications Argo CD: `video-system-prod` (overlay prod no namespace `video-system`) e `monitoring-stack` (kube-prometheus-stack no namespace `monitoring`). Os três serviços usam ServiceAccount `video-system/app` com IRSA (role `hackathon-prod-app`). Recursos AWS do ms-notify: SQS, DynamoDB, SES (Terraform em `environments/prod/ms-notify.tf`). Bucket S3 para vídeos: `cks-hackathon-video-system`.

---

## 7. Pré-requisitos para pipelines

No GitHub (Settings → Secrets and variables → Actions):

- **Secrets:** `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` (IAM user para Terraform plan/apply/destroy); `AWS_ROLE_ARN` (output `github_actions_role_arn`, para build-push ECR); `TF_STATE_BUCKET` (bucket S3 do state).
- **Variables:** `AWS_REGION`, `TF_STATE_REGION`, `ECR_REGISTRY` (prefixo do ECR, ex.: `123456789012.dkr.ecr.us-east-1.amazonaws.com`).

Terraform no CI usa IAM user (access key). Build-push usa OIDC (`AWS_ROLE_ARN`).

---

## 8. Procedimentos

### 8.1 Levantar o ambiente após terraform destroy

1. Rodar a pipeline **Terraform Apply** ou, local: `cd infra/terraform/environments/prod && terraform apply`.
2. Kubeconfig: `aws eks update-kubeconfig --region <region> --name hackathon-prod`.
3. Instalar AWS Load Balancer Controller (Helm) com `vpcId`, `clusterName`, `lb_controller_role_arn` (outputs do Terraform).
4. Instalar Argo CD (Helm) no namespace `argocd`.
5. Aplicar as Applications: `kubectl apply -f infra/k8s/argocd/application-prod.yaml` e `kubectl apply -f infra/k8s/argocd/application-monitoring.yaml`.
6. Verificar: `kubectl get pods -n video-system`, `kubectl get ingress -n video-system`. Testar: `curl -s http://<alb-hostname>/auth`.

### 8.2 Antes do terraform destroy

- Deletar o Ingress para o AWS Load Balancer Controller remover o ALB:
  ```bash
  kubectl delete ingress video-system -n video-system
  ```
- Esvaziar o bucket S3 de vídeos (senão o destroy falha):
  ```bash
  aws s3 rm s3://cks-hackathon-video-system --recursive
  ```
  Com versionamento habilitado, remover também versões e delete markers antes.

### 8.3 Docker Compose (local)

Na raiz do repositório:

```bash
docker compose -f infra/docker-compose.yml up -d --build
```

Variáveis opcionais: `cp infra/.env.example infra/.env` e editar. Parar: `docker compose -f infra/docker-compose.yml down`.

### 8.4 Cluster local (kind / minikube)

Manifests em `infra/k8s/` (base + overlay dev). Exemplo com kind:

```bash
kind create cluster
docker build -t ms-stub:local packages/ms-stub
kind load docker-image ms-stub:local
kubectl apply -k infra/k8s/overlays/dev
```

Remover: `kubectl delete -k infra/k8s/overlays/dev` e `kind delete cluster`.
