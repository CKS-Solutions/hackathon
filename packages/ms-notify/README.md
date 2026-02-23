# MS-Notify - Serviço de Notificações

Microserviço responsável pelo envio de notificações por email utilizando arquitetura hexagonal (Ports & Adapters).

## 🏗️ Arquitetura

O serviço é composto por dois componentes principais:

- **HTTP Server** (porta 8080): Recebe requisições para envio de notificações e publica na fila SQS
- **SQS Consumer**: Processa mensagens da fila e envia emails via AWS SES

### Stack Tecnológica

- **Linguagem**: Go 1.23
- **AWS Services**: SQS, DynamoDB, SES
- **Infraestrutura Local**: LocalStack
- **Containerização**: Docker

## 📋 Pré-requisitos

- Docker & Docker Compose instalados
- Go 1.23+ (para desenvolvimento local sem Docker)

## 🚀 Como Rodar o Projeto Localmente

### Opção 1: Usando Docker Compose (Recomendado)

1. **Entre no diretório do ms-notify**:
   ```bash
   cd packages/ms-notify
   ```

2. **Suba os containers**:
   ```bash
   docker-compose up --build
   ```

   Isso iniciará:
   - LocalStack (simulando AWS na porta 4566)
   - Aplicação Go (na porta 8080)

3. **Aguarde até ver a mensagem**:
   ```
   ✅ Recursos AWS criados com sucesso!
   ```

### Opção 2: Rodando Localmente (Desenvolvimento)

1. **Suba apenas o LocalStack**:
   ```bash
   docker-compose up localstack
   ```

2. **Configure as variáveis de ambiente**:
   ```bash
   export AWS_REGION=us-east-1
   export AWS_ACCESS_KEY_ID=test
   export AWS_SECRET_ACCESS_KEY=test
   export AWS_ENDPOINT_URL=http://localhost:4566
   export AWS_STAGE=local
   ```

3. **Entre no diretório da aplicação**:
   ```bash
   cd app
   ```

4. **Baixe as dependências**:
   ```bash
   go mod download
   ```

5. **Execute a aplicação**:
   ```bash
   go run cmd/main.go
   ```

## 🧪 Testando o Serviço

### 1. Verificar se o serviço está rodando

```bash
curl http://localhost:8080/health
```

### 2. Enviar uma notificação

```bash
curl -X POST http://localhost:8080/notification \
  -H "Content-Type: application/json" \
  -d '{
    "subject": "Teste de Notificação",
    "to": ["destinatario@exemplo.com"],
    "html": "<h1>Olá!</h1><p>Esta é uma mensagem de teste.</p>"
  }'
```

**Resposta esperada**:
```json
{"message":"notification request accepted"}
```

### 3. Verificar logs do consumer

Os logs mostrarão o processamento das mensagens:
```
[QUEUE_READ] Message received
[USE_CASE_ERR] ou sucesso no envio
```

### 4. Consultar notificações no DynamoDB (LocalStack)

```bash
docker exec localstack awslocal dynamodb scan \
  --table-name "MSNotify.Notification" \
  --region us-east-1
```

## 📊 Estrutura do Projeto

```
ms-notify/
├── app/
│   ├── cmd/
│   │   ├── main.go           # Entrypoint da aplicação
│   │   ├── http/             # HTTP server
│   │   └── sqs/              # SQS consumer
│   ├── internal/
│   │   ├── adapters/
│   │   │   ├── driven/       # Adaptadores de saída (AWS)
│   │   │   └── driver/       # Adaptadores de entrada (Controllers)
│   │   ├── core/
│   │   │   ├── entities/     # Entidades de domínio
│   │   │   ├── ports/        # Interfaces (contratos)
│   │   │   └── usecases/     # Lógica de negócio
│   │   └── infra/
│   │       └── aws/          # Configuração AWS
│   └── pkg/
│       └── utils/            # Utilitários
├── localstack/
│   └── init.sh               # Script de inicialização AWS
├── docker-compose.yml
├── Dockerfile
└── README.md
```

## 🔧 Variáveis de Ambiente

| Variável | Descrição | Padrão |
|----------|-----------|---------|
| `AWS_REGION` | Região AWS | `us-east-1` |
| `AWS_STAGE` | Ambiente (local/api) | `local` |
| `AWS_ACCESS_KEY_ID` | Credencial AWS | `test` (local) |
| `AWS_SECRET_ACCESS_KEY` | Credencial AWS | `test` (local) |
| `AWS_ENDPOINT_URL` | URL do LocalStack | `http://localstack:4566` |

## 🐛 Troubleshooting

### Container não inicia

```bash
# Limpe containers antigos
docker-compose down -v

# Reconstrua as imagens
docker-compose up --build
```

### Erro de conexão com LocalStack

Verifique se o LocalStack está rodando:
```bash
docker ps | grep localstack
```

### Mensagens não são processadas

Verifique se a fila foi criada:
```bash
docker exec localstack awslocal sqs list-queues --region us-east-1
```

## 🏗️ Infraestrutura (produção)

A IaC para SQS, DynamoDB e SES está em **`infra/terraform`** (não mais em `packages/ms-notify/terraform`):

- **Módulos genéricos:** `infra/terraform/modules/sqs-queue`, `dynamodb-table`, `ses-email-identity`
- **Prod:** `infra/terraform/environments/prod/ms-notify.tf` instancia os três módulos para o ms-notify

**Variáveis de ambiente em produção:** obtenha os valores com `terraform output` em `infra/terraform/environments/prod`:

- `ms_notify_sqs_queue_url` → env no Deployment (ex.: `SQS_QUEUE_URL`)
- `ms_notify_dynamodb_table_name` → env no Deployment
- `ms_notify_ses_sender_email` → env no Deployment

Injeção pode ser manual (ConfigMap/Secret preenchidos a partir dos outputs) ou, futuramente, External Secrets. Para **desenvolvimento local**, use LocalStack e o script `packages/ms-notify/localstack/init.sh` para criar fila, tabela e identidade SES no LocalStack.

## 📝 Notas de Desenvolvimento

- O serviço usa **LocalStack** para simular AWS localmente
- Emails não são realmente enviados localmente (SES mock)
- Notificações são persistidas no DynamoDB mesmo em caso de falha no envio
- O consumer SQS só deleta mensagens após processamento bem-sucedido
