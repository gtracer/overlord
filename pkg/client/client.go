package client

import (
	"github.com/pkg/errors"

	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlcfg "sigs.k8s.io/controller-runtime/pkg/client/config"
)

// New ...
func New() (ctrlclient.Client, error) {
	cfg, err := ctrlcfg.GetConfig()
	if err != nil {
		return nil, errors.Errorf("failed to get config %v", err)
	}

	return ctrlclient.New(cfg, ctrlclient.Options{})
}
