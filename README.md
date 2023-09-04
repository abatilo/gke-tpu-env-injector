# gke-tpu-env-injector
[![CI](https://github.com/abatilo/gke-tpu-env-injector/actions/workflows/main.yaml/badge.svg)](https://github.com/abatilo/gke-tpu-env-injector/actions/workflows/main.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/abatilo/gke-tpu-env-injector)](https://goreportcard.com/report/github.com/abatilo/gke-tpu-env-injector)
[![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/abatilo/gke-tpu-env-injector/main/LICENSE)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/abatilo/gke-tpu-env-injector)](https://github.com/abatilo/gke-tpu-env-injector/releases/latest)
![Kubernetes 1.24](https://img.shields.io/badge/Kubernetes-v1.24-green?logo=Kubernetes&style=flat&color=326CE5&logoColor=white)
![Kubernetes 1.25](https://img.shields.io/badge/Kubernetes-v1.25-green?logo=Kubernetes&style=flat&color=326CE5&logoColor=white)
![Kubernetes 1.26](https://img.shields.io/badge/Kubernetes-v1.26-green?logo=Kubernetes&style=flat&color=326CE5&logoColor=white)
![Kubernetes 1.27](https://img.shields.io/badge/Kubernetes-v1.27-green?logo=Kubernetes&style=flat&color=326CE5&logoColor=white)

Automatically inject the environment variables used by libtpu when running TPUs
on GKE.

On [August 31,
2023](https://web.archive.org/web/20230904230811/https://cloud.google.com/blog/products/compute/how-to-use-cloud-tpus-with-gke),
Google officially released support for running their TPU VMs (v4 and v5e) on
Google Kubernetes Engine.

The `tpu_driver` application, with the accompanying TPU device driver
application that gets installed to GKE clusters with TPU support enabled
actually require two interesting environment variables to be available to your
applications when you run them. These environment variables are `TPU_WORKER_ID`
and `TPU_WORKER_HOSTNAMES`.

Taken directly from the [GCP
documentation](https://web.archive.org/web/20230904230344/https://cloud.google.com/kubernetes-engine/docs/how-to/tpus#:~:text=TPU_WORKER_ID%3A%20A%20unique,the%20TPU_WORKER_ID.):

```
TPU_WORKER_ID: A unique integer for each Pod. This ID denotes a unique
worker-id in the TPU slice. The supported values for this field range from zero
to the number of Pods minus one.

TPU_WORKER_HOSTNAMES: A comma-separated list of TPU VM hostnames or IP addresses
that need to communicate with each other within the slice. There should be a
hostname or IP address for each TPU VM in the slice. The list of IP addresses or
hostnames are ordered and zero indexed by the TPU_WORKER_ID.
```

These two environment variables require that you can dynamically inject the
`TPU_WORKER_ID` into each application, and that `TPU_WORKER_HOSTNAMES` contains
individually addressable DNS names for each specific worker, which will
represent the pieces of a TPU PodSlice.

Conveniently, GKE will automatically inject these environment variables into
pods for you BUT they will only do that under [very specific
conditions](https://web.archive.org/web/20230904230344/https://cloud.google.com/kubernetes-engine/docs/how-to/tpus#:~:text=GKE%20automatically%20injects%20these%20environment%20variables%20by%20using%20a%20mutating%20webhook%20when%20a%20Job%20is%20created%20with%20the%20completionMode%3A%20Indexed%2C%20subdomain%2C%20parallelism%20%3E%201%2C%20and%20requesting%20google.com/tpu%20properties.):

```
GKE automatically injects these environment variables by using a mutating
webhook when a Job is created with the completionMode: Indexed, subdomain,
parallelism > 1, and requesting google.com/tpu properties.
```

However, what if you're not launching a Kubernetes `Job` at all? What if you
have your own applications to launch that still need these environment
variables? That's what `gke-tpu-env-injector` is for.

`gke-tpu-env-injector` will do this same environment variable injection for
Kubernetes `StatefulSet`s, which can also leverage a Kubernetes headless
service, in order to get predictable, addressable individual pod DNS addresses.

`gke-tpu-env-injector` does this through the Kubernetes native
[MutatingAdmissionWebhook](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#mutatingadmissionwebhook)
functionality which will intercept all scheduled `StatefulSet`s and `Pod`s that
are annotated with [gke-tpu-env-injector.aaronbatilo.dev/inject:
enabled](https://github.com/abatilo/gke-tpu-env-injector/blob/c64e8bd9c9c60413c2032dea68a8ccb4b3d87138/chart/templates/mutatingwebhook.yaml#L29-L29).

## Getting started

To install `gke-tpu-env-injector`, we've provided a helm chart that's hosted on
the GitHub Container Registry as an OCI artifact:

```
helm upgrade --install gke-tpu-env-injector oci://ghcr.io/abatilo/gke-tpu-env-injector --set cert-manager.enabled=true
```

Setting `cert-manager.enabled=true` will request the required TLS certificates
from `cert-manager` and mount them for `gke-tpu-env-injector` to be able to
receive encrypted webhooks from the Kubernetes control plane.

## Configuration

| CLI flag | Environment variable | Description | Default |
| -------- | -------------------- | ----------- | ------- |
| `--tls-cert-file` | `GTEI_TLS_CERT_FILE` | The path to the file containing the default x509 certificate for HTTPS.                | `/etc/tls/tls.crt`
| `--tls-key-file`  | `GTEI_TLS_KEY_FILE`  | The path to the file containing the default x509 private key matching --tls-cert-file. | `/etc/tls/tls.key`
