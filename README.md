# Kubernetes Time Machine (KTM)

Welcome to KTM, this project aims to allow Devs and AI models replay an entire kubernetes cluster.

See [vision.md](./vision.md) for more details.


# Features

- ## Kubernetes watch

To watch for pods just run `ktm watch` and it will watch for pods and nodes and print them to stdout.

In addition to printing to stdout, it also stores the the pod and node manifests in a local bbolt database.