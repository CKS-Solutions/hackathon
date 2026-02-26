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

## Testar se o Loki está recebendo (Grafana vazio)

Se nem `{namespace="monitoring"}` mostra nada no Grafana, confira:

1. **Loki está up e aceita push?** (use timestamp atual; Loki rejeita logs “too old”)
   ```bash
   kubectl get pods -n monitoring -l app=loki
   kubectl run curl-loki --rm -it --restart=Never --image=curlimages/curl -n monitoring -- \
     sh -c 'TS=$(date +%s)000000000; curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "http://loki:3100/loki/api/v1/push" -H "Content-Type: application/json" -d "{\"streams\":[{\"stream\":{\"test\":\"manual\"},\"values\":[[\""$TS"\",\"teste manual push\"]]}]}"'
   ```
   A última linha deve ser `HTTP_CODE:204`. No Grafana (Explore → Loki), use `{test="manual"}` e intervalo "Last 5 minutes". Se aparecer "teste manual push", Loki e Grafana estão ok e o problema é o Promtail não enviando.

2. **Promtail está enviando ou dando erro?**
   ```bash
   kubectl logs -n monitoring promtail-s67pj --tail=300 2>&1 | grep -iE "error|fail|push|sent"
   ```

3. **Reiniciar Promtail e limpar positions** (para forçar nova leitura dos arquivos):
   ```bash
   kubectl rollout restart daemonset/promtail -n monitoring
   ```
   Aguarde 2 minutos e teste de novo no Grafana: `{namespace="monitoring"}` e `{namespace="video-system"}`, intervalo "Last 15 minutes".

---

## Se os logs ainda não aparecerem

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
