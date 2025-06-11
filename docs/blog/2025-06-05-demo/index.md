---
slug: demo
title: Demo System for Vidra Operator
authors: [slinder, rstutz]
tags: [Vidra, Infrahub, Demo]
---

## 🚀 From Zero to GitOps: Spinning Up the Infrahub + Vidra Demo in Codespaces

Are you curious about how modern infrastructure modeling meets hands-off Kubernetes deployment? The [infrahub-vidra-demo](https://github.com/infrahub-operator/infrahub-vidra-demo) is your playground! In this post, we’ll walk through how you can launch a complete, production-like GitOps environment—right in your browser, with no local setup—using GitHub Codespaces.

### Why This Demo?

Traditional GitOps workflows often require you to handcraft YAML files and wire up complex pipelines. With Infrahub and Vidra, you get a streamlined, user-friendly approach: model your infrastructure visually, let Vidra handle the deployment, and enjoy full traceability—all in a few clicks.

### Step 1: Launch Codespaces and Initialize

Start by opening the demo repository in GitHub Codespaces. Just hit the green “Code” button and select “Open with Codespaces”. Once your environment is ready, initialize everything with:

```bash
task init
```

This single command spins up a local [kind](https://kind.sigs.k8s.io/) Kubernetes cluster and installs Infrahub, Vidra Operator, Vidra CLI, and a self-service frontend. No Docker, no Kubernetes setup—just code.

### Step 2: Check Your Cluster

Verify that Vidra Operator is running:

```bash
kubectl get pod -n vidra-system
```

::: note

> Heads up: Shell completion and the `k` alias might not work in Codespaces, so use the full `kubectl` command.

:::

### Step 3: Explore the UI

- Open the **Ports** tab in Codespaces.
- Click the globe icon for port `8000` (Infrahub) and `5001` (Frontend).
- In the frontend, submit a webserver request. This action creates a proposed change in Infrahub.
- Log in to Infrahub (`admin` / `infrahub`), review, and merge the change. Artifacts are generated and ready for deployment.

### Step 4: Sync with Vidra

Bring your modeled infrastructure to life in the cluster:

```bash
host_ip=$(hostname -I | awk '{print $1}')
vidra-cli credentials apply https://${host_ip} --username admin --password infrahub
vidra-cli infrahubsync apply "http://${host_ip}:8000" -a Webserver_Manifest -b main -N default -e
```

If you’re running locally (not in Codespaces), you can use:

```bash
task vidra-add-sync
```

### Step 5: Dive Deeper

Want to see what Vidra is managing? Try:

```bash
kubectl get infrahubsyncs.infrahub.operators.com -o yaml
kubectl get vidraresources.infrahub.operators.com
kubectl get vidraresources.infrahub.operators.com -o yaml
kubectl get <kind> <name> -n <namespace>
```

Or launch [k9s](https://k9scli.io/) for a terminal UI:

```bash
k9s
```

List all available tasks with:

```bash
task
```

---

### Final Thoughts

This demo isn’t just a quickstart—it’s a glimpse into the future of infrastructure management. Model your systems in Infrahub, let Vidra automate the deployment, and manage everything through intuitive UIs. Whether you’re a platform engineer or just exploring GitOps, this Codespaces-powered demo is the fastest way to experience the workflow.

**Ready to try it? Jump in, experiment, and see how easy GitOps can be!**


