# Migration from v2 to v3

Make sure you understand the [differences between Kubebuilder v2 and v3](./v2vsv3.md)
before continuing.

Please ensure you have followed the [installation guide](/quick-start.md#installation)
to install the required components.

The recommended way to migrate a v2 project is to create a new v3 project and
copy over the API and the reconciliation code. The conversion will end up with a
project that looks like a native v3 project. However, in some cases, it's
possible to do an in-place upgrade (i.e. reuse the v2 project layout, upgrading
controller-runtime and controller-tools.  

## Initialize a v3 Project

Create a new directory with the name of your project. Note that
this name is used in the scaffolds to create the name of your manager Pod and of the Namespace where the Manager is deployed by default.  

```bash
$ mkdir <my-project-name>
$ cd my-project-name
```

Now, we need to initialize a v3 project.  Before we do that, though, we'll need
to initialize a new go module if we're not on the `GOPATH`.

```bash
go mod init tutorial.kubebuilder.io/project
```

<aside class="note warning">
<h1>The module of your project can found in the `go.mod`:</h1>

```
module tutorial.kubebuilder.io/project
```

</aside>

Then, we can finish initializing the project with kubebuilder.

```bash
kubebuilder init --domain tutorial.kubebuilder.io
```

<aside class="note warning">
<h1>The domain of your project can be found in the PROJECT file:</h1>

```
...
domain: tutorial.kubebuilder.io
...
```
</aside>

## Migrate APIs and Controllers

Next, we'll re-scaffold out the API types and controllers. Since we want both,
we'll say yes to both the API and controller prompts when asked what parts we
want to scaffold:

```bash
kubebuilder create api --group batch --version v1 --kind CronJob
```

<aside class="note warning">
<h1>How to still using `apiextensions.k8s.io/v1beta1` for CRDs?</h1>

From now on, the CRDs that will be created by the tool will be using the Kubenetes API version `apiextensions.k8s.io/v1`  by default, instead of `apiextensions.k8s.io/v1beta1` . So, if you would like to keep using the previous version, use the flag `--crd-version=v1beta` in the above command.  

</aside>

<aside class="note warning">
<h1>If you're using multiple groups</h1>

Please run `kubebuilder edit --multigroup=true` to enable multi-group support before migrating the APIs and controllers. Please see [this](/migration/multi-group.md) for more details.

</aside>

### Migrate the APIs

Now, let's copy the API definition from `api/v1/cronjob_types.go` our old project to the new one. 

### Migrate the Controllers

Now, let's copy the controller code from `controllers/cronjob_controller.go` our old project to the new one and fix the following breaking change:  
 
Replace:

```go 
func (r *CronJobReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
 	ctx := context.Background() 
 	log := r.Log.WithValues("cronjob", req.NamespacedName)
```

With:

```go 
func (r *CronJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("cronjob", req.NamespacedName)
```

<aside class="note warning">
<h1>Controller-runtime version updated has breaking changes</h1>

Check [sigs.k8s.io/controller-runtime release docs from 0.7.0+ version]( https://github.com/kubernetes-sigs/controller-runtime/releases ) for breaking changes.

</aside>

## Migrate the Webhooks

<aside class="note warning">
<h1>Note</h1>

If you don't have a webhook, you can skip this section.

</aside>

Now let's scaffold the webhooks for our CRD (CronJob). We'll need to run the
following command with the `--defaulting` and `--programmatic-validation` flags
(since our test project uses defaulting and validating webhooks):

```bash
kubebuilder create webhook --group batch --version v1 --kind CronJob --defaulting --programmatic-validation
```

<aside class="note warning">
<h1>How to still using `apiextensions.k8s.io/v1beta1` for Webhooks?</h1>

From now on, the Webhooks that will be created by the tool will be using by default the Kubernetes API version `apiextensions.k8s.io/v1` instead of `apiextensions.k8s.io/v1beta1`,  `admissionregistration.k8s.io/v1` instead of `admissionregistration.k8s.io/v1beta1` and the `cert-manager.io/v1` to replace `cert-manager.io/v1alpha2`. So, if you would like to keep using the previous version use the flag `--webhook-version=v1beta` in the above command.  
</aside>

Now, let's copy the API definition from `api/v1/cronjob_webhook.go` our old project to the new one. 

## Verification

Finally, we can run `make` and `make docker-build` to ensure things are working
fine.
