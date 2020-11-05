package utils

import (
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"strconv"
)

func GetK8sVersion() (major, minor int, ver string, err error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return 0, 0, "", err
	}
	client, err := kube_client.NewForConfig(cfg)
	if err != nil {
		return 0, 0, "", err
	}

	v, err := client.Discovery().ServerVersion()
	if err != nil {
		return 0, 0, "", err
	}
	major, err = strconv.Atoi(v.Major)
	minor, err = strconv.Atoi(v.Minor)
	ver = v.String()
	return
}
