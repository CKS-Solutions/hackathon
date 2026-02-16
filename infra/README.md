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
   - **Secrets:**  
     - `AWS_ROLE_ARN` = output `github_actions_role_arn` (role só para build-push no ECR).  
     - `TF_AWS_ROLE_ARN` = output `github_actions_terraform_role_arn` (role para os workflows Terraform plan/apply/destroy).  
     - `TF_STATE_BUCKET` = nome do bucket S3 do state (para Terraform no CI).
   - **Variables:** `AWS_REGION` (ex.: `us-east-1`), `TF_STATE_REGION` (região do bucket), `ECR_MS_AUTH_URL`, `ECR_MS_VIDEO_URL`, `ECR_MS_NOTIFY_URL` = URLs do output `ecr_repository_urls`. Não commitar secrets; OIDC evita chaves de longa duração.

   **Nota:** As roles de OIDC são duas: uma para build-push (ECR) e outra para Terraform. A primeira vez que você aplicar o Terraform que cria a role `github-actions-terraform`, rode `terraform apply` **localmente** (credenciais com permissão para criar IAM roles); depois configure o secret `TF_AWS_ROLE_ARN` com o output `github_actions_terraform_role_arn`. A partir daí o CI usa essa role e não é preciso ficar adicionando permissões no `github-oidc.tf` (a role de Terraform usa a managed policy PowerUserAccess).

3. **Disparar o workflow**  
   Dê push na branch principal (ex.: `master`). O workflow faz build, scan (Trivy) e push para o ECR. Em **pull_request** só rodam build e scan (push não é feito).

4. **Validar**  
   No console AWS (ECR), confira os repositórios ms-auth, ms-video e ms-notify com imagens tagadas pelo **SHA** do commit e por **latest**.

**Rollback:** Em caso de imagem quebrada, use no Kustomize (Entrega 5) a tag do commit anterior (SHA) que já está no ECR.

---

### Entrega 5 — Deploy no EKS (kubectl / Kustomize)

**Objetivo:** Rodar a aplicação no **EKS** usando os manifests Kustomize, com imagens do ECR. Deploy direto em **prod** (sem staging).

- Criar ou ajustar o overlay **infra/k8s/overlays/prod** para apontar para as imagens no ECR (ms-auth, ms-video, ms-notify).
- Configurar **kubeconfig** (local ou no CI) para o cluster EKS prod e aplicar: `kubectl apply -k infra/k8s/overlays/prod`.
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
   Use as URLs dos outputs do Terraform (`ecr_repository_urls`) ou as variáveis do GitHub (ECR_MS_AUTH_URL, etc.).

3. **Aplicar os manifests**  
   Na raiz do repositório:
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

**Rollback:** Reverter o overlay (git revert) e rodar `kubectl apply -k infra/k8s/overlays/prod` de novo; ou alterar a tag da imagem nos patches (ex.: `:latest` → `:<sha-anterior>`) e reaplicar.

---

### Entrega 6 — Ingress e ALB na AWS

**Objetivo:** Expor os serviços via **Ingress** e **ALB** (controller AWS Load Balancer ou nginx-ingress).

- Adicionar no Kustomize o **Ingress** (e IngressClass se necessário) para ms-auth, ms-video e ms-notify.
- Instalar e configurar **AWS Load Balancer Controller** (ou outro Ingress controller) no EKS via Terraform ou Helm.
- Configurar DNS (opcional: usar o hostname do ALB em prod).

**Critério de sucesso:** Acesso aos serviços em prod via URL do ALB (ou domínio apontando para o ALB).

---

### Entrega 7 — Argo CD (GitOps)

**Objetivo:** Deploy e atualização da aplicação via **Argo CD**, usando o repositório como fonte da verdade.

- Instalar **Argo CD** no EKS (Helm ou manifest oficial).
- Registrar **Application** apontando para o repositório (pasta **infra/k8s**) e o overlay **prod**. Argo CD usa Kustomize nativamente.
- Desligar o apply manual do pipeline para os manifests (o pipeline só builda e faz push das imagens; Argo CD aplica os manifests). Ou manter o pipeline apenas atualizando a tag da imagem no Kustomize (image updater) e deixar o Argo CD reconciliar.

**Critério de sucesso:** Alteração em **infra/k8s** (ex.: nova tag de imagem) é refletida no cluster após sync do Argo CD.

---

### Entrega 8 — Observabilidade (Prometheus + Grafana)

**Objetivo:** Métricas e dashboards para o cluster e para os microsserviços.

- Instalar **Prometheus** (e talvez **kube-prometheus-stack** ou Prometheus Operator) no cluster.
- Instalar **Grafana** e configurar datasource Prometheus; criar um dashboard básico (CPU/memória dos pods, requests por serviço se as apps expuserem métricas).
- Expor métricas nos serviços Go (ex.: `/metrics` em formato Prometheus).

**Critério de sucesso:** Grafana acessível; dashboard mostrando métricas do cluster e dos serviços.

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
