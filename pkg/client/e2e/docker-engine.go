package e2e

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/helm/helm-classic/codec"
)

/* https://github.com/docker/docker/blob/master/cliconfig/configfile/file.go */
type DockerConfigFile struct {
	AuthConfigs      map[string]DockerAuthConfig `json:"auths"`
	HTTPHeaders      map[string]string           `json:"HttpHeaders,omitempty"`
	PsFormat         string                      `json:"psFormat,omitempty"`
	ImagesFormat     string                      `json:"imagesFormat,omitempty"`
	NetworksFormat   string                      `json:"networksFormat,omitempty"`
	VolumesFormat    string                      `json:"volumesFormat,omitempty"`
	DetachKeys       string                      `json:"detachKeys,omitempty"`
	CredentialsStore string                      `json:"credsStore,omitempty"`
	Filename         string                      `json:"-"` // Note: for internal use only
}

/* https://github.com/docker/engine-api/blob/master/types/auth.go */
type DockerAuthConfig struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Auth     string `json:"auth,omitempty"`

	// Email is an optional value associated with the username.
	// This field is deprecated and will be removed in a later
	// version of docker.
	Email string `json:"email,omitempty"`

	ServerAddress string `json:"serveraddress,omitempty"`

	// IdentityToken is used to authenticate the user and get
	// an access token for the registry.
	IdentityToken string `json:"identitytoken,omitempty"`

	// RegistryToken is a bearer token to be sent to a registry
	RegistryToken string `json:"registrytoken,omitempty"`
}

func SerializeDockerConfigFile(dcf *DockerConfigFile) ([]byte, error) {
	b := &bytes.Buffer{}
	if err := codec.JSON.Encode(b).One(dcf); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func SerializeIntoDockerConfigFile(auth *DockerAuthConfig) ([]byte, *DockerConfigFile, error) {
	basicauth := fmt.Sprintf("%s:%s", auth.Username, auth.Password)
	auth.Auth = base64.StdEncoding.EncodeToString([]byte(basicauth))

	dcf := &DockerConfigFile{
		AuthConfigs: map[string]DockerAuthConfig{
			auth.ServerAddress: {
				Auth: auth.Auth,
			},
		},
	}
	b, err := SerializeDockerConfigFile(dcf)
	if err != nil {
		return nil, dcf, err
	}
	return b, dcf, nil
}
