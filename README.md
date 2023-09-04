# gke-tpu-env-injector
[![CI](https://github.com/abatilo/gke-tpu-env-injector/actions/workflows/main.yaml/badge.svg)](https://github.com/abatilo/gke-tpu-env-injector/actions/workflows/main.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/abatilo/gke-tpu-env-injector)](https://goreportcard.com/report/github.com/abatilo/gke-tpu-env-injector)
[![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/abatilo/gke-tpu-env-injector/main/LICENSE)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/abatilo/gke-tpu-env-injector)](https://github.com/abatilo/gke-tpu-env-injector/releases/latest)
![Kubernetes 1.24](https://img.shields.io/badge/Kubernetes-v1.24-green?logo=Kubernetes&style=flat&color=326CE5&logoColor=white)
![Kubernetes 1.25](https://img.shields.io/badge/Kubernetes-v1.25-green?logo=Kubernetes&style=flat&color=326CE5&logoColor=white)
![Kubernetes 1.26](https://img.shields.io/badge/Kubernetes-v1.26-green?logo=Kubernetes&style=flat&color=326CE5&logoColor=white)
![Kubernetes 1.27](https://img.shields.io/badge/Kubernetes-v1.27-green?logo=Kubernetes&style=flat&color=326CE5&logoColor=white)

Automatically inject the environment variables used by libtpu when running TPUs on GKE

## Configuration

| CLI flag | Environment variable | Description | Default |
| -------- | -------------------- | ----------- | ------- |
| `--tls-cert-file` | `GTEI_TLS_CERT_FILE` | The path to the file containing the default x509 certificate for HTTPS.                | `/etc/tls/tls.crt`
| `--tls-key-file`  | `GTEI_TLS_KEY_FILE`  | The path to the file containing the default x509 private key matching --tls-cert-file. | `/etc/tls/tls.key`
