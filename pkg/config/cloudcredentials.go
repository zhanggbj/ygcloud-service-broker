package config

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// CloudCredentials define
type CloudCredentials struct {
	K8sCfgPath   string `json:"K8s_config_path"`
	ClientConfig clientcmd.ClientConfig
	ClientSet    kubernetes.Interface
}

func (c *CloudCredentials) Validate() error {
	if c.ClientSet == nil {
		restConfig, err := c.RestConfig()
		if err != nil {
			return err
		}

		c.ClientSet, err = kubernetes.NewForConfig(restConfig)
		if err != nil {
			fmt.Println("failed to create client:", err)
			os.Exit(1)
		}
	}
	return nil
}

// RestConfig returns REST config, which can be to use to create specific clientset
func (c *CloudCredentials) RestConfig() (*rest.Config, error) {
	var err error

	if c.ClientConfig == nil {
		c.ClientConfig, err = c.GetClientConfig()
		if err != nil {
			return nil, err
		}
	}

	config, err := c.ClientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	return config, nil
}

// GetClientConfig gets ClientConfig from KubeCfgPath
func (c *CloudCredentials) GetClientConfig() (clientcmd.ClientConfig, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if len(c.K8sCfgPath) == 0 {
		return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{}), nil
	}

	_, err := os.Stat(c.K8sCfgPath)
	if err == nil {
		loadingRules.ExplicitPath = c.K8sCfgPath
		return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{}), nil
	}

	if !os.IsNotExist(err) {
		return nil, err
	}

	paths := filepath.SplitList(c.K8sCfgPath)
	if len(paths) > 1 {
		return nil, fmt.Errorf("Can not find config file. '%s' looks like a path. "+
			"Please use the env var KUBECONFIG if you want to check for multiple configuration files", c.K8sCfgPath)
	}
	return nil, fmt.Errorf("Config file '%s' can not be found", c.K8sCfgPath)
}
