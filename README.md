# Installing Prometheus and Grafana with Argo CD

This repository contains two Argo CD Application manifests that deploy Prometheus (the kube-prometheus-stack) and Grafana (separate Helm chart) to a Kubernetes cluster using Argo CD.

Files included:

- `prometheus.yml` — Argo CD Application that installs the kube-prometheus-stack Helm chart from the prometheus-community repository with Grafana disabled in the stack.
- `grafana.yml` — Argo CD Application that installs Grafana from the official Grafana Helm repository.

---

## Overview

These Argo CD `Application` resources deploy the Helm charts directly from the upstream Helm chart repositories. The pattern used here keeps Prometheus and Grafana separate so you can manage Grafana configuration independently and customize admin credentials, service type, persistence, and dashboards.

> NOTE: The `prometheus.yml` file disables Grafana in the kube-prometheus-stack (grafana.enabled: false). Grafana is deployed separately by `grafana.yml`.

---

## Prerequisites

- A Kubernetes cluster (Minikube, kind, EKS, AKS, GKE, etc.)
- kubectl configured to talk to your cluster
- Argo CD installed in your cluster (the `argocd` namespace must exist and the `argocd` server accessible)
  - Follow https://argo-cd.readthedocs.io/en/stable/getting_started/ to install Argo CD if you don't have it already.

Optional but useful:
- `argocd` CLI configured (argocd login) — makes querying and syncing easier
- A LoadBalancer-capable environment if you plan to expose Grafana via a LoadBalancer service

---

## Quick installation steps

1. Make sure `argocd` namespace exists and Argo CD is running.

  ```bash
  kubectl get ns argocd || kubectl create ns argocd
  kubectl -n argocd get pods
  ```

2. Create the `monitoring` namespace (destinations in the examples use `monitoring`).

  ```bash
  kubectl get ns monitoring || kubectl create ns monitoring
  ```

3. Apply the Argo CD Application resources provided in this repo (these create Application CRs in the `argocd` namespace):

  ```bash
  kubectl apply -f prometheus.yml
  kubectl apply -f grafana.yml
  ```

   These `Application` resources cause Argo CD to sync the selected upstream Helm charts into the `monitoring` namespace.

4. Verify from Argo CD that both applications are created and healthy.

   Using the `argocd` CLI (if configured):

  ```bash
  argocd app list
  argocd app get prometheus
  argocd app get grafana
  ```

   Or with kubectl:

  ```bash
  kubectl -n argocd get applications
  kubectl -n argocd describe application prometheus
  kubectl -n argocd describe application grafana
  ```

---

## Accessing Grafana

The `grafana.yml` chart in this repo configures the service type as `LoadBalancer`. How you access Grafana depends on your cluster environment:

- If your environment provides a LoadBalancer (cloud provider), get the external IP:

  ```bash
  kubectl -n monitoring get svc
  # use the EXTERNAL-IP / ADDRESS for the grafana service
  ```

- If you don't have a LoadBalancer (e.g., kind/minikube), use port-forwarding to access Grafana locally:

  ```bash
  # Option 1 — find a grafana pod and port-forward the pod or deployment
  kubectl -n monitoring get pods | grep -i grafana
  kubectl -n monitoring port-forward deployment/grafana 3000:3000
  # open http://localhost:3000

  # Option 2 — port-forward the service (works if the service exposes port 3000)
  kubectl -n monitoring port-forward svc/grafana 3000:3000
  ```

Default credentials (as configured in `grafana.yml` for example purposes):

- Username: `admin`
- Password: `admin1234` (change this in production or use a secret)

---

## Prometheus verification

Check Prometheus pods and services in the `monitoring` namespace:

```bash
kubectl -n monitoring get pods
kubectl -n monitoring get svc
```

Prometheus and related components (node-exporter, kube-state-metrics) should show up when the Helm chart is installed.

You can port-forward Prometheus or the Prometheus Server service to inspect the UI:

```bash
# find the Prometheus server deployment/pod name, then port-forward
kubectl -n monitoring get pods -l app.kubernetes.io/name=prometheus
kubectl -n monitoring port-forward deployment/prometheus-server 9090:9090
# open http://localhost:9090
```

---

## Customization and security

- To change Helm values (e.g., admin password, persistence, ingress), edit the respective `values` block in `prometheus.yml` and `grafana.yml` or use a values file referenced by the Application.
- Never store plaintext production credentials in YAML. Use secrets (Sealed Secrets, External Secrets) or integrate with a secret management solution.
- Adjust `targetRevision` in the Application manifests to lock to specific chart versions.

---

## Cleanup

To remove everything created by these Application manifests:

```bash
kubectl -n argocd delete application prometheus
kubectl -n argocd delete application grafana
kubectl delete ns monitoring
```

If you used `kubectl apply -f` in this repo and want to fully cleanup Argo CD objects, delete the Application resources from `argocd` namespace as shown above.

---

## Notes & next steps

- This repo shows an example pattern where Prometheus (kube-prometheus-stack) and Grafana are managed separately by Argo CD.
- For production setups, consider adopting:
  - A GitOps pattern where these Application manifests are stored and managed in Git branches and promoted via Argo CD Projects
  - Secure secrets management (do not commit passwords to Git)
  - Using Argo CD App of Apps pattern for large-scale organization


If you'd like, I can also:
- Harden these manifests (move sensitive data to Kubernetes Secrets / Sealed Secrets)
- Add README instructions for authenticating to the Argo CD UI/CLI
- Add sample values files for Grafana dashboards and datasources

---

Happy monitoring! :)