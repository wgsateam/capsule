package utils

import (
	"path/filepath"
	"strconv"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func GetK8sVersion() (major, minor int, ver string, err error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if err != nil {
		return 0, 0, "", err
	}
	client, err := kubernetes.NewForConfig(cfg)
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
