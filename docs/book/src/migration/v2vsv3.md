# Kubebuilder v2 vs v3

This document covers all breaking changes when migrating from v2 to v3.

The details of all changes (breaking or otherwise) can be found in [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder/releases) release notes.

## Kubebuilder

- A plugin design was introduced to the project. For more info see the [Extensible CLI and Scaffolding Plugins][plugins-phase1-design-doc]
- The GO supported version was upgraded from 1.13+ to 1.15+
- Support for `apiextensions.k8s.io/v1` to replace `apiextensions.k8s.io/v1beta1` which was deprecated in Kubernetes version `1.16`.
- Support for `admissionregistration.k8s.io/v1` to replace `admissionregistration.k8s.io/v1beta1` which was deprecated in Kubernetes version `1.16`.
- Support for `cert-manager.io/v1` to replace `cert-manager.io/v1alpha2`. Check [CertManager v1.0 docs]( https://cert-manager.io/docs/installation/upgrading/) for breaking changes.
- The manager flags `--metrics-addr` and `enable-leader-election` were renamed for `--metrics-bind-address` and `--leader-elect` to be more aligned with k8s Kubernetes Components. More info: [#1839][issue-1893] 
- Liveness and Readiness probes are now generate by default using [`healthz.Ping`][healthz-ping] 
- New option to create the projects using ComponentConfig was introduced. For more info see its [enhancement proposal](https://github.com/kubernetes/enhancements/tree/master/keps/sig-cluster-lifecycle/wgs) and the [Component config tutorial][component-config-tutorial]
- Improvements in the scaffolds to address security concerns

<aside class="note warning">
<h1>Note</h1>

From now on, the new projects build with the tool will be using by default the go plugin `go.kubebuilder.io/v3+` which has the support for the new features and APIs. However, it has breaking changes as well.

In this way, users will still be able to continue using the legacy layout without breaking changes and get its bug fixes via the plugin `go.kubebuilder.io/v2` by only upgrading their PROJECT
version.  
</aside>

## Migration Guide v2 to V3

The following guide will cover the steps in the most straightforward way to allow you to upgrade your project to get all latest changes and improvements.

- [Migration Guide v2 to V3][migration-guide-v2-to-v3]
              
## Upgrade ONLY the project version

The following guide will describe the steps required for
you to upgrade only your PROJECT version.

It is recommended for who would like to still using `go.kubebuilder.io/v2` and get its bug fixes but avoid breaking changes and for who is looking for to follow up the [Migrate Plugin Version v2 to v3][plugin-v3] and instead of re-create the project from scratch and only apply the required changes. 

- [Migrate Project Version v2 to v3][project-v3]

## Upgrade Plugin versions (`go.kubebuilder.io`)

- [Migrate Plugin Version v2 to v3][plugin-v3]

[plugins-phase1-design-doc]: https://github.com/kubernetes-sigs/kubebuilder/blob/master/designs/extensible-cli-and-scaffolding-plugins-phase-1.md
[project-v3]: /migration/project/v2_v3.md
[plugin-v3]: /migration/plugin/v2_v3.md
[component-config-tutorial]: ../component-config-tutorial/tutorial.md
[issue-1893]: https://github.com/kubernetes-sigs/kubebuilder/issues/1839
[migration-guide-v2-to-v3]: migration_guide_v2tov3.md
[healthz-ping]: https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/healthz#CheckHandler 