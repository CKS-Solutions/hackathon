# Infraestrutura — Sistema de Upload e Processamento de Vídeo

Este diretório concentra toda a infraestrutura do projeto: provisionamento (Terraform), definição de workloads no Kubernetes (Kustomize), GitOps (Argo CD) e pipelines de CI/CD (GitHub Actions).

---

## 1. Visão geral da arquitetura

O sistema roda na **AWS**, com o cluster **Kubernetes (EKS)** orquestrando os microsserviços. O processamento de vídeo é **assíncrono**: upload e enfileiramento via **mensageria**, e o usuário é **notificado** quando o processamento termina.

```
                    ┌─────────────────────────────────────────────────────────────────┐
                    │                         AWS                                       │
                    │  ┌───────────────────────────────────────────────────────────┐  │
                    │  │                    Amazon EKS (Kubernetes)                    │  │
                    │  │                                                              │  │
                    │  │   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐   │  │
                    │  │   │  ms-auth     │    │  ms-video    │    │  ms-notify  │   │  │
                    │  │   │  (JWT)       │    │  upload/     │    │  (email)    │   │  │
                    │  │   │              │    │  list/download│   │             │   │  │
                    │  │   └──────┬───────┘    └──────┬──────┘    └──────┬──────┘   │  │
                    │  │          │                   │                   │          │  │
                    │  │          │              ┌────▼────┐              │          │  │
                    │  │          │              │ RabbitMQ│◄─────────────┘          │  │
                    │  │          │              │ (queue) │   evento "vídeo        │  │
                    │  │          │              └────┬────┘    processado"          │  │
                    │  │          │                   │                              │  │
                    │  │   ┌───────▼───────┐    ┌─────▼─────┐                        │  │
                    │  │   │  PostgreSQL   │    │ PostgreSQL│  (um DB por serviço)   │  │
                    │  │   │  (auth)        │    │ (video)   │                        │  │
                    │  │   └───────────────┘    └───────────┘                        │  │
                    │  │                                                              │  │
                    │  │   Ingress / ALB  ◄─── tráfego HTTPS                          │  │
                    │  └───────────────────────────────────────────────────────────┘  │ │
                    │                                                                   │
                    │  Terraform: VPC, EKS, RDS/instances, S3, secrets, IAM             │
                    └─────────────────────────────────────────────────────────────────┘
                                         ▲
                                         │ deploy (imagens + manifests)
                    ┌────────────────────┴────────────────────┐
                    │  GitHub Actions (CI/CD)                  │
                    │  • Build & push Docker → ECR             │
                    │  • Argo CD aplica Kustomize (GitOps)     │
                    └─────────────────────────────────────────┘
```

---

## 2. Responsabilidades por componente

### 2.1 Microsserviços (Go)

| Serviço      | Responsabilidade | API principal | Banco |
|-------------|------------------|---------------|--------|
| **ms-auth** | Registro, login e emissão de JWT | `POST /register`, `POST /login` | PostgreSQL (usuários) |
| **ms-video** | Upload, listagem e download de vídeos processados; publica evento na fila | `POST /video`, `GET /video`, `GET /video/:id/download`, consumo da fila `/video/upload` | PostgreSQL (vídeos, status) |
| **ms-notify** | Envio de notificações (ex.: e-mail quando o vídeo termina) | `POST /notification` (chamado por ms-video ou worker) | Opcional (log de notificações) |

### 2.2 Mensageria (processamento assíncrono)

- **RabbitMQ** (ou alternativa compatível): fila para eventos de vídeo (ex.: `video.uploaded`, `video.processed`).
- **Fluxo resumido:** cliente faz upload → **ms-video** grava metadados e publica mensagem → worker/processor consome, extrai frames → ao terminar, publica "processado" → **ms-notify** envia e-mail (ou **ms-video** chama **ms-notify**).
- Garante que em picos as requisições não se percam: a fila absorve a carga e o processamento escala com workers.

### 2.3 Kubernetes (EKS)

- **Orquestração** dos três microsserviços, RabbitMQ e bancos (ou referência a RDS).
- **Ingress + ALB** para expor HTTPS e rotear para os serviços.
- **Namespaces** sugeridos: `video-system` (apps) e opcionalmente `video-system-infra` (RabbitMQ, etc.).
- **ConfigMaps/Secrets** para config e credenciais; preferência por External Secrets ou Sealed Secrets em produção.

### 2.4 CI/CD (GitHub Actions)

- **CI:** testes, lint, build das imagens Docker e push para **Amazon ECR**.
- **CD:** GitOps via **Argo CD** — o repositório é a fonte da verdade; Argo CD aplica os manifests (Kustomize) no EKS. O pipeline pode apenas garantir que a imagem nova está no ECR e que o Git (branch/tag) está atualizado; Argo CD faz o deploy.

### 2.5 GitOps (Argo CD + Kustomize)

- **Kustomize:** organiza bases e overlay **prod** para o EKS (Deployments, Services, Ingress, etc.). O overlay **dev** existe apenas para testes em cluster local (kind/minikube).
- **Argo CD** aponta para o repositório (pasta de manifests/Kustomize) e mantém o cluster alinhado ao Git. Deploys e rollbacks passam por commit/PR, com histórico e auditoria.

---

## 3. Fluxo de dados (resumido)

1. **Usuário** → login em **ms-auth** → recebe JWT.
2. **Usuário** → envia vídeo para **ms-video** (com JWT).
3. **ms-video** → valida JWT (ms-auth ou lib local), persiste metadados, publica mensagem **video.uploaded** na fila.
4. **Worker/processor** (dentro ou fora do ms-video) consome a fila, processa o vídeo (extração de frames), atualiza status e publica **video.processed** (ou chama **ms-notify**).
5. **ms-notify** → envia e-mail (ou outro canal) ao usuário.
6. **Usuário** → lista vídeos e status em **ms-video** e faz download do ZIP quando pronto.

---

## 4. Decisões técnicas

| Decisão | Escolha | Motivo |
|--------|---------|--------|
| **Cloud** | AWS | Requisito; ecossistema EKS, ECR, RDS, S3 bem integrado. |
| **Orquestração** | Kubernetes (EKS) | Escalabilidade, padrão para microsserviços, alinhado ao hackathon. |
| **IaC (cloud)** | Terraform (ou OpenTofu) | Provisionar VPC, EKS, ECR, RDS, S3, IAM de forma versionada e reproduzível. |
| **Manifests K8s** | Kustomize | Bases reutilizáveis + overlay prod para EKS (dev só para cluster local). |
| **CD** | Argo CD (GitOps) | Estado desejado no Git; reconciliação contínua; rollback por revert de commit. |
| **CI** | GitHub Actions | Integração nativa com o repositório; build, test e push para ECR. |
| **Mensageria** | RabbitMQ | Atende “RabbitMQ, Kafka ou similar”; simples de operar no K8s; filas e exchanges bem definidas. |
| **Banco** | PostgreSQL por serviço | Alinhado à stack recomendada; um schema ou instância por serviço para bounded context. |
| **Imagens** | Docker → ECR | Build multi-stage (Docker expert), scan de vulnerabilidades no pipeline; ECR próximo ao EKS. |
| **Secrets** | Kubernetes Secrets + (futuro) External Secrets / Vault | Começar com Secrets; evoluir para operador que injeta da AWS Secrets Manager ou Vault. |

---

## 5. Estrutura de pastas sugerida (infra)

```
infra/
├── README.md                 # Este arquivo (arquitetura + plano de entregas)
├── terraform/                # IaC AWS (VPC, EKS, ECR, etc.)
│   ├── modules/
│   │   ├── eks/
│   │   ├── vpc/
│   │   └── ...
│   ├── environments/
│   │   ├── dev/
│   │   └── prod/
│   └── ...
├── k8s/                      # Manifests e Kustomize (fonte do GitOps)
│   ├── base/                 # Bases comuns (ms-auth, ms-video, ms-notify, rabbitmq, etc.)
│   │   ├── ms-auth/
│   │   ├── ms-video/
│   │   ├── ms-notify/
│   │   └── ...
│   └── overlays/
│       ├── dev/
│       └── prod/
└── .github/                  # Opcional: workflows podem ficar em .github/workflows/ na raiz do repo
```

*(Os workflows do GitHub Actions costumam ficar em `.github/workflows/` na raiz do repositório.)*

---

## 6. Plano de entregas incrementais

Cada entrega é **pequena**, **testável** e prepara a próxima. Você pode validar em ambiente local (ex.: kind/minikube) antes de usar EKS.

**Escopo do hackathon:** uso apenas do ambiente **prod** na AWS (EKS, ECR, etc.). Não há staging; deploys e pipelines de infra apontam direto para prod. O overlay **dev** em Kustomize existe só para rodar em cluster local (Entrega 2); o deploy no EKS usa somente o overlay **prod**.

**Pré-requisito obrigatório antes de rodar o projeto (Terraform ou pipelines):** As pipelines de Terraform (plan/apply/destroy) usam um **IAM user** com access key, e não OIDC, para permitir rodar destroy e apply em sequência pelo CI sem passo local. Você deve **criar um IAM user** na AWS (ex.: `github-terraform-ci`) com permissão para aplicar/destruir a infra (recomendado: policy **AdministratorAccess** no escopo do hackathon), gerar uma **access key** e configurar no repositório (Settings → Secrets and variables → Actions) os **Secrets** `AWS_ACCESS_KEY_ID` e `AWS_SECRET_ACCESS_KEY`. Sem isso, os workflows de Terraform no CI falham ao obter credenciais. O build-push (ECR) continua usando OIDC (secret `AWS_ROLE_ARN`), criado após o primeiro apply do Terraform.

#### Resumo do que temos até agora (Entregas 1–8)

| Entrega | O que está pronto |
|--------|--------------------|
| **1** | Docker + Compose local (ms-stub, RabbitMQ, Postgres). |
| **2** | Kustomize (base + overlay dev) para cluster local (kind/minikube). |
| **3** | Terraform: VPC, EKS (prod), backend S3 para state. |
| **4** | ECR (repos ms-auth, ms-video, ms-notify), OIDC para GitHub Actions, pipeline **build-push** (build, Trivy, push com tags `sha` e `latest`). |
| **4b** | Pipelines **terraform-apply** e **terraform-destroy** (prod), usando IAM user (access key) para Terraform. |
| **5** | Overlay prod com imagens ECR; deploy no EKS (por Argo CD ou um `kubectl apply -k` inicial). |
| **6** | Ingress + ALB: subnet tags no VPC, IRSA do AWS Load Balancer Controller, Helm do controller (com `vpcId`), Ingress path-based (`/auth`, `/video`, `/notify`). |
| **7** | Argo CD instalado no EKS; Application apontando para `infra/k8s/overlays/prod` (branch master, repo público); sync automático — não é mais necessário rodar `kubectl apply -k` para prod. |
| **8** | Observabilidade: Application Argo CD (Helm) para kube-prometheus-stack no namespace `monitoring`; ms-stub expõe `/metrics`; ServiceMonitor no overlay prod para scrape dos três microsserviços; Grafana acessível por port-forward. |

**Componentes atuais:** VPC, EKS, ECR, IAM (OIDC para ECR + IAM user para Terraform), AWS Load Balancer Controller (Helm), Argo CD (Helm), Applications `video-system-prod` e `monitoring-stack`. A aplicação (ms-auth, ms-video, ms-notify com imagem real no ECR) é deployada pelo Argo CD; o stack de observabilidade (Prometheus + Grafana) é instalado via Argo CD a partir do chart Helm. Os três serviços usam uma **ServiceAccount genérica** `video-system/app` com IRSA (role `hackathon-prod-app`); output Terraform `app_irsa_role_arn`. Recursos AWS do **ms-notify** (SQS, DynamoDB, SES) estão em `infra/terraform`: módulos genéricos em `modules/sqs-queue`, `dynamodb-table`, `ses-email-identity` e instância em `environments/prod/ms-notify.tf`; use os outputs `ms_notify_sqs_queue_url`, `ms_notify_dynamodb_table_name` e `ms_notify_ses_sender_email` para configurar o Deployment do ms-notify (env ou External Secrets).

#### Antes do `terraform destroy`: evitar DependencyViolation (subnet / IGW)

O EKS e o **Ingress** criam recursos de rede que **não estão no Terraform**: o ALB (Application Load Balancer) é criado pelo AWS Load Balancer Controller quando existe um Ingress no cluster. Esse ALB mantém ENIs nas subnets; o **NAT Gateway** (no módulo VPC) mantém um Elastic IP. Se você rodar `terraform destroy` direto, pode aparecer:

- `DependencyViolation: The subnet has dependencies and cannot be deleted`
- `DependencyViolation: Network has some mapped public address(es). Please unmap before detaching the gateway`

**Ordem recomendada antes de rodar destroy:**

1. **Com o cluster ainda de pé:** apagar o Ingress para o controller destruir o ALB:
   ```bash
   kubectl delete ingress video-system -n video-system
   ```
   Aguardar 2–3 minutos para o ALB sumir no console AWS (EC2 → Load Balancers).

2. **Rodar o destroy:** pela pipeline ou `cd infra/terraform/environments/prod && terraform destroy`. A destruição do EKS (node group + cluster) pode levar vários minutos.

3. **Se o destroy falhar com `Invalid credentials` / `Unauthorized` (em `kubernetes_manifest.argocd_app_*`):** o IAM usado no CI (o usuário das secrets `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY`) não está autorizado no EKS. O Terraform precisa conectar ao cluster para apagar os manifests do Argo CD. **Solução:** inclua o ARN desse IAM em `cluster_access_principal_arns` no `terraform.tfvars` (obtenha com `aws sts get-caller-identity --query Arn --output text` usando as mesmas credenciais do CI), rode **um apply** para criar o EKS Access Entry e, em seguida, rode o destroy de novo. **Alternativa (cluster já inacessível):** remova os recursos do state e destrua o resto: `terraform state rm 'kubernetes_manifest.argocd_app_prod' 'kubernetes_manifest.argocd_app_monitoring'` e depois `terraform destroy`.

4. **Se ainda falhar (subnet / IGW):** o cluster já pode ter sido removido do state, mas a AWS ainda está encerrando recursos (ENIs, etc.). Opções:
   - **Esperar 5–10 minutos** e rodar `terraform destroy` de novo.
   - **Console AWS:** EC2 → Load Balancers → apagar qualquer ALB com nome tipo `k8s-videosys-...`; EC2 → Network Interfaces → encerrar ENIs órfãs da VPC em questão (só se tiver certeza). Depois rodar `terraform destroy` de novo.

(Os itens 3 e 4 acima tratam de erros diferentes: 3 = credenciais EKS no destroy; 4 = dependências de rede após o cluster sumir.)

#### Passo a passo: levantar tudo após `terraform destroy`

Depois de rodar **terraform destroy** (pela pipeline ou local), para ter o projeto rodando de novo:

1. **Pré-requisito**  
   Garantir que existem no GitHub (Settings → Secrets and variables → Actions):
   - **Secrets:** `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` (IAM user para Terraform).
   - **Secret:** `AWS_ROLE_ARN` (será preenchido após o primeiro apply — output `github_actions_role_arn`).
   - **Variáveis:** `AWS_REGION`, `ECR_REGISTRY` (prefixo do ECR, ex.: `123456789012.dkr.ecr.us-east-1.amazonaws.com` — use um item de `ecr_repository_urls` sem o sufixo `/ms-auth`), `terraform_state_bucket`, etc., conforme a seção da Entrega 4.

2. **Terraform apply**  
   Rodar a pipeline **Terraform Apply** (Environment prod) ou, local, `cd infra/terraform/environments/prod && terraform apply`. Isso recria VPC, EKS, ECR, IRSA (ECR), subnet tags, IRSA do Load Balancer Controller e, no mesmo apply, instala via **Helm provider** o AWS Load Balancer Controller, o Argo CD e as duas Applications (video-system-prod e monitoring-stack). O state fica no S3 (bucket configurado no backend). Quem roda o apply precisa ter **AWS CLI** configurado (os providers Kubernetes e Helm usam `aws eks get-token` para autenticar no cluster).

3. **Kubeconfig**  
   Na sua máquina: `aws eks update-kubeconfig --region <region> --name hackathon-prod`. Se `kubectl get nodes` pedir credenciais, incluir seu IAM no `cluster_access_principal_arns` no Terraform e rodar apply de novo. Este é o único passo manual após o apply; o restante (LB controller, Argo CD, Applications) é feito pelo Terraform.

4. **Conferir**  
   - Pods: `kubectl get pods -n video-system` e `kubectl get pods -n monitoring`
   - Ingress (ALB): `kubectl get ingress -n video-system` — o ADDRESS pode levar alguns minutos.
   - Testar: `curl -s http://<alb-hostname>/auth` (e `/video`, `/notify`).
   - Grafana (opcional): `kubectl port-forward svc/monitoring-stack-grafana -n monitoring 3000:80` e acessar `http://localhost:3000` (login admin; senha no Secret ou nos values da Application).

5. **Build-push (opcional)**  
   Se quiser imagens novas no ECR, rodar a pipeline **Build and Push to ECR** (ou push na branch master). O `AWS_ROLE_ARN` no GitHub deve ser o `github_actions_role_arn` do Terraform (configurado após o apply do passo 2).

**Resumo:** Terraform apply (cria EKS e instala LB controller, Argo CD e Applications via Helm/Kubernetes providers) → kubeconfig na sua máquina → Argo CD sincroniza a app e o stack de observabilidade. Não é necessário rodar comandos Helm nem `kubectl apply -f` das Applications manualmente; o Terraform cuida disso.

**Instalação manual (alternativa)**  
Se precisar instalar o LB controller, Argo CD ou as Applications fora do Terraform (ex.: troubleshooting), use os comandos da seção da Entrega 6 (Helm LB controller) e Entrega 7 (Argo CD e Applications) neste README.

---

### Entrega 1 — Fundação local (Docker + Compose)

**Objetivo:** Subir os três serviços (ms-auth, ms-video, ms-notify) e dependências em **Docker Compose** na sua máquina, sem K8s ainda.

- Criar **Dockerfiles** multi-stage para cada ms (Go), com usuário não-root e healthcheck.
- Criar **docker-compose.yml** com ms-auth, ms-video, ms-notify, RabbitMQ e Postgres (um por serviço ou compartilhado para dev).
- Documentar como rodar `docker compose up` e testar um fluxo mínimo (ex.: register, login, upload de um vídeo fake).

**Critério de sucesso:** `docker compose up` sobe tudo; um script ou curl consegue registrar usuário, fazer login e chamar um endpoint de ms-video.

#### Como rodar a Entrega 1

Os três microsserviços usam a imagem **ms-stub** (placeholder em Go) até que os serviços reais existam. RabbitMQ e Postgres sobem como dependências.

1. **Variáveis de ambiente (opcional)**  
   Copie o exemplo e ajuste se quiser:  
   `cp infra/.env.example infra/.env`  
   Os valores padrão (dev/dev/video_system) funcionam para desenvolvimento local. Se usar `infra/.env`, rode a partir de `infra/` com `docker compose up -d --build` para que o arquivo seja carregado, ou use `--env-file infra/.env` ao rodar da raiz.

2. **Subir todos os serviços** (a partir da raiz do repositório):
   ```bash
   docker compose -f infra/docker-compose.yml up -d --build
   ```
   (Ou, a partir de `infra/`: `cd infra && docker compose up -d --build`.)

3. **Verificar** que todos estão em execução e saudáveis:
   ```bash
   docker compose -f infra/docker-compose.yml ps
   ```

4. **Teste do fluxo mínimo (com stubs)**  
   Com os stubs, "registrar usuário / login" é simulado pela resposta do ms-auth em `GET /`; "chamar endpoint de ms-video" é simulado por `GET /` no ms-video. Quando os microsserviços reais existirem, basta trocar a imagem no Compose.

   - ms-auth:   `curl -s http://localhost:8081/`  → esperado: `{"service":"ms-auth"}`
   - ms-video:  `curl -s http://localhost:8082/`  → esperado: `{"service":"ms-video"}`
   - ms-notify: `curl -s http://localhost:8083/`  → esperado: `{"service":"ms-notify"}`

   RabbitMQ Management UI: http://localhost:15672 (usuário/senha padrão: guest/guest).

5. **Smoke test**  
   Para validar o critério de sucesso de forma automatizada:
   ```bash
   ./infra/scripts/smoke-test.sh
   ```

6. **Parar os serviços**:
   ```bash
   docker compose -f infra/docker-compose.yml down
   ```

---

### Entrega 2 — Manifests Kubernetes (Kustomize) local

**Objetivo:** Ter os manifests (Deployments, Services, ConfigMaps) em **Kustomize** e rodar em um cluster **local** (kind ou minikube).

- Criar pasta **infra/k8s/base/** com bases para ms-auth, ms-video, ms-notify e RabbitMQ (e Postgres se quiser no K8s em dev).
- Criar **infra/k8s/overlays/dev/** que usa essas bases (replicas 1, imagens locais ou de um registry acessível).
- Subir cluster com **kind** ou **minikube**; fazer build das imagens e carregar no cluster; aplicar com `kubectl apply -k overlays/dev`.
- Validar: pods Running, serviços acessíveis (port-forward ou Ingress simples).

**Critério de sucesso:** `kubectl get pods` mostra os 3 microsserviços + RabbitMQ (e opcionalmente Postgres) rodando; consegue chamar um endpoint via port-forward.

#### Como rodar a Entrega 2

Os manifests estão em **infra/k8s/** (base + overlay dev). Use um cluster local **kind** ou **minikube** e carregue a imagem **ms-stub:local** no cluster antes de aplicar.

**Opção A — kind**

1. Criar o cluster: `kind create cluster`
2. Build da imagem (na raiz do repo): `docker build -t ms-stub:local packages/ms-stub`
3. Carregar a imagem no cluster: `kind load docker-image ms-stub:local`
4. Aplicar os manifests: `kubectl apply -k infra/k8s/overlays/dev`
5. Aguardar os pods ficarem Ready: `kubectl get pods -w` (Ctrl+C quando todos estiverem Running/Ready)
6. Testar um endpoint (port-forward): `kubectl port-forward svc/ms-auth 8081:8080` e em outro terminal: `curl -s http://localhost:8081/` (esperado: `{"service":"ms-auth"}`)

**Opção B — minikube**

1. Iniciar o cluster: `minikube start`
2. Usar o daemon Docker do minikube e fazer o build: `eval $(minikube docker-env)` e `docker build -t ms-stub:local packages/ms-stub` (ou: `minikube image build -t ms-stub:local packages/ms-stub`)
3. Aplicar os manifests: `kubectl apply -k infra/k8s/overlays/dev`
4. Aguardar os pods: `kubectl get pods -w`
5. Port-forward e teste: `kubectl port-forward svc/ms-auth 8081:8080` e `curl -s http://localhost:8081/`

**Comandos úteis**

- Listar pods: `kubectl get pods`
- Logs de um pod: `kubectl logs -l app=ms-auth -f`
- Remover os recursos: `kubectl delete -k infra/k8s/overlays/dev`
- (kind) Deletar o cluster: `kind delete cluster`

---

### Entrega 3 — Terraform: VPC e EKS

**Objetivo:** Provisionar **VPC** e **cluster EKS** na AWS com Terraform, sem ainda deployar a aplicação.

- Criar **infra/terraform/modules/vpc** (subnets públicas/privadas, NAT, etc.).
- Criar **infra/terraform/modules/eks** (cluster, node group, OIDC para IRSA se for usar depois).
- Usar **infra/terraform/environments/prod** (ou dev para testes opcionais) com esses módulos e provider AWS.
- Configurar **backend remoto** S3 para o state (sem DynamoDB neste escopo).
- Executar `terraform plan` / `terraform apply` e conectar `kubectl` ao EKS (update kubeconfig).

**Critério de sucesso:** EKS criado; `kubectl get nodes` mostra os nodes do cluster. No hackathon o ambiente alvo é **prod**.

#### Como rodar a Entrega 3

**Pré-requisitos:** AWS CLI configurado (`aws configure` ou variáveis de ambiente), Terraform instalado, `kubectl` instalado.

1. **Criar o bucket S3 para o state**  
   O backend usa S3. Crie um bucket antes do primeiro `terraform init`. Neste escopo usamos um bucket **sem versionamento** para reduzir custo; em produção o recomendado é habilitar versionamento no bucket (histórico do state e recuperação).  
   Exemplo (substitua `SEU_BUCKET` por um nome único):
   ```bash
   aws s3 mb s3://SEU_BUCKET --region us-east-1
   ```

2. **Configurar variáveis**  
   Copie o exemplo e ajuste:
   ```bash
   cd infra/terraform/environments/dev
   cp terraform.tfvars.example terraform.tfvars
   # Edite terraform.tfvars com aws_region, availability_zones, cluster_name, etc.
   ```

3. **Inicializar Terraform com o backend**  
   **Recomendado (boa prática):** use um arquivo de config do backend para deixar o comando curto e repetível.
   ```bash
   cp backend.hcl.example backend.hcl
   # Edite backend.hcl e defina bucket (e region se precisar)
   terraform init -backend-config=backend.hcl
   ```
   Alternativa (tudo na linha):  
   `terraform init -backend-config=bucket=SEU_BUCKET -backend-config=key=dev/terraform.tfstate -backend-config=region=us-east-1`

4. **Planejar e aplicar**
   ```bash
   terraform plan
   terraform apply
   ```

5. **Atualizar o kubeconfig**  
   Use o comando exibido no output `kubeconfig_command` ou:
   ```bash
   aws eks update-kubeconfig --region <sua-region> --name <cluster_name>
   ```
   Exemplo: `aws eks update-kubeconfig --region us-east-1 --name hackathon-dev`

6. **Validar**  
   Os nodes do node group devem aparecer como Ready:
   ```bash
   kubectl get nodes
   ```

**Opcional:** Para destruir os recursos e evitar custo quando não estiver usando: `terraform destroy`. O bucket S3 pode ser removido manualmente depois (`aws s3 rb s3://SEU_BUCKET --force`).

---

### Entrega 4 — ECR e primeiro pipeline CI (GitHub Actions)

**Objetivo:** Build e push de **imagens Docker** para **ECR** via **GitHub Actions**.

- Criar repositórios ECR (manual ou Terraform) para ms-auth, ms-video, ms-notify.
- Criar workflow **.github/workflows/build-push.yml**: em push/PR na branch principal, build das 3 imagens (multi-stage), push para ECR. Usar OIDC ou credenciais AWS (recomendado OIDC).
- Garantir que as imagens estão tagadas (ex.: commit SHA ou tag do Git).

**Critério de sucesso:** Push no repo dispara o workflow; imagens aparecem no ECR.

#### Como rodar a Entrega 4

**Pré-requisitos:** Conta AWS configurada, Terraform instalado, repositório no GitHub com o código, permissão para criar Secrets e Variables no repo.

1. **Terraform (apply rápido)**  
   Para criar só ECR e OIDC (sem esperar EKS), comente em [infra/terraform/environments/prod/main.tf](infra/terraform/environments/prod/main.tf) os blocos dos módulos **vpc** e **eks**, e em [outputs.tf](infra/terraform/environments/prod/outputs.tf) comente os outputs **vpc_id**, **cluster_id**, **cluster_endpoint** e **kubeconfig_command**. Depois:
   ```bash
   cd infra/terraform/environments/prod
   cp terraform.tfvars.example terraform.tfvars   # se ainda não tiver
   # Edite terraform.tfvars e defina github_repo (ex.: "sua-org/hackathon")
   terraform init -backend-config=backend.hcl
   terraform plan && terraform apply
   ```
   Anote os outputs **ecr_repository_urls** e **github_actions_role_arn**. Para a Entrega 5, descomente vpc, eks e os outputs e rode `terraform apply` de novo.

2. **GitHub — Secrets e Variables**  
   No repositório (Settings → Secrets and variables → Actions):
   - **Secrets (obrigatórios antes de rodar as pipelines):**
     - `AWS_ACCESS_KEY_ID` e `AWS_SECRET_ACCESS_KEY` = access key de um **IAM user** usado pelas pipelines de Terraform (plan/apply/destroy). Crie o user no console AWS (ex.: nome `github-terraform-ci`), anexe a policy **AdministratorAccess** (ou equivalente) e gere uma access key; guarde o ID e o secret nos dois secrets. Assim o CI consegue rodar destroy e, em seguida, apply sem passo local.
     - `AWS_ROLE_ARN` = output `github_actions_role_arn` (role para build-push no ECR, criada pelo Terraform; configure após o primeiro apply).
     - `TF_STATE_BUCKET` = nome do bucket S3 do state (para Terraform no CI).
   - **Variables:** `AWS_REGION` (ex.: `us-east-1`), `TF_STATE_REGION` (região do bucket), `ECR_REGISTRY` = prefixo do registry ECR (ex.: do output `ecr_repository_urls`, use uma URL e remova o sufixo `/ms-auth` — ex.: `123456789012.dkr.ecr.us-east-1.amazonaws.com`). O workflow build-push concatena `ECR_REGISTRY` + nome do serviço (ms-auth, ms-video, ms-notify).

   **Resumo:** Terraform no CI usa IAM user (access key); build-push usa OIDC (role `AWS_ROLE_ARN`). Não commitar secrets.

3. **Disparar o workflow**  
   Dê push na branch principal (ex.: `master`). O workflow faz build, scan (Trivy) e push para o ECR. Em **pull_request** só rodam build e scan (push não é feito).

4. **Validar**  
   No console AWS (ECR), confira os repositórios ms-auth, ms-video e ms-notify com imagens tagadas pelo **SHA** do commit e por **latest**.

**Rollback:** Em caso de imagem quebrada, use no Kustomize (Entrega 5) a tag do commit anterior (SHA) que já está no ECR.

---

### Entrega 5 — Deploy no EKS (kubectl / Kustomize)

**Objetivo:** Rodar a aplicação no **EKS** usando os manifests Kustomize, com imagens do ECR. Deploy direto em **prod** (sem staging).

- Criar ou ajustar o overlay **infra/k8s/overlays/prod** para apontar para as imagens no ECR (ms-auth, ms-video, ms-notify).
- Configurar **kubeconfig** (local ou no CI) para o cluster EKS prod. O deploy dos manifests em prod é feito pelo **Argo CD** (ver Entrega 7). Para um primeiro deploy antes de instalar o Argo CD, você pode aplicar manualmente: `kubectl apply -k infra/k8s/overlays/prod`.
- Garantir **Secrets** (JWT, DB, RabbitMQ, etc.) no cluster — inicialmente Secrets manuais ou gerados pelo pipeline.

**Critério de sucesso:** Aplicação rodando no EKS prod; endpoints acessíveis via port-forward (Ingress/ALB na Entrega 6).

#### Como rodar a Entrega 5

**Pré-requisitos:** Terraform prod com VPC e EKS já aplicados (descomente os módulos vpc e eks em [main.tf](infra/terraform/environments/prod/main.tf) e rode `terraform apply`); imagens no ECR (workflow build-push já rodou); `kubectl` e AWS CLI instalados.

**Se `kubectl get nodes` der "server has asked for the client to provide credentials":** o seu IAM (user/role) ainda não está autorizado no cluster. Adicione o ARN em [terraform.tfvars](infra/terraform/environments/prod/terraform.tfvars) como `cluster_access_principal_arns = ["arn:aws:iam::ACCOUNT:user/SEU_USER"]` (obtenha com `aws sts get-caller-identity --query Arn --output text`) e rode `terraform apply` (local ou pelo CI). Depois disso, `kubectl get nodes` deve funcionar.

1. **Configurar kubeconfig**  
   Conecte o `kubectl` ao cluster EKS prod (use o nome do cluster e a região do seu Terraform). Suas credenciais AWS precisam estar ativas (`aws sts get-caller-identity` deve funcionar).
   ```bash
   aws eks update-kubeconfig --region us-east-1 --name hackathon-prod
   ```
   Confirme que a conexão funciona antes de aplicar (evita o erro "server has asked for the client to provide credentials"):
   ```bash
   kubectl get nodes
   ```
   Se der erro de credenciais: verifique `AWS_PROFILE` ou `AWS_ACCESS_KEY_ID`/`AWS_SECRET_ACCESS_KEY`; use a mesma conta/região do Terraform que criou o cluster.

2. **Ajustar URLs do ECR no overlay prod**  
   Se a sua conta/região for diferente, edite as imagens nos arquivos:
   - [infra/k8s/overlays/prod/patch-ms-auth-image.yaml](infra/k8s/overlays/prod/patch-ms-auth-image.yaml)
   - [infra/k8s/overlays/prod/patch-ms-video-image.yaml](infra/k8s/overlays/prod/patch-ms-video-image.yaml)
   - [infra/k8s/overlays/prod/patch-ms-notify-image.yaml](infra/k8s/overlays/prod/patch-ms-notify-image.yaml)  
   Use as URLs dos outputs do Terraform (`ecr_repository_urls`). A variável do GitHub para o build-push é `ECR_REGISTRY` (prefixo); a URL completa por serviço é `ECR_REGISTRY/ms-auth`, `ECR_REGISTRY/ms-video`, `ECR_REGISTRY/ms-notify`.

3. **Aplicar os manifests (bootstrap inicial)**  
   Com a **Entrega 7** (Argo CD), o deploy em prod é feito pelo Argo CD a partir do Git; não é necessário rodar este comando após configurar o Argo CD. Se você ainda não instalou o Argo CD, aplique os manifests uma vez na raiz do repositório:
   ```bash
   kubectl apply -k infra/k8s/overlays/prod
   ```

4. **Verificar**  
   Os pods devem ficar Running no namespace `video-system`:
   ```bash
   kubectl get pods -n video-system
   kubectl get svc -n video-system
   ```
   Para testar um endpoint (port-forward no ms-auth):
   ```bash
   kubectl port-forward -n video-system svc/ms-auth 8081:8080
   curl -s http://localhost:8081/
   ```
   Esperado (stub): `{"service":"ms-auth"}`.

5. **Secrets (quando integrar serviços reais)**  
   Os stubs não dependem de Secrets. Quando usar ms-auth/ms-video/ms-notify reais com JWT, Postgres e RabbitMQ, crie os secrets manualmente; veja exemplos em [infra/k8s/overlays/prod/secrets.example.yaml](infra/k8s/overlays/prod/secrets.example.yaml).

**Rollback:** Com Argo CD (Entrega 7), reverta o commit no Git e sincronize a aplicação no Argo CD. Sem Argo CD, reverta o overlay (git revert) e rode `kubectl apply -k infra/k8s/overlays/prod` de novo; ou altere a tag da imagem nos patches e reaplique.

---

### Entrega 6 — Ingress e ALB na AWS

**Objetivo:** Expor os serviços via **Ingress** e **ALB** (AWS Load Balancer Controller).

- Ingress e IngressClass no overlay prod (path-based: `/auth`, `/video`, `/notify`).
- Terraform: tags nas subnets para descoberta do ALB; IRSA (IAM role) para o controller.
- Instalar o **AWS Load Balancer Controller** no EKS via Helm (documentado abaixo).
- DNS (opcional): usar o hostname do ALB em prod ou apontar um domínio Route53 para o ALB.

**Critério de sucesso:** Acesso aos serviços em prod via URL do ALB (ex.: `https://<alb-dns>/auth`).

#### Como rodar a Entrega 6

**Pré-requisitos:** Entrega 5 aplicada (Terraform com VPC + EKS; overlay prod aplicado no cluster); `kubectl` e `helm` instalados; kubeconfig apontando para o cluster prod.

1. **Terraform (tags + IRSA)**  
   O ambiente prod já inclui subnet tags (quando `cluster_name` é passado ao módulo VPC) e a role IRSA para o controller. Rode apply se ainda não aplicou:
   ```bash
   cd infra/terraform/environments/prod
   terraform plan && terraform apply
   ```
   Anote os outputs **lb_controller_role_arn** e **vpc_id** (serão usados no Helm).

2. **Instalar o AWS Load Balancer Controller (Helm)**  
   Adicione o repositório e instale o chart com o nome do cluster, a role IRSA e o **VPC ID**. Informar o VPC ID evita que o controller use o instance metadata (que pode falhar por timeout em redes privadas ou IMDSv2):
   ```bash
   helm repo add eks https://aws.github.io/eks-charts
   helm repo update
   helm upgrade --install aws-load-balancer-controller eks/aws-load-balancer-controller \
     -n kube-system \
     --set clusterName=<CLUSTER_NAME> \
     --set serviceAccount.create=true \
     --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"=<LB_CONTROLLER_ROLE_ARN> \
     --set region=<AWS_REGION> \
     --set vpcId=<VPC_ID>
   ```
   Substitua `<CLUSTER_NAME>` pelo nome do cluster (ex.: `hackathon-prod`), `<LB_CONTROLLER_ROLE_ARN>` pelo output `lb_controller_role_arn`, `<AWS_REGION>` pela região (ex.: `us-east-1`) e `<VPC_ID>` pelo output `vpc_id` do Terraform (ex.: `terraform output -raw vpc_id`). Aguarde os pods do controller ficarem Running: `kubectl get pods -n kube-system -l app.kubernetes.io/name=aws-load-balancer-controller`.

3. **Aplicar o overlay (Ingress + IngressClass)**  
   Com a **Entrega 7** (Argo CD), o overlay (incluindo Ingress) é aplicado pelo Argo CD; não é necessário rodar este comando manualmente. Se você ainda não instalou o Argo CD, aplique uma vez na raiz do repositório:
   ```bash
   kubectl apply -k infra/k8s/overlays/prod
   ```
   O Ingress `video-system` e a IngressClass `alb` serão criados; o controller criará o ALB e os target groups.

4. **Obter a URL do ALB e testar**  
   O endereço do ALB pode levar alguns minutos para aparecer. Liste o Ingress:
   ```bash
   kubectl get ingress -n video-system
   ```
   Use o hostname em ADDRESS (ou no console AWS, em EC2 → Load Balancers). Teste:
   ```bash
   curl -s http://<alb-hostname>/auth
   curl -s http://<alb-hostname>/video
   curl -s http://<alb-hostname>/notify
   ```
   (Se o ALB estiver com listener HTTPS apenas, use `https://` e `-k` se o certificado for inválido.) Esperado com stubs: resposta JSON do serviço (ex.: `{"service":"ms-auth"}`).

5. **DNS (opcional)**  
   Para usar um domínio próprio, crie um registro CNAME no Route53 (ou outro DNS) apontando para o hostname do ALB.

**Rollback:** Com Argo CD, remova o Ingress do overlay no Git e sincronize a aplicação; o controller removerá o ALB. Sem Argo CD, remova do overlay e rode `kubectl apply -k infra/k8s/overlays/prod` de novo. Para desinstalar o controller: `helm uninstall aws-load-balancer-controller -n kube-system`.

---

### Entrega 7 — Argo CD (GitOps)

**Objetivo:** Deploy e atualização da aplicação via **Argo CD**, usando o repositório como fonte da verdade.

- Instalar **Argo CD** no EKS (Helm ou manifest oficial).
- Registrar **Application** apontando para o repositório (pasta **infra/k8s/overlays/prod**). Argo CD usa Kustomize nativamente. Repositório público; branch **master**.
- A partir desta entrega, **não é necessário** rodar `kubectl apply -k infra/k8s/overlays/prod` manualmente para prod: o Argo CD aplica e reconcilia o estado a partir do Git.

**Critério de sucesso:** Alteração em **infra/k8s** (ex.: nova tag de imagem) é refletida no cluster após sync do Argo CD.

#### Como rodar a Entrega 7

**Pré-requisitos:** Entrega 6 concluída (EKS, AWS Load Balancer Controller e overlay prod aplicado ao menos uma vez); `kubectl` e `helm` instalados; kubeconfig apontando para o cluster prod.

1. **Instalar o Argo CD (Helm)**  
   Adicione o repositório e instale o chart no namespace `argocd`:
   ```bash
   helm repo add argo https://argoproj.github.io/argo-helm
   helm repo update
   kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -
   helm upgrade --install argocd argo/argo-cd -n argocd
   ```
   Aguarde os pods ficarem Running: `kubectl get pods -n argocd`.

2. **Senha inicial do admin**  
   A senha do usuário `admin` é gerada e armazenada no Secret `argocd-initial-admin-secret`. Para obter:
   ```bash
   kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d; echo
   ```

3. **Acessar a UI**  
   Por padrão a UI não é exposta publicamente. Use port-forward e acesse `https://localhost:8443` (aceite o certificado autoassinado). Login: `admin` e a senha do passo anterior.
   ```bash
   kubectl port-forward svc/argocd-server -n argocd 8443:443
   ```

4. **Aplicar a Application (overlay prod)**  
   O repositório é **público** e a branch é **master**. Edite [infra/k8s/argocd/application-prod.yaml](infra/k8s/argocd/application-prod.yaml) e substitua `<REPO_URL>` pela URL HTTPS do seu repositório (ex.: `https://github.com/org/hackathon.git`). Depois aplique a Application uma vez:
   ```bash
   kubectl apply -f infra/k8s/argocd/application-prod.yaml
   ```
   No Argo CD (UI ou CLI), a aplicação `video-system-prod` aparecerá; o primeiro sync aplicará os recursos do overlay prod no namespace `video-system`. A Application está configurada com sync automático (prune e selfHeal).

5. **Fluxo de deploy a partir da Entrega 7**  
   O deploy em prod é feito pelo Argo CD a partir do Git. Não é necessário rodar `kubectl apply -k infra/k8s/overlays/prod` manualmente. Para levar uma nova imagem ao cluster: atualize a tag nos patches do overlay prod (ex.: em [patch-ms-auth-image.yaml](infra/k8s/overlays/prod/patch-ms-auth-image.yaml)), faça push no repositório; o Argo CD sincronizará e atualizará os Deployments.

6. **Validar o critério de sucesso**  
   Faça uma alteração de baixo risco em `infra/k8s/overlays/prod` (ex.: um label ou a tag de uma imagem), dê push, aguarde o sync (ou dispare Sync na UI). Verifique no cluster: `kubectl get deployment -n video-system -o wide` e confirme que a mudança foi aplicada.

**Rollback:** Reverta o commit no Git e sincronize a aplicação no Argo CD (ou aguarde o sync automático). Para desinstalar o Argo CD: `helm uninstall argocd -n argocd`.

---

### Entrega 8 — Observabilidade (Prometheus + Grafana)

**Objetivo:** Métricas e dashboards para o cluster e para os microsserviços.

- Instalar **Prometheus** (kube-prometheus-stack / Prometheus Operator) no cluster via Argo CD (Helm).
- Instalar **Grafana** com datasource Prometheus; dashboards pré-instalados pelo chart (CPU/memória dos pods, cluster).
- Expor **métricas** nos serviços Go (ex.: `/metrics` em formato Prometheus); o ms-stub já expõe `/metrics`; ServiceMonitor no overlay prod permite ao Prometheus scrape os três microsserviços.

**Critério de sucesso:** Grafana acessível; dashboard mostrando métricas do cluster e dos serviços.

#### Como rodar a Entrega 8

**Pré-requisitos:** Entrega 7 concluída (Argo CD instalado; Application `video-system-prod` aplicada); overlay prod já contém o ServiceMonitor e os serviços expõem `/metrics`; `kubectl` com kubeconfig apontando para o cluster prod.

1. **Aplicar a Application de monitoring**  
   Uma vez (após o Argo CD estar rodando):
   ```bash
   kubectl apply -f infra/k8s/argocd/application-monitoring.yaml
   ```
   O Argo CD fará o sync e instalará o kube-prometheus-stack (Prometheus, Grafana, node-exporter, kube-state-metrics, etc.) no namespace `monitoring`.

2. **Aguardar sync e verificar pods**  
   Aguarde os pods do stack ficarem Running:
   ```bash
   kubectl get pods -n monitoring
   ```
   O nome do release no namespace pode ser `monitoring-stack` (prefixo do chart). Verifique também que a aplicação `video-system-prod` está em sync para que o ServiceMonitor no overlay prod seja aplicado.

3. **Acesso ao Grafana (port-forward)**  
   Por padrão o Grafana não é exposto por Ingress. Use port-forward (na raiz do repositório ou com o caminho correto para o YAML):
   ```bash
   kubectl port-forward svc/monitoring-stack-grafana -n monitoring 3000:80
   ```
   Se o nome do Service for outro (ex.: `kube-prometheus-stack-grafana`), liste com `kubectl get svc -n monitoring` e use o nome correto. Acesse `http://localhost:3000`. Login: usuário `admin`; senha definida nos values da Application (ex.: `admin` no exemplo; em produção use um Secret ou value mais seguro).

4. **Senha admin do Grafana**  
   Se não tiver a senha, leia o Secret gerado pelo chart:
   ```bash
   kubectl get secret -n monitoring monitoring-stack-grafana -o jsonpath="{.data.admin-password}" | base64 -d; echo
   ```
   (Ajuste o nome do Secret se o release tiver outro nome.)

5. **Dashboards**  
   O chart já instala vários dashboards (Kubernetes / Compute resources / Pods, Workload, etc.). No Grafana: Menu → Dashboards → Browse. Para métricas dos microsserviços (ms-auth, ms-video, ms-notify), verifique em Prometheus → Status → Targets que os targets do job `video-system-services` estão UP; depois use Explore com queries como `up{namespace="video-system"}` ou `process_resident_memory_bytes{namespace="video-system"}`. Opcional: importar o dashboard customizado [infra/k8s/monitoring/grafana-dashboard-video-system.json](infra/k8s/monitoring/grafana-dashboard-video-system.json) (Grafana → Dashboards → New → Import → upload do JSON ou colar o conteúdo) para um painel com status *up* e memória RSS dos serviços do video-system.

6. **Validar o critério de sucesso**  
   Grafana acessível em `http://localhost:3000` (via port-forward); ao menos um dashboard de cluster aberto; targets do video-system em Prometheus com estado UP.

**Rollback:** Remover a Application: `kubectl delete application monitoring-stack -n argocd` (o Argo CD pode remover os recursos do namespace `monitoring` se prune estiver ativo). Ou desabilitar sync e desinstalar o release manualmente: `helm uninstall monitoring-stack -n monitoring`.

---

### Resumo do plano

| # | Entrega | O que você testa |
|---|--------|------------------|
| 1 | Docker + Compose | Tudo sobe local; fluxo auth + video + notify |
| 2 | Kustomize + cluster local | Manifests aplicados; pods rodando |
| 3 | Terraform VPC + EKS | Cluster EKS acessível |
| 4 | ECR + GitHub Actions | Imagens no ECR após push |
| 5 | Deploy no EKS (Kustomize) | App rodando no EKS |
| 6 | Ingress + ALB | Acesso via URL/ALB |
| 7 | Argo CD | GitOps: mudança no Git reflete no cluster |
| 8 | Prometheus + Grafana | Métricas e dashboard |

A partir da **Entrega 1** você já tem um ambiente funcional; as entregas 2–8 vão trazendo Kubernetes, AWS, CI/CD e GitOps de forma incremental e testável. Sempre que quiser, podemos implementar a próxima entrega passo a passo (arquivos, comandos e checagens), usando as skills de Kubernetes, Terraform e Docker que você indicou.
