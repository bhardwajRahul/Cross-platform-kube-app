#!/usr/bin/env bash
# Manages local Kind clusters for Luxury Yacht development/testing.
# Usage: ./clusters.sh start | stop

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"
require_commands kind kubectl

rename_context() {
  local kubeconfig="$1"
  local cluster="$2"
  local contexts
  contexts="$(kubectl config get-contexts --kubeconfig "${kubeconfig}" -o name)"
  if grep -Fxq -- "${cluster}" <<< "${contexts}"; then
    kubectl config delete-context "${cluster}" --kubeconfig "${kubeconfig}"
  fi
  kubectl config rename-context \
    --kubeconfig "${kubeconfig}" \
    "kind-${cluster}" "${cluster}"
}

start_clusters() {
  local existing_clusters env cluster kubeconfig
  existing_clusters="$(kind get clusters)"
  mkdir -p "${KUBECONFIG_DIR}"

  for env in dev stg prod; do
    cluster="${env}-cluster"
    kubeconfig="$(kubeconfig_for "${env}")"

    ensure_kind_cluster "${cluster}" "${SCRIPT_DIR}/clusters/${env}.yaml" "${kubeconfig}" "${existing_clusters}"

    rename_context "${kubeconfig}" "${cluster}"

    echo "Kubeconfig written to ${kubeconfig} (context: ${cluster})"

    # Install metrics-server (patched for Kind's self-signed certs)
    echo "Installing metrics-server in '${cluster}'..."
    kubectl_for "${kubeconfig}" "${cluster}" apply \
      -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
    kubectl_for "${kubeconfig}" "${cluster}" patch deployment metrics-server \
      -n kube-system \
      --type=json \
      -p '[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--kubelet-insecure-tls"}]'
  done

  echo ""
  echo "All clusters ready."
  echo "  dev + stg: ${KUBECONFIG_DIR}/dev-stg-clusters"
  echo "  prod:      ${KUBECONFIG_DIR}/prod-clusters"
}

stop_clusters() {
  local existing_clusters env cluster kubeconfig
  existing_clusters="$(kind get clusters)"
  for env in dev stg prod; do
    cluster="${env}-cluster"
    kubeconfig="$(kubeconfig_for "${env}")"
    delete_kind_cluster "${cluster}" "${kubeconfig}" "${existing_clusters}"
  done

  rm -f "${KUBECONFIG_DIR}/dev-stg-clusters" "${KUBECONFIG_DIR}/prod-clusters"
  echo "All clusters stopped."
}

case "${1:-}" in
  start)
    start_clusters
    ;;
  stop)
    stop_clusters
    ;;
  *)
    echo "Usage: $0 {start|stop}" >&2
    exit 1
    ;;
esac
