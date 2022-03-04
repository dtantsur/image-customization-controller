package imageprovider

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/metal3-io/baremetal-operator/pkg/imageprovider"

	aiv1beta1 "github.com/openshift/assisted-service/api/v1beta1"
)

var (
	InfraEnvImageCooldownPeriod = 60 * time.Second
)

func GetInfraEnvName(imageMetadata *metav1.ObjectMeta) string {
	return imageMetadata.GetLabels()[aiv1beta1.InfraEnvNameLabel]
}

// GetInfraEnv returns a linked InfraEnv object for the Host or nil if none is linked.
func GetInfraEnv(reader client.Reader, imageMetadata *metav1.ObjectMeta) (*aiv1beta1.InfraEnv, error) {
	infraenvName := GetInfraEnvName(imageMetadata)
	if infraenvName == "" {
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

func jsonEqual(our string, their string) (bool, error) {
	if their == "" {
		return false, nil
	}

	var ourParsed, theirParsed map[string]interface{}

	// FIXME(dtantsur): this code is parsing something we've just generated.
	if err := json.Unmarshal([]byte(our), &ourParsed); err != nil {
		return false, errors.Wrap(err, "could not parse our ignition for comparison")
	}

	if err := json.Unmarshal([]byte(their), &theirParsed); err != nil {
		return false, errors.Wrap(err, "could not parse the current InfraEnv ignition")
	}

	return reflect.DeepEqual(ourParsed, theirParsed), nil
}

// UpdateInfraEnv updates the provided InfraEnv with the provided ignitionConfig.
// If the InfraEnv is already ready, nil is returned.
func UpdateInfraEnv(client client.Client, infraenv *aiv1beta1.InfraEnv, ignitionConfig string, log logr.Logger) error {
	equivalent, err := jsonEqual(ignitionConfig, infraenv.Spec.IgnitionConfigOverride)
	if err != nil {
		return err
	}

	if !equivalent {
		log.Info("updating InfraEnv with the merged ignition", "infraEnv", infraenv.Name)

		infraenv.Spec.IgnitionConfigOverride = ignitionConfig
		if err := client.Update(context.TODO(), infraenv); err != nil {
			return err
		}

		// Tell the caller that the image is not ready
		return imageprovider.ImageNotReady{}
	}

	condition := conditionsv1.FindStatusCondition(infraenv.Status.Conditions, aiv1beta1.ImageCreatedCondition)
	if condition != nil && condition.Status == corev1.ConditionTrue && condition.Reason == aiv1beta1.ImageCreatedReason {
		imageReadyTime := infraenv.Status.CreatedTime.Time.Add(InfraEnvImageCooldownPeriod)
		if imageReadyTime.After(time.Now()) {
			// NOTE(dtantsur): this replicates the logic in the assisted service
			log.Info("InfraEnv is too recent, requeueing", "infraEnv", infraenv.Name, "until", imageReadyTime)
			return imageprovider.ImageNotReady{}
		}
		return nil
	}

	if condition != nil {
		log.Info("InfraEnv is not ready", "infraEnv", infraenv.Name, "reason", condition.Reason, "message", condition.Message)
	} else {
		log.Info("InfraEnv is not reconciled yet", "infraEnv", infraenv.Name)
	}
	return imageprovider.ImageNotReady{}
}
