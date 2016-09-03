package docker

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/docker/docker/cliconfig"
	"github.com/docker/engine-api/types"
	"github.com/golang/glog"
	"github.com/helm/helm-classic/codec"
)

func SerializeConfigFile(cf *cliconfig.ConfigFile) ([]byte, error) {
	b := &bytes.Buffer{}
	if err := codec.JSON.Encode(b).One(cf); err != nil {
		glog.Errorf("Failed to encode: %+v", err)
		return nil, err
	}
	return b.Bytes(), nil
}

func SerializeIntoConfigFile(ac *types.AuthConfig) ([]byte, *cliconfig.ConfigFile, error) {
	if ac == nil {
		err := fmt.Errorf("auth config could not null")
		glog.Errorln("args required")
		return nil, nil, err
	}
	basicauth := fmt.Sprintf("%s:%s", ac.Username, ac.Password)
	auth := base64.StdEncoding.EncodeToString([]byte(basicauth))
	if len(ac.Auth) == 0 || strings.Compare(ac.Auth, auth) != 0 {
		ac.Auth = auth
	}

	cf := &cliconfig.ConfigFile{
		AuthConfigs: map[string]cliconfig.AuthConfig{
			ac.ServerAddress: {
				Auth: ac.Auth,
			},
		},
	}
	b, err := SerializeConfigFile(cf)
	if err != nil {
		return nil, cf, err
	}
	return b, cf, nil
}
