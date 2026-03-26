# Provider Development

Scaffolding a new Crossplane provider from the template. Follow the steps exactly in order.

## Prerequisites

### angryjet

```bash
go install github.com/crossplane/crossplane-tools/cmd/angryjet@latest
```

## Init

### 1. Create the repo from the template

Go to [crossplane/provider-template](https://github.com/crossplane/provider-template) → click **Use this template** → create your repo.

```bash
git clone https://github.com/YOUR_ORG/provider-YOUR_SYSTEM
cd provider-YOUR_SYSTEM
```

### 2. Init the build submodule

```bash
make submodules
```

### 3. Rename the provider

```bash
export provider_name=Clusterbook
make provider.prepare provider=${provider_name}
```

### 4. Add your first type

```bash
export group=ipreservation
export type=IPReservation
make provider.addtype provider=Clusterbook group=${group} kind=${type}
```

### 5. Fix API group domain

The template defaults to `clusterbook.crossplane.io` and the `addtype` target introduces a double-prefix. Replace both with your org domain:

```bash
# Replace base domain
find . -type f \( -name "*.go" -o -name "*.yaml" \) \
  | xargs grep -l "clusterbook.crossplane.io" \
  | xargs sed -i 's|clusterbook.crossplane.io|clusterbook.stuttgart-things.com|g'

# Fix double-prefix introduced by addtype
find . -type f \( -name "*.go" -o -name "*.yaml" \) \
  | xargs grep -l "ipreservation.ipreservation" \
  | xargs sed -i 's|ipreservation.ipreservation|ipreservation|g'
```

Verify no stale references remain:

```bash
grep -r "crossplane.io" apis/ --include="*.go" | grep -v zz_generated
# expected: no output
```
