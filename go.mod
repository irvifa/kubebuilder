module sigs.k8s.io/kubebuilder

go 1.12

require (
	github.com/gobuffalo/flect v0.1.5
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/pkg/errors v0.8.1
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	golang.org/x/sys v0.0.0-20190621203818-d432491b9138 // indirect
	golang.org/x/tools v0.0.0-20190614205625-5aca471b1d59
	google.golang.org/appengine v1.5.0 // indirect
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/apimachinery v0.0.0-20190927035529-0104e33c351d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/kube-state-metrics v1.7.2
	sigs.k8s.io/controller-runtime v0.2.2
)
