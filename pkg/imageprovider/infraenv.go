package imageprovider

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pkg/errors"

	aiv1beta1 "github.com/openshift/assisted-service/api/v1beta1"
)

const infraEnvLabel = "infraenvs.agent-install.openshift.io"

// GetInfraEnv returns a linked InfraEnv object for the Host or nil if none is linked.
func GetInfraEnv(reader client.Reader, imageMetadata *metav1.ObjectMeta) (*aiv1beta1.InfraEnv, error) {
	infraenvName, ok := imageMetadata.GetLabels()[infraEnvLabel]
	if !ok {
		return nil, nil
	}

	infraenv := &aiv1beta1.InfraEnv{}
	infraenvKey := client.ObjectKey{
		Namespace: imageMetadata.Namespace,
		Name:      infraenvName,
	}
	// NOTE(dtantsur): using a Reader since InfraEnv objects are not cached
	if err := reader.Get(context.TODO(), infraenvKey, infraenv); err != nil {
		return nil, errors.Wrapf(err, "could not get linked InfraEnv %s for host %s in namespace %s", infraenvKey, imageMetadata.Name, imageMetadata.Namespace)
	}

	return infraenv, nil
}
