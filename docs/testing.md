# Testing

## Kind Cluster Setup

### 1. Create a Kind cluster

```bash
kind create cluster --name dev
```

### 2. Install Cilium

```bash
kubectl apply -k https://github.com/stuttgart-things/helm/infra/crds/cilium

dagger call -m github.com/stuttgart-things/dagger/helm@v0.57.0 \
  helmfile-operation \
  --helmfile-ref "git::https://github.com/stuttgart-things/helm.git@infra/cilium.yaml.gotmpl" \
  --operation apply \
  --state-values "config=kind,clusterName=dev,configureLB=false" \
  --kube-config file:///home/sthings/.kube/dev \
  --progress plain -vv
```

### 3. Install Crossplane

```bash
dagger call -m github.com/stuttgart-things/dagger/helm@v0.57.0 \
  helmfile-operation \
  --helmfile-ref "git::https://github.com/stuttgart-things/helm.git@cicd/crossplane.yaml.gotmpl" \
  --operation apply \
  --state-values "version=2.2.0" \
  --kube-config file:///home/sthings/.kube/dev \
  --progress plain -vv
```

### 4. Install CRDs and run the provider

```bash
kubectl apply -R -f package/crds
go run cmd/provider/main.go --debug
```

## End-to-End Test with Released Provider Package

After the cluster and Crossplane are ready, install the released provider xpkg and test.

### 1. Install the provider

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-clusterbook
spec:
  package: ghcr.io/stuttgart-things/provider-clusterbook-xpkg:latest
```

```bash
kubectl get providers provider-clusterbook
# Wait until INSTALLED=True and HEALTHY=True
```

### 2. Create the ClusterProviderConfig

```bash
kubectl apply -f - <<'EOF'
apiVersion: clusterbook.stuttgart-things.com/v1alpha1
kind: ClusterProviderConfig
metadata:
  name: default
spec:
  url: "http://clusterbook.clusterbook-system:8080"
EOF
```

### 3. Create an IPReservation

```bash
kubectl apply -f - <<'EOF'
apiVersion: ipreservation.clusterbook.stuttgart-things.com/v1alpha1
kind: IPReservation
metadata:
  name: test-reservation
spec:
  forProvider:
    networkKey: "10.31.103"
    clusterName: "test-cluster"
    count: 1
    createDNS: false
  providerConfigRef:
    name: default
EOF
```

### 4. Verify

```bash
# IPReservation should be Ready + Synced
kubectl get ipreservation test-reservation

# Check assigned IPs in status
kubectl get ipreservation test-reservation -o jsonpath='{.status.atProvider.ipAddresses}'
```

## Unit Tests

```bash
go test ./internal/... -v -count=1
```

## Lint

```bash
golangci-lint run ./...
```
