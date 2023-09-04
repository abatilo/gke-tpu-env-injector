# gke-tpu-env-injector
Automatically inject the environment variables used by libtpu when running TPUs on GKE

## Configuration

| CLI flag | Environment variable | Description | Default |
| -------- | -------------------- | ----------- | ------- |
| `--tls-cert-file` | `GTEI_TLS_CERT_FILE` | The path to the file containing the default x509 certificate for HTTPS.                | `/etc/tls/tls.crt`
| `--tls-key-file`  | `GTEI_TLS_KEY_FILE`  | The path to the file containing the default x509 private key matching --tls-cert-file. | `/etc/tls/tls.key`
