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

- **Kustomize:** organiza bases e overlays (ex.: `dev`, `staging`, `prod`) para os manifests do EKS (Deployments, Services, Ingress, etc.).
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
| **Manifests K8s** | Kustomize | Bases reutilizáveis + overlays por ambiente (dev/staging/prod) sem duplicar YAML. |
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

### Entrega 3 — Terraform: VPC e EKS (dev)

**Objetivo:** Provisionar **VPC** e **cluster EKS** na AWS com Terraform, sem ainda deployar a aplicação.

- Criar **infra/terraform/modules/vpc** (subnets públicas/privadas, NAT, etc.).
- Criar **infra/terraform/modules/eks** (cluster, node group, OIDC para IRSA se for usar depois).
- Criar **infra/terraform/environments/dev** que usa esses módulos e chama o provider AWS.
- Configurar **backend remoto** (S3 + DynamoDB para lock) para o state.
- Executar `terraform plan` / `terraform apply` e conectar `kubectl` ao EKS (update kubeconfig).

**Critério de sucesso:** EKS criado; `kubectl get nodes` mostra os nodes do cluster em dev.

---

### Entrega 4 — ECR e primeiro pipeline CI (GitHub Actions)

**Objetivo:** Build e push de **imagens Docker** para **ECR** via **GitHub Actions**.

- Criar repositórios ECR (manual ou Terraform) para ms-auth, ms-video, ms-notify.
- Criar workflow **.github/workflows/build-push.yml**: em push/PR na branch principal, build das 3 imagens (multi-stage), push para ECR. Usar OIDC ou credenciais AWS (recomendado OIDC).
- Garantir que as imagens estão tagadas (ex.: commit SHA ou tag do Git).

**Critério de sucesso:** Push no repo dispara o workflow; imagens aparecem no ECR.

---

### Entrega 5 — Deploy no EKS (kubectl / Kustomize)

**Objetivo:** Rodar a aplicação no **EKS** usando os manifests Kustomize, com imagens do ECR.

- Ajustar **overlays** (ex.: criar **overlays/staging** ou usar **dev** com imagens ECR) para apontar para as imagens no ECR.
- Configurar **kubeconfig** no CI (ou usar um runner com acesso ao cluster) e aplicar: `kubectl apply -k infra/k8s/overlays/dev` (ou equivalente).
- Garantir **Secrets** (JWT, DB, RabbitMQ, etc.) no cluster — inicialmente pode ser Secrets manuais ou gerados pelo pipeline.

**Critério de sucesso:** Aplicação rodando no EKS; endpoints acessíveis via ALB ou port-forward.

---

### Entrega 6 — Ingress e ALB na AWS

**Objetivo:** Expor os serviços via **Ingress** e **ALB** (controller AWS Load Balancer ou nginx-ingress).

- Adicionar no Kustomize o **Ingress** (e IngressClass se necessário) para ms-auth, ms-video e ms-notify.
- Instalar e configurar **AWS Load Balancer Controller** (ou outro Ingress controller) no EKS via Terraform ou Helm.
- Configurar DNS (opcional para dev: usar o hostname do ALB).

**Critério de sucesso:** Acesso aos serviços via URL do ALB (ou domínio apontando para o ALB).

---

### Entrega 7 — Argo CD (GitOps)

**Objetivo:** Deploy e atualização da aplicação via **Argo CD**, usando o repositório como fonte da verdade.

- Instalar **Argo CD** no EKS (Helm ou manifest oficial).
- Registrar **Application** apontando para o repositório (pasta **infra/k8s**) e o overlay (ex.: dev/prod). Argo CD usa Kustomize nativamente.
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
