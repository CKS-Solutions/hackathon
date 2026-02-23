# MS-Auth - Serviço de Autenticação

Microserviço responsável pela autenticação e autorização de usuários utilizando JWT e PostgreSQL, seguindo arquitetura hexagonal (Ports & Adapters).

## 🏗️ Arquitetura

O serviço fornece endpoints para:

- **Registro de usuários**: Criação de contas com email/senha
- **Login**: Autenticação e geração de token JWT
- **Validação de token**: Verificação de tokens para outros microsserviços

### Stack Tecnológica

- **Linguagem**: Go 1.23
- **Banco de Dados**: PostgreSQL 16
- **Autenticação**: JWT (JSON Web Tokens)
- **Criptografia**: bcrypt para senhas
- **Containerização**: Docker

## 📋 Pré-requisitos

- Docker & Docker Compose instalados
- Go 1.23+ (para desenvolvimento local sem Docker)

## 🚀 Como Rodar o Projeto Localmente

### Opção 1: Usando Docker Compose (Recomendado)

1. **Entre no diretório do ms-auth**:
   ```bash
   cd packages/ms-auth
   ```

2. **Suba os containers**:
   ```bash
   docker-compose up --build
   ```

   Isso iniciará:
   - PostgreSQL (na porta 5432)
   - Aplicação Go (na porta 8080)

3. **Aguarde até ver a mensagem**:
   ```
   ✅ Database connected successfully
   ✅ Database schema initialized
   🚀 Server starting on port 8080
   ```

### Opção 2: Rodando Localmente (Desenvolvimento)

1. **Suba apenas o PostgreSQL**:
   ```bash
   docker-compose up postgres
   ```

2. **Configure as variáveis de ambiente**:
   ```bash
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=postgres
   export DB_PASSWORD=postgres
   export DB_NAME=auth_db
   export DB_SSLMODE=disable
   export JWT_SECRET=your-secret-key-change-in-production
   export JWT_EXPIRATION_HOURS=24
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

**Resposta esperada**:
```json
{"status":"healthy","service":"ms-auth"}
```

### 2. Registrar um novo usuário

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "usuario@exemplo.com",
    "password": "senhaSegura123",
    "name": "João Silva"
  }'
```

**Resposta esperada** (Status 201):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "usuario@exemplo.com",
  "name": "João Silva",
  "created_at": "2026-02-17T12:00:00Z"
}
```

### 3. Fazer login

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "usuario@exemplo.com",
    "password": "senhaSegura123"
  }'
```

**Resposta esperada** (Status 200):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2026-02-18T12:00:00Z",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "usuario@exemplo.com",
    "name": "João Silva",
    "created_at": "2026-02-17T12:00:00Z"
  }
}
```

### 4. Validar um token

```bash
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

**Resposta esperada** (Status 200):
```json
{
  "valid": true,
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "usuario@exemplo.com"
}
```

## 📊 Estrutura do Projeto

```
ms-auth/
├── app/
│   ├── cmd/
│   │   ├── main.go           # Entrypoint da aplicação
│   │   └── http/             # HTTP server
│   ├── internal/
│   │   ├── adapters/
│   │   │   ├── driven/       # Adaptadores de saída
│   │   │   │   ├── jwt/      # Serviço de tokens JWT
│   │   │   │   └── postgres/ # Repositório PostgreSQL
│   │   │   └── driver/       # Adaptadores de entrada
│   │   │       ├── controller/
│   │   │       └── dto/
│   │   ├── core/
│   │   │   ├── entities/     # Entidades de domínio
│   │   │   ├── ports/        # Interfaces (contratos)
│   │   │   └── usecases/     # Lógica de negócio
│   │   └── infra/
│   │       └── database/     # Configuração PostgreSQL
│   └── pkg/
│       └── utils/            # Utilitários
├── docker-compose.yml
├── Dockerfile
└── README.md
```

## 🔧 Variáveis de Ambiente

| Variável | Descrição | Padrão |
|----------|-----------|---------|
| `DB_HOST` | Host do PostgreSQL | `localhost` |
| `DB_PORT` | Porta do PostgreSQL | `5432` |
| `DB_USER` | Usuário do banco | `postgres` |
| `DB_PASSWORD` | Senha do banco | `postgres` |
| `DB_NAME` | Nome do banco | `auth_db` |
| `DB_SSLMODE` | Modo SSL | `disable` |
| `JWT_SECRET` | Chave secreta JWT | ⚠️ **Alterar em produção** |
| `JWT_EXPIRATION_HOURS` | Tempo de expiração do token (horas) | `24` |
| `PORT` | Porta do servidor | `8080` |

## 🔒 Segurança

- **Senhas**: Criptografadas com bcrypt (cost 10)
- **JWT**: Assinados com HS256
- **Tokens**: Incluem tempo de expiração configurável
- **Banco de dados**: Email indexado e único

## 📡 Endpoints

| Método | Endpoint | Descrição | Auth |
|--------|----------|-----------|------|
| GET | `/health` | Health check | Não |
| POST | `/register` | Criar novo usuário | Não |
| POST | `/login` | Autenticar e gerar token | Não |
| POST | `/validate` | Validar token JWT | Não |

## 📝 Modelo de Dados

### Tabela: users

| Campo | Tipo | Descrição |
|-------|------|-----------|
| id | VARCHAR(36) | UUID do usuário (PK) |
| email | VARCHAR(255) | Email único |
| password_hash | VARCHAR(255) | Hash bcrypt da senha |
| name | VARCHAR(255) | Nome do usuário |
| created_at | TIMESTAMP | Data de criação |
| updated_at | TIMESTAMP | Data de atualização |
