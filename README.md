# Hackathon — Sistema de Upload e Processamento de Vídeo

Sistema distribuído na AWS para upload, processamento assíncrono e notificação de vídeos, com autenticação JWT e orquestração em Kubernetes (EKS).

---

## Arquitetura

O sistema roda na **AWS**: o cluster **EKS** orquestra três microsserviços em Go. O processamento de vídeo é **assíncrono** (upload → fila → worker → notificação por e-mail).

```
    ┌─────────────────────────────────────────────────────────────────────────┐
    │                              AWS                                         │
    │  ┌───────────────────────────────────────────────────────────────────┐  │
    │  │                   Amazon EKS (Kubernetes)                           │  │
    │  │                                                                   │  │
    │  │   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐           │  │
    │  │   │  ms-auth    │   │  ms-video   │   │  ms-notify  │           │  │
    │  │   │  (JWT)      │   │  upload/    │   │  (email)    │           │  │
    │  │   │  register,  │   │  list/      │   │  SQS->SES   │           │  │
    │  │   │  login      │   │  download   │   │             │           │  │
    │  │   └──────┬──────┘   └──────┬──────┘   └──────┬──────┘           │  │
    │  │          │                 │                 │                   │  │
    │  │          │            ┌────▼────┐            │                   │  │
    │  │          │            │   SQS   │<────────────┘                   │  │
    │  │          │            │ (queue) │  evento video processado        │  │
    │  │          │            └────┬────┘                                  │  │
    │  │   ┌──────▼──────┐   ┌──────▼──────┐                                │  │
    │  │   │ PostgreSQL  │   │ S3+DynamoDB │  (videos + metadata)            │  │
    │  │   │ (auth)      │   │ (video)     │                                │  │
    │  │   └─────────────┘   └─────────────┘                                │  │
    │  │                                                                   │  │
    │  │   Ingress / ALB  <--- trafego HTTPS                               │  │
    │  └───────────────────────────────────────────────────────────────────┘  │
    │  Terraform: VPC, EKS, ECR, RDS, S3, SQS, DynamoDB, SES, IAM             │
    └─────────────────────────────────────────────────────────────────────────┘
                                 ▲
                                 | deploy (imagens + manifests)
    ┌────────────────────────────┴────────────────────────────┐
    │  GitHub Actions (CI/CD)                                   │
    │  - Build & push Docker -> ECR                            │
    │  - Commit image tags -> Argo CD (GitOps)                 │
    └─────────────────────────────────────────────────────────────┘
```

### Fluxo de dados

1. **Usuário** → registro/login em **ms-auth** → recebe JWT.
2. **Usuário** → envia vídeo para **ms-video** (com JWT).
3. **ms-video** → valida JWT, persiste em S3/DynamoDB, publica mensagem na fila **SQS**.
4. **Worker** (consumer no ms-video) consome a fila, processa o vídeo, atualiza status e chama **ms-notify**.
5. **ms-notify** → consome fila SQS, envia e-mail via **SES** e persiste status no **DynamoDB**.
6. **Usuário** → lista vídeos e status em **ms-video** e faz download (presigned URL S3) quando pronto.

---

## Microsserviços

| Serviço      | Responsabilidade | Stack | Documentação |
|-------------|------------------|--------|--------------|
| **ms-auth** | Registro, login e validação de JWT | Go, PostgreSQL, bcrypt, JWT | [packages/ms-auth/README.md](packages/ms-auth/README.md) |
| **ms-video** | Upload, listagem e download de vídeos; worker SQS (processamento) | Go, S3, SQS, DynamoDB, JWT | [packages/ms-video/README.md](packages/ms-video/README.md) |
| **ms-notify** | API de notificação e consumer SQS → envio de e-mail (SES) | Go, SQS, DynamoDB, SES | [packages/ms-notify/README.md](packages/ms-notify/README.md) |

### ms-auth

- **Endpoints:** `POST /register`, `POST /login`, `POST /validate`, `GET /health`.
- **Banco:** PostgreSQL (usuários); senhas com bcrypt, tokens JWT (HS256).
- **Local:** Docker Compose com Postgres; variáveis `DB_*`, `JWT_SECRET`, `JWT_EXPIRATION_HOURS`.

### ms-video

- **Endpoints:** `POST /video/upload`, `GET /video/list`, `GET /video/download?id=`, `GET /health` (todos protegidos por JWT, exceto health).
- **AWS:** S3 (raw + processed), SQS (fila de processamento), DynamoDB (metadados).
- **Processamento:** Consumer SQS baixa vídeo do S3, gera ZIP processado, atualiza status e notifica via ms-notify.
- **Local:** Docker Compose; opcional LocalStack para S3/SQS/DynamoDB; `S3_BUCKET_NAME`, `VIDEO_QUEUE_URL`, `JWT_SECRET`.

### ms-notify

- **Endpoints:** `POST /notification` (aceita pedido e publica na SQS), `GET /health`.
- **Consumer:** Lê mensagens da SQS, envia e-mail via SES e persiste resultado no DynamoDB.
- **Local:** Docker Compose com LocalStack (SQS, DynamoDB, SES); variáveis `AWS_*`, `SQS_QUEUE_URL`.

---

## Decisões técnicas

| Área | Escolha | Motivo |
|------|---------|--------|
| Cloud | AWS | EKS, ECR, RDS, S3, SQS, SES, DynamoDB integrados. |
| Orquestração | Kubernetes (EKS) | Escalabilidade e padrão para microsserviços. |
| IaC | Terraform | VPC, EKS, ECR, RDS, S3, SQS, DynamoDB, SES, IAM versionados. |
| Manifests K8s | Kustomize | Bases reutilizáveis + overlay prod (e dev para cluster local). |
| CD | Argo CD (GitOps) | Estado desejado no Git; reconciliação contínua. |
| CI | GitHub Actions | Build, testes, push para ECR; commit de image tags para Argo. |
| Mensageria | SQS | Processamento assíncrono de vídeo e notificações. |
| Banco auth | PostgreSQL (RDS) | Usuários e sessões. |
| Banco vídeo | DynamoDB + S3 | Metadados e arquivos. |
| Notificação | SES | Envio de e-mails. |

---

## Estrutura do repositório

```
├── packages/
│   ├── ms-auth/          # Autenticação (Go, PostgreSQL)
│   ├── ms-video/         # Upload e processamento de vídeo (Go, S3, SQS, DynamoDB)
│   └── ms-notify/        # Notificações (Go, SQS, DynamoDB, SES)
├── infra/
│   ├── terraform/        # IaC AWS (VPC, EKS, ECR, RDS, S3, SQS, etc.)
│   │   ├── modules/
│   │   └── environments/prod/
│   ├── k8s/              # Manifests e Kustomize (GitOps)
│   │   ├── base/         # ms-auth, ms-video, ms-notify
│   │   ├── overlays/dev/ # Cluster local
│   │   ├── overlays/prod/
│   │   └── argocd/       # Applications Argo CD
│   └── scripts/         # Bootstrap pós-Terraform (opcional)
└── .github/workflows/    # CI/CD (build-push, SonarQube, Terraform)
```

---

## Como rodar

### Local (por microsserviço)

Cada serviço pode subir com Docker Compose na própria pasta:

- **ms-auth:** `cd packages/ms-auth && docker-compose up --build`
- **ms-video:** `cd packages/ms-video && docker-compose up --build` (depende de fila S3/SQS; ver README do ms-video)
- **ms-notify:** `cd packages/ms-notify && docker-compose up --build` (LocalStack para SQS/DynamoDB/SES)

Ou usar o Compose da infra (stubs ou serviços reais): `docker compose -f infra/docker-compose.yml up -d --build`.

### Produção (AWS)

1. **Terraform:** `cd infra/terraform/environments/prod && terraform apply` (VPC, EKS, ECR, RDS, S3, SQS, DynamoDB, SES, IAM, etc.).
2. **Kubeconfig:** `aws eks update-kubeconfig --region <region> --name hackathon-prod`
3. **Bootstrap (Helm):** AWS Load Balancer Controller e Argo CD (ver infra — passo a passo “levantar tudo”).
4. **Argo CD:** aplicar as Applications que apontam para `infra/k8s/overlays/prod`; deploy dos três microsserviços no namespace `video-system`.
5. **Build-push:** push em `packages/**` ou workflow_dispatch → build das imagens no ECR e job que faz commit da tag da imagem nos patches → Argo CD faz o deploy da nova tag.

---

## Documentação dos microsserviços

3. **Bootstrap (Helm):** instalar AWS Load Balancer Controller e Argo CD no cluster.
- **ms-video (upload, download, S3, SQS, worker):** [packages/ms-video/README.md](packages/ms-video/README.md)
- **ms-notify (notificações, SQS, SES, DynamoDB):** [packages/ms-notify/README.md](packages/ms-notify/README.md)
