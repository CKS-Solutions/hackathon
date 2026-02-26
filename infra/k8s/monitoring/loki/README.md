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

1. **Cada node só tem os pods que rodam nele.** O `ls /var/log/pods/` que você rodou foi num node onde não há pods do `video-system` (só argocd, kube-system, monitoring). Os ms-auth/ms-video/ms-notify estão em outro node.
   - No Grafana, teste primeiro se **algum** log chega: use `{namespace="monitoring"}` ou `{namespace="argocd"}` (intervalo "Last 1 hour"). Se aparecer, o Loki está recebendo; aí o que falta é ver o node onde está o video-system.
   - Para achar o Promtail que está no mesmo node que o ms-video:  
     `kubectl get pods -n video-system -o wide`  
     `kubectl get pods -n monitoring -l app=promtail -o wide`  
     Escolha o Promtail que está no mesmo **NODE** que um pod do video-system e faça `kubectl exec -n monitoring <promtail-pod> -- ls /var/log/pods/` — aí deve aparecer pasta `video-system_...`.

2. Sync da application `loki-stack` no Argo CD e reinicie o Promtail para carregar a config:
   ```bash
   kubectl rollout restart daemonset/promtail -n monitoring
   ```
3. Aguarde 1–2 minutos e teste no Grafana: `{namespace="video-system"}` ou `{job="kubernetes-pods"}` (intervalo "Last 1 hour").
