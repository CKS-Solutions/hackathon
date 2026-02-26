# Loki + Promtail (logs)

- **Loki**: um Deployment no namespace `monitoring`; recebe logs na porta 3100.
- **Promtail**: DaemonSet em cada node; lê os logs dos containers em `/var/log/pods` e envia para o Loki.

Os microsserviços (ms-auth, ms-video, ms-notify) já escrevem em stdout com `log.*` em Go; esses logs são capturados pelo kubelet e lidos pelo Promtail.

## Deploy

1. Garanta que o namespace `monitoring` existe (criado pelo `monitoring-stack` / Prometheus + Grafana).
2. Aplique a Application do Argo CD:  
   `kubectl apply -f infra/k8s/argocd/application-loki.yaml`
3. No **Grafana**, adicione o data source **Loki**:
   - URL: `http://loki.monitoring.svc.cluster.local:3100` (ou `http://loki:3100` se o Grafana estiver no mesmo namespace).
   - Save & Test.

## Consultar logs no Grafana

- **Explore** → escolha **Loki**.
- Exemplo por namespace: `{namespace="video-system"}`.
- Por app: `{namespace="video-system", app="ms-auth"}`.
- Por pod: `{namespace="video-system", pod=~"ms-auth-.*"}`.
