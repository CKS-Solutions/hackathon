# ms-stub

Serviço HTTP mínimo em Go usado como **placeholder** dos microsserviços (ms-auth, ms-video, ms-notify) para validar a infraestrutura (Docker Compose, Kubernetes) antes dos serviços reais existirem.

## Endpoints

| Endpoint   | Método | Descrição |
|-----------|--------|-----------|
| `/`       | GET    | Retorna JSON `{"service":"<SERVICE_NAME>"}` |
| `/health` | GET    | Liveness probe; retorna 200 e `{"status":"ok"}` |
| `/ready`  | GET    | Readiness probe; retorna 200 |

## Variáveis de ambiente

| Variável        | Default   | Descrição |
|-----------------|-----------|-----------|
| `PORT`          | `8080`    | Porta em que o servidor escuta |
| `SERVICE_NAME`  | `ms-stub` | Nome do serviço (ex.: `ms-auth`, `ms-video`, `ms-notify`) |

## Rodar localmente

```bash
# Com defaults (porta 8080, service name ms-stub)
go run .

# Customizado
PORT=8081 SERVICE_NAME=ms-auth go run .
```

Testar:

```bash
curl -s http://localhost:8080/
curl -s http://localhost:8080/health
curl -s http://localhost:8080/ready
```

## Build da imagem Docker

Na pasta `packages/ms-stub`:

```bash
docker build -t ms-stub .
```

## Rodar como container (uma instância)

```bash
docker run --rm -e SERVICE_NAME=ms-auth -e PORT=8080 -p 8080:8080 ms-stub
```

## Uso como placeholder no Docker Compose / Kubernetes

Use a **mesma imagem** três vezes, variando `SERVICE_NAME` e a porta (ou o port do Service no K8s):

- **ms-auth:**   `SERVICE_NAME=ms-auth`,   porta 8081 (ou 8080 no pod)
- **ms-video:**  `SERVICE_NAME=ms-video`,  porta 8082 (ou 8080 no pod)
- **ms-notify:** `SERVICE_NAME=ms-notify`, porta 8083 (ou 8080 no pod)

Exemplo em Compose (três serviços, mesma imagem):

```yaml
services:
  ms-auth:
    image: ms-stub
    environment:
      SERVICE_NAME: ms-auth
      PORT: "8080"
    ports:
      - "8081:8080"
  ms-video:
    image: ms-stub
    environment:
      SERVICE_NAME: ms-video
      PORT: "8080"
    ports:
      - "8082:8080"
  ms-notify:
    image: ms-stub
    environment:
      SERVICE_NAME: ms-notify
      PORT: "8080"
    ports:
      - "8083:8080"
```

Quando os microsserviços reais estiverem prontos, substitua a imagem `ms-stub` pela imagem de cada serviço (ms-auth, ms-video, ms-notify) e ajuste portas/health checks se necessário.

## Testes

```bash
go test ./...
```

## Build (binário local)

```bash
go build -o ms-stub .
```
