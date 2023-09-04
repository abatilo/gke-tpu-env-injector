#!/usr/bin/env bash
# This script is an incredibly simple integration test for the TPU environment
# variable injector

helm repo add jetstack https://charts.jetstack.io
helm repo update
helm upgrade --install cert-manager jetstack/cert-manager --set installCRDs=true --namespace cert-manager --create-namespace --version v1.12.3
skaffold run

echo "====================================================================="
echo "Testing what happens when no extra environment variables are provided"
echo "====================================================================="
kubectl apply -f ./samples/no_extra_env_vars.yaml
sleep 10

if [ "$(kubectl get pods web-0 -o jsonpath='{.spec.containers[0].env[0].name}')" != "TPU_WORKER_HOSTNAMES" ]; then
  echo "Expected TPU_WORKER_HOSTNAMES to be set, but it was not"
  exit 1
else
  echo "TPU_WORKER_HOSTNAMES was set as expected"
fi

if [ "$(kubectl get pods web-0 -o jsonpath='{.spec.containers[0].env[0].value}')" != "web-0.nginx,web-1.nginx,web-2.nginx" ]; then
  echo "Expected TPU_WORKER_HOSTNAMES to be set to web-0.nginx,web-1.nginx,web-2.nginx, but it was not"
  exit 1
else
  echo "TPU_WORKER_HOSTNAMES was set to expected value"
fi

kubectl delete -f ./samples/no_extra_env_vars.yaml

echo "---"
echo

sleep 5

echo "=================================================================="
echo "Testing what happens when there are existing environment variables"
echo "=================================================================="
kubectl apply -f ./samples/with_extra_env_vars.yaml
sleep 10

if [ "$(kubectl get pods web-0 -o jsonpath='{.spec.containers[0].env[0].name}')" != "FOO" ]; then
  echo "Expected first envvar to be FOO to be set, but it was not"
  exit 1
else
  echo "FOO was the first envvar as expected"
fi

if [ "$(kubectl get pods web-0 -o jsonpath='{.spec.containers[0].env[1].name}')" != "TPU_WORKER_HOSTNAMES" ]; then
  echo "Expected TPU_WORKER_HOSTNAMES to be set, but it was not"
  exit 1
else
  echo "TPU_WORKER_HOSTNAMES was set as expected"
fi

if [ "$(kubectl get pods web-0 -o jsonpath='{.spec.containers[0].env[1].value}')" != "web-0.nginx,web-1.nginx,web-2.nginx" ]; then
  echo "Expected TPU_WORKER_HOSTNAMES to be set to web-0.nginx,web-1.nginx,web-2.nginx, but it was not"
  exit 1
else
  echo "TPU_WORKER_HOSTNAMES was set to expected value"
fi

kubectl delete -f ./samples/with_extra_env_vars.yaml

echo "---"
echo
