---
title: Neat-Enhancement-Idea
authors:
  - "@camilamacedo86"
reviewers:
  - TBD
approvers:
  - TBD
creation-date: 2020-11-09
last-updated: 2020-11-09
status: implementable
---

# New Subcommand Plugin to generate the common code

## Summary

This proposal defines a new feature [subcommand plugin][subcommand-plugin], which allow users get the scaffold with the
 required code to have a projects that will deploy and manage an image on the cluster following the good practices.  

## Motivation

The biggest part of the Kubebuilder users looking for to create a project that will at the end only deploy an image. In this way, one of the  mainly motivations of this proposal is to abstract the complexities to achieve this goal and still giving the possibility of users improve and customize their projects according to their requirements. 

### Goals

- Add a new plugin/subcommand to generate the code required to deploy and manage an image on the cluster
- Promote the best practices as give example of common implementations
- Make the process to develop  operators projects easier and more agil. 
- Give flexibility to the users and allow them to change the code according to their needs
- Provide examples of code implementations and of the most common features usage and reduce the learning curve
 
### Non-Goals

The idea of this proposal is provide a facility for the users which indeed could be improved 
in the future, however, this proposal just covers the basic requirements. In this way, is a non-goal;
allow  extra configurations such as scaffold the project using webhooks for example. However, it also could be a add in the future to this feature.

## Proposal

Add the new subcommand plugin code generate which will scaffold code implementation to deploy the image informed which would like such as; `kubebuilder create code --image=myexample:0.0.1 --group=example --kind=App --version=v1` which will:

- Create an API for the GVK informed with its controller which is the same steps used the subcommand plugin `create api` with `--resource` and `--controller` flags as true
- Add a code implementation which will do the Custom Resource reconciliation and create a Deployment resource for the `--image`:
- Add an EnvVar on the manager manifest (`config/manager/manager.yaml`) which will store the image informed and shows its possibility to users:

```yaml
    ..   
    spec:
      containers:
        - name: manager
          env:
            - name: {{ resource}}-IMAGE 
              value: {{image:tag}}
          image: controller:latest
      ...
```

- Add a check into reconcile to ensure that the replicas of the deployment on cluster are equals the size defined in the CR:

```go
	// Ensure the deployment size is the same as the spec
	size := {{ resource }}.Spec.Size
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}
```

- Add the watch feature for the Deployment managed by the controller: 

```go 
func (r *{{ resource }}Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.{{ resource }}{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
```
- Add the RBAC permissions required for the scenario such as:

```go
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
```

- A status [conditions][conditions] to allow users check that if the deployment occurred successfully or its errors
- The controller test implementation for this scenario
- Add a [marker][markers] in the spec definition to demonstrate how to use OpenAPI schemas validation such as `+kubebuilder:validation:Minimum=1`
- Add the specs on the `_types.go` to generate the CRD/CR sample with default values for `ImagePullPolicy` (`Always`), `ContainerPort` (`80`) and the `Replicas Size` (`3`) 
   
### User Stories

- I am as user, would like to use a command to scaffold my common need which is deploy an image of my application, so that I do not need know exactly how to implement it 
 - I am as user, would like to have a good example code base which uses the common features, so that I can easily learn its concepts and have a good start point to address my needs.  
 - I am as maintainer, would like to have a good example to address the common questions, so that I can easily describe how to implement the projects and/or use the common features.
 - I am as consumer(e.g sdk), I would like to be able to extending the new subcommand plugin such as I do with Create API, so that I can re-use this implementation for my tool and keep the both projects aligned
 
### Implementation Details/Notes/Constraints 

**Example of the controller template**

```go
// +kubebuilder:rbac:groups=cache.example.com,resources={{ resource.plural }},verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cache.example.com,resources={{ resource.plural }}/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cache.example.com,resources={{ resource.plural }}/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

func (r *{{ resource }}.Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("{{ resource }}", req.NamespacedName)

	// Fetch the {{ resource }} instance
	{{ resource }} := &{{ apiimportalias }}.{{ resource }}{}
	err := r.Get(ctx, req.NamespacedName, {{ resource }})
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("{{ resource }} resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get {{ resource }}")
		return ctrl.Result{}, err
	}

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: {{ resource }}.Name, Namespace: {{ resource }}.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentFor{{ resource }}({{ resource }})
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Ensure the deployment size is the same as the spec
	size := {{ resource }}.Spec.Size
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}
	
    // TODO: add here code implementation to update/manage the status
    
	return ctrl.Result{}, nil
}

// deploymentFor{{ resource }} returns a {{ resource }} Deployment object
func (r *{{ resource }}Reconciler) deploymentFor{{ resource }}(m *{{ apiimportalias }}.{{ resource }}) *appsv1.Deployment {
	ls := labelsFor{{ resource }}(m.Name)
	replicas := m.Spec.Size

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:   imageFor{{ resource }}(m.Name),
						Name:    {{ resource }},
                        ImagePullPolicy: {{ resource }}.Spec.ContainerImagePullPolicy,
						Command: []string{"{{ resource }}"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: {{ resource }}.Spec.ContainerPort,
							Name:          "{{ resource }}",
						}},
					}},
				},
			},
		},
	}
	// Set {{ resource }} instance as the owner and controller
	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

// labelsFor{{ resource }} returns the labels for selecting the resources
// belonging to the given {{ resource }} CR name.
func labelsFor{{ resource }}(name string) map[string]string {
	return map[string]string{"type": "{{ resource }}", "{{ resource }}_cr": name}
}

// imageFor{{ resource }} returns the image for the resources
// belonging to the given {{ resource }} CR name.
func imageFor{{ resource }}(name string) string {
	// TODO: this method will return the value of the envvar create to store the image:tag informed 
}

func (r *{{ resource }}Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.{{ resource }}{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

``` 

**Example of the spec for the <kind>_types.go template**

```go
// {{ resource }}Spec defines the desired state of {{ resource }}
type {{ resource }}Spec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

    // +kubebuilder:validation:Minimum=1
	// Size defines the number of {{ resource }} instances
	Size int32 `json:"size,omitempty"`

    // ImagePullPolicy defines the policy to pull the container images
	ImagePullPolicy string `json:"image-pull-policy,omitempty"`

    // ContainerPort specifies the port which will be used by the image container 
	ContainerPort int `json:"container-port,omitempty"`

}
```

## Design Details

### Test Plan

To ensure this implementation a new project example should be generated in the [testdata](../testdata/) directory of the project. See the [generate_testdata.sh](../generate_testdata.sh). Also, we should use this scaffold in the [integration tests](../test/e2e/) to ensure that the data scaffolded with works on the cluster as expected.  

### Graduation Criteria

- The new subcommand plugin should only be implemented for `v3+` plugin
- The new subcommand should re-use the `create api` plugin in order to centralize its code implementation and keep a better maintainability 
- The attribute image with the value informed should be added to the resources model in the PROJECT file to let the tool know that the Resource get done with the common basic code implementation. 

## Open Questions 

1. Should we allow to scaffold the code for an API that is already created for the project? 
No, at least in the first moment to keep the simplicity.  

2. Should we support StatefulSet and Deployments?
The idea is we start it by using a Deployment. However, we can improve the feature in follow-ups to support more default types of scaffolds which could be like `kubebuilder create code --image=myexample:0.0.1 --group=example --kind=App --version=v1 --type=[deployment|statefulset|webhook]`

3. Could this feature be useful to other languages or is it just valid to Go based operators?

For its integration with SDK, it might be valid for the Ansible based-operators where a new playbook/role could be generated as well. However, for Helm for example it might be useless.    

4. Why not use the `create api` subcommand to generate the code?

The generated code might not valid for other types of operators. So, Helm in SDK, for example, which extending the `create api` should not have a flag to generate the code by default in this subcommand. 

In this way, the new subcommand is required to not hurt use cases such as; _I am as consumer of Kubebuilder pkg, I would like to decide what subcommand I want to extend, so that I can for example use the `create api` but not the `create code` in my tool_ and _I am as maintainer of Kubebuiler, I would like to see the subcommands implemented with separation of concerns, so that I do not care about one subcommand when I am developing another_. 

Other valid argumentation to address this need in a new subcommand shows be that add it as an option of the `create api` subcommand would reduce the maintainability since add this feature offered via the `create api` might hurting concepts such as single responsibility and cohesion. 

It is important to highlight that the `--resouce` and `--controler` flags available for `create api` which allow users define if the API and controller should or not be generated are not valid ones for this feature. The `create code` requires both always.   

However, address this new feature via a new subcommand do not remove the possibility to change it as well in a long term and provide this option by other command since it would be easily achieved after all get done by a follow up.     

5. This feature looks a lot like the existing addon plugin, could not it work as addon option currently? 

The addon should be port to be a plugin as well. More info: #1543. Note that, the addon option shows not fit well as a default option for the `create api` for the same reasons described above. So we might need to looking for to port this option as a new subcommand either.
   
6. What about we aggregate this feature with the `create api` when the plugin phase 2 be in place? 

It shows a valid option, however, the plugin phase 2 is not implemented or well specified yet. So, we might need to wait for the plugin phase 2 be in place to see how we can improve this feature as others in order to take advantage of this new design. More info: #1378.  

[markers]: ../docs/book/src/reference/markers.md 
[conditions]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
[subcommand-plugin]: ../pkg/plugin/v3/plugin.go