# Kubernetes Time Machine (KTM)

Welcome to KTM, this project aims to allow Devs and AI models replay an entire kubernetes cluster.

See [vision.md](./vision.md) for more details.


# Features

- ## Kubernetes PodWatch

To watch for pods just run `ktm podwatch` and it will watch for pods and print them to stdout.

In addition to printing to stdout, it also stores the pod manifest in a local bbolt database.