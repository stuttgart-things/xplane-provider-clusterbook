# provider-clusterbook

(ansible-venv) sthings@dev-test1:~/projects$ git clone https://github.com/stuttgart-things/xplane-provider-clusterbook.git
Cloning into 'xplane-provider-clusterbook'...
remote: Enumerating objects: 97, done.
remote: Counting objects: 100% (97/97), done.
remote: Compressing objects: 100% (77/77), done.
remote: Total 97 (delta 16), reused 79 (delta 10), pack-reused 0 (from 0)
Receiving objects: 100% (97/97), 63.09 KiB | 3.32 MiB/s, done.
Resolving deltas: 100% (16/16), done.
(ansible-venv) sthings@dev-test1:~/projects$ cd xplane-provider-clusterbook/
(ansible-venv) sthings@dev-test1:~/projects/xplane-provider-clusterbook$ make submodules
Submodule 'build' (https://github.com/crossplane/build) registered for path 'build'
Cloning into '/home/sthings/projects/xplane-provider-clusterbook/build'...
Submodule path 'build': checked out 'b964dbe0ff0856a762f1a06fe554c647d22af7f0'
(ansible-venv) sthings@dev-test1:~/projects/xplane-provider-clusterbook$ export provider_name=Clusterbook
(ansible-venv) sthings@dev-test1:~/projects/xplane-provider-clusterbook$ make provider.prepare provider=${provider_name}
rm 'apis/sample/sample.go'
rm 'apis/sample/v1alpha1/doc.go'
rm 'apis/sample/v1alpha1/groupversion_info.go'
rm 'apis/sample/v1alpha1/mytype_types.go'
rm 'apis/sample/v1alpha1/zz_generated.deepcopy.go'
rm 'apis/sample/v1alpha1/zz_generated.managed.go'
rm 'apis/sample/v1alpha1/zz_generated.managedlist.go'
rm 'internal/controller/mytype/mytype.go'
rm 'internal/controller/mytype/mytype_test.go'
Removing Makefile.bak
Removing PROVIDER_CHECKLIST.md.bak
Removing README.md.bak
Removing apis/template.go.bak
Removing apis/v1alpha1/doc.go.bak
Removing apis/v1alpha1/register.go.bak
Removing apis/v1alpha1/types.go.bak
Removing cluster/images/provider-template/Dockerfile.bak
Removing cluster/local/integration_tests.sh.bak
Removing cmd/provider/main.go.bak
Removing examples/provider/config.yaml.bak
Removing examples/sample/mytype.yaml.bak
Removing go.mod.bak
Removing internal/controller/config/config.go.bak
Removing internal/controller/register.go.bak
Removing package/crds/sample.template.crossplane.io_mytypes.yaml.bak
Removing package/crds/template.crossplane.io_clusterproviderconfigs.yaml.bak
Removing package/crds/template.crossplane.io_clusterproviderconfigusages.yaml.bak
Removing package/crds/template.crossplane.io_providerconfigs.yaml.bak
Removing package/crds/template.crossplane.io_providerconfigusages.yaml.bak
Removing package/crossplane.yaml.bak









`provider-clusterbook` is a minimal [Crossplane](https://crossplane.io/) Provider
that is meant to be used as a clusterbook for implementing new Providers. It comes
with the following features that are meant to be refactored:

- A `ProviderConfig` type that only points to a credentials `Secret`.
- A `MyType` resource type that serves as an example managed resource.
- A managed resource controller that reconciles `MyType` objects and simply
  prints their configuration in its `Observe` method.

## Developing

1. Use this repository as a clusterbook to create a new one.
1. Run `make submodules` to initialize the "build" Make submodule we use for CI/CD.
1. Rename the provider by running the following command:
```shell
  export provider_name=MyProvider # Camel case, e.g. GitHub
  make provider.prepare provider=${provider_name}
```
4. Add your new type by running the following command:
```shell
  export group=sample # lower case e.g. core, cache, database, storage, etc.
  export type=MyType # Camel casee.g. Bucket, Database, CacheCluster, etc.
  make provider.addtype provider=${provider_name} group=${group} kind=${type}
```
5. Replace the *sample* group with your new group in apis/{provider}.go
5. Replace the *mytype* type with your new type in internal/controller/{provider}.go
5. Replace the default controller and ProviderConfig implementations with your own
5. Register your new type into `SetupGated` function in `internal/controller/register.go`
5. Run `make reviewable` to run code generation, linters, and tests.
5. Run `make build` to build the provider.

Refer to Crossplane's [CONTRIBUTING.md] file for more information on how the
Crossplane community prefers to work. The [Provider Development][provider-dev]
guide may also be of use.

[CONTRIBUTING.md]: https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md
[provider-dev]: https://github.com/crossplane/crossplane/blob/master/contributing/guide-provider-development.md
