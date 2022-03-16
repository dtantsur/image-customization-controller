module github.com/openshift/image-customization-controller

go 1.16

require (
	github.com/coreos/ignition/v2 v2.12.0
	github.com/coreos/vcontext v0.0.0-20210407161507-4ee6c745c8bd
	github.com/go-logr/logr v1.2.2
	github.com/golangci/golangci-lint v1.32.0
	github.com/google/go-cmp v0.5.5
	github.com/google/uuid v1.1.2
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/metal3-io/baremetal-operator v0.0.0
	github.com/metal3-io/baremetal-operator/apis v0.0.0
	github.com/openshift/assisted-image-service v0.0.0-20220301135350-10a987fbc261
	github.com/openshift/assisted-service/api v0.0.0-20220311025016-e574bc2de2fd
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/vincent-petithory/dataurl v0.0.0-20160330182126-9a301d65acbb
	k8s.io/api v0.23.4
	k8s.io/apimachinery v0.23.4
	k8s.io/client-go v0.23.0
	k8s.io/utils v0.0.0-20211116205334-6203023598ed
	sigs.k8s.io/controller-runtime v0.11.1
	sigs.k8s.io/yaml v1.3.0
)

replace (
	github.com/metal3-io/baremetal-operator => github.com/openshift/baremetal-operator v0.0.0-20220314153030-57fef6af93c7
	github.com/metal3-io/baremetal-operator/apis => github.com/openshift/baremetal-operator/apis v0.0.0-20220314153030-57fef6af93c7
	github.com/metal3-io/baremetal-operator/pkg/hardwareutils => github.com/openshift/baremetal-operator/pkg/hardwareutils v0.0.0-20220314153030-57fef6af93c7
	github.com/openshift/assisted-service/models => github.com/openshift/assisted-service/models v0.0.0-20220311025016-e574bc2de2fd
)
