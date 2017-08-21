package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher/v3"
	"k8s.io/client-go/rest"
)

const (
	rancherCredentialsFolder = "/rancher-credentials"
	urlKeyFilename           = "url"
	accessKeyFilename        = "access-key"
	secretKeyFilename        = "secret-key"

	kubernetesServiceHostKey = "KUBERNETES_SERVICE_HOST"
	kubernetesServicePortKey = "KUBERNETES_SERVICE_PORT"
)

func main() {
	if err := runReporter(); err != nil {
		log.Fatalf("Failed to report back to Rancher: %v", err)
	}
}

func runReporter() error {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	if err := populateCAData(cfg); err != nil {
		return err
	}

	kubernetesServiceHost, err := getenv(kubernetesServiceHostKey)
	if err != nil {
		return err
	}
	kubernetesServicePort, err := getenv(kubernetesServicePortKey)
	if err != nil {
		return err
	}

	rancherClient, err := getRancherClient()
	if err != nil {
		return err
	}

	for {
		_, err = rancherClient.Register.Create(&client.Register{
			K8sClientConfig: &client.K8sClientConfig{
				Address:     fmt.Sprintf("%s:%s", kubernetesServiceHost, kubernetesServicePort),
				BearerToken: cfg.BearerToken,
				CaCert:      string(cfg.CAData),
			},
		})
		if err == nil {
			return nil
		}
		log.Errorf("Failed to create registration in Rancher: %v", err)
		time.Sleep(5 * time.Second)
	}
}

func getenv(env string) (string, error) {
	value := os.Getenv(env)
	if value == "" {
		return "", fmt.Errorf("%s is empty", env)
	}
	return value, nil
}

func populateCAData(cfg *rest.Config) error {
	bytes, err := ioutil.ReadFile(cfg.CAFile)
	if err != nil {
		return err
	}
	cfg.CAData = bytes
	return nil
}

func getRancherClient() (*client.RancherClient, error) {
	url, err := readKey(urlKeyFilename)
	if err != nil {
		return nil, err
	}
	accessKey, err := readKey(accessKeyFilename)
	if err != nil {
		return nil, err
	}
	secretKey, err := readKey(secretKeyFilename)
	if err != nil {
		return nil, err
	}
	return client.NewRancherClient(&client.ClientOpts{
		Url:       url,
		AccessKey: accessKey,
		SecretKey: secretKey,
	})
}

func readKey(key string) (string, error) {
	bytes, err := ioutil.ReadFile(path.Join(rancherCredentialsFolder, key))
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
