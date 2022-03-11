package imageprovider

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metal3 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/metal3-io/baremetal-operator/pkg/imageprovider"
	"github.com/openshift/image-customization-controller/pkg/env"
	"github.com/openshift/image-customization-controller/pkg/ignition"
	"github.com/openshift/image-customization-controller/pkg/imagehandler"
)

type rhcosImageProvider struct {
	ImageHandler   imagehandler.ImageHandler
	EnvInputs      *env.EnvInputs
	RegistriesConf []byte
	client         client.Client
	apiReader      client.Reader
}

func NewRHCOSImageProvider(imageServer imagehandler.ImageHandler, inputs *env.EnvInputs, client client.Client, apiReader client.Reader) imageprovider.ImageProvider {
	registries, err := inputs.RegistriesConf()
	if err != nil {
		panic(err)
	}

	return &rhcosImageProvider{
		ImageHandler:   imageServer,
		EnvInputs:      inputs,
		RegistriesConf: registries,
		client:         client,
		apiReader:      apiReader,
	}
}

func (ip *rhcosImageProvider) SupportsArchitecture(arch string) bool {
	return true
}

func (ip *rhcosImageProvider) SupportsFormat(format metal3.ImageFormat) bool {
	switch format {
	case metal3.ImageFormatISO, metal3.ImageFormatInitRD:
		return true
	default:
		return false
	}
}

func (ip *rhcosImageProvider) buildIgnitionConfig(networkData imageprovider.NetworkData, hostname string, mergeWith []byte) ([]byte, error) {
	nmstateData := networkData["nmstate"]

	builder, err := ignition.New(nmstateData, ip.RegistriesConf,
		ip.EnvInputs.IronicBaseURL,
		ip.EnvInputs.IronicAgentImage,
		ip.EnvInputs.IronicAgentPullSecret,
		ip.EnvInputs.IronicRAMDiskSSHKey,
		ip.EnvInputs.IpOptions,
		ip.EnvInputs.HttpProxy,
		ip.EnvInputs.HttpsProxy,
		ip.EnvInputs.NoProxy,
		hostname,
	)
	if err != nil {
		return nil, imageprovider.BuildInvalidError(err)
	}

	err, message := builder.ProcessNetworkState()
	if message != "" {
		return nil, imageprovider.BuildInvalidError(errors.New(message))
	}
	if err != nil {
		return nil, err
	}

	return builder.GenerateAndMergeWith(mergeWith)
}

func imageKey(data imageprovider.ImageData) string {
	return fmt.Sprintf("%s-%s-%s-%s.%s",
		data.ImageMetadata.Namespace,
		data.ImageMetadata.Name,
		data.ImageMetadata.UID,
		data.Architecture,
		data.Format,
	)
}

func (ip *rhcosImageProvider) BuildImage(data imageprovider.ImageData, networkData imageprovider.NetworkData, log logr.Logger) (string, error) {
	url, err := ip.buildImageWithInfraEnv(data, log)
	if url != "" || err != nil {
		return url, err
	}

	ignitionConfig, err := ip.buildIgnitionConfig(networkData, data.ImageMetadata.Name, nil)
	if err != nil {
		return "", err
	}

	url, err = ip.ImageHandler.ServeImage(imageKey(data), ignitionConfig,
		data.Format == metal3.ImageFormatInitRD, false)
	if errors.As(err, &imagehandler.InvalidBaseImageError{}) {
		return "", imageprovider.BuildInvalidError(err)
	}
	return url, err
}

func (ip *rhcosImageProvider) buildImageWithInfraEnv(data imageprovider.ImageData, log logr.Logger) (string, error) {
	infraenv, err := GetInfraEnv(ip.apiReader, data.ImageMetadata)
	if err != nil {
		return "", err
	}
	if infraenv == nil {
		// Fall back to the regular path
		return "", nil
	}

	log.Info("using InfraEnv to build an image, network data will be ignored", "hostName", data.ImageMetadata.Name, "infraEnv", infraenv.Name)

	ignitionConfig, err := ip.buildIgnitionConfig(nil, data.ImageMetadata.Name, []byte(infraenv.Spec.IgnitionConfigOverride))
	if err != nil {
		return "", err
	}

	if err = UpdateInfraEnv(ip.client, infraenv, string(ignitionConfig), log); err != nil {
		return "", err
	}

	return GetImageFromInfraEnv(infraenv, data.Format, log)
}

func (ip *rhcosImageProvider) DiscardImage(data imageprovider.ImageData) error {
	if GetInfraEnvName(data.ImageMetadata) == "" {
		ip.ImageHandler.RemoveImage(imageKey(data))
	}

	return nil
}
