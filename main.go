package main

import (
	"fmt"
	"os"
	"path"
	"io/ioutil"

	"k8s.io/client-go/rest"
	"github.com/rancher/go-rancher/v3"
	"k8s.io/client-go/kubernetes"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	rancherCredentialsFolder = "/rancher-credentials"
	urlKeyFilename = "url"
	accessKeyFilename = "access-key"
	secretKeyFilename = "secret-key"

	kubernetesServiceHostKey = "KUBERNETES_SERVICE_HOST"
	kubernetesServicePortKey = "KUBERNETES_SERVICE_PORT"
)

func main() {
	if err := runReporter(); err != nil {
		panic(err)
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

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}

	kubeSystem, err := clientset.Namespaces().Get("kube-system", v1.GetOptions{})
	if err != nil {
		return err
	}

	rancherClient, err := getRancherClient()
	if err != nil {
		return err
	}

	_, err = rancherClient.Register.Create(&client.Register{
		Key: fmt.Sprint(kubeSystem.UID),
		K8sClientConfig: &client.K8sClientConfig{
			Address: fmt.Sprintf("%s:%s", kubernetesServiceHost, kubernetesServicePort),
			BearerToken: cfg.BearerToken,
			CaCert: string(cfg.CAData),
		},
	})
	return err
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
		Url: url,
		AccessKey: accessKey,
		SecretKey: secretKey,
	})
}

func readKey(key string) (string, error){
	bytes, err := ioutil.ReadFile(path.Join(rancherCredentialsFolder, key))
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}