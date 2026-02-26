# Loki + Promtail (logs)

- **Loki**: um Deployment no namespace `monitoring`; recebe logs na porta 3100.
- **Promtail**: DaemonSet em cada node; lê os logs dos containers em `/var/log/pods` e envia para o Loki.

Os microsserviços (ms-auth, ms-video, ms-notify) já escrevem em stdout com `log.*` em Go; esses logs são capturados pelo kubelet e lidos pelo Promtail.

## Deploy

1. Garanta que o namespace `monitoring` existe (criado pelo `monitoring-stack` / Prometheus + Grafana).
2. Aplique a Application do Argo CD:  
   `kubectl apply -f infra/k8s/argocd/application-loki.yaml`
3. O data source **Loki** é provisionado via ConfigMap (`grafana-datasource-loki.yaml`). Para o Grafana carregar:
   ```bash
   kubectl rollout restart deployment -l app.kubernetes.io/name=grafana -n monitoring
   ```
   Se o Loki ainda não aparecer no Explore, faça sync da application `loki-stack` no Argo CD e repita o restart.

## Consultar logs no Grafana

- **Explore** → escolha **Loki**.
- Exemplo por namespace: `{namespace="video-system"}`.
- Por app: `{namespace="video-system", app="ms-auth"}`.
- Por pod: `{namespace="video-system", pod=~"ms-auth-.*"}`.
- Intervalo: use "Last 15 minutes" ou "Last 1 hour" para ver logs recentes.

## Se os logs não aparecerem

1. Sync da application `loki-stack` no Argo CD e reinicie o Promtail para carregar a config:
   ```bash
   kubectl rollout restart daemonset/promtail -n monitoring
   ```
2. Aguarde 1–2 minutos e teste no Grafana com `{job="kubernetes-pods"}` (sem filtro de namespace).
3. Confira os targets: `kubectl exec -n monitoring daemonset/promtail -c promtail -- wget -qO- http://127.0.0.1:9080/targets 2>/dev/null | head -100`
