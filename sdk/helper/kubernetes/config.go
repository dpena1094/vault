package kubernetes

import (
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net"
	"os"

	"github.com/hashicorp/vault/sdk/helper/certutil"
)

var ErrNotInCluster = errors.New("unable to load in-cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined")

// TODO write a test using this fixture, note no newline at end.
/*
root@minikube:/# cat /var/run/secrets/kubernetes.io/serviceaccount/token
eyJhbGciOiJSUzI1NiIsImtpZCI6IjZVQU91ckJYcTZKRHQtWHpaOExib2EyUlFZQWZObms2d25mY3ZtVm1NNUUifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImRlZmF1bHQtdG9rZW4tNWZqdDkiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoiZGVmYXVsdCIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6ImY0NGUyMDIxLTU2YWItNDEzNC1hMjMxLTBlMDJmNjhmNzJhNiIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmRlZmF1bHQifQ.hgMbuT0hlxG04fDvI_Iyxtbwc8M-i3q3K7CqIGC_jYSjVlyezHN_0BeIB3rE0_M2xvbIs6chsWFZVsK_8Pj6ho7VT0x5PWy5n6KsqTBz8LPpjWpsaxpYQos0RzgA3KLnuzZE8Cl-v-PwWQK57jgbS4AdlXujQXdtLXJNwNAKI0pvCASA6UXP55_X845EsJkyT1J-bURSS3Le3g9A4pDoQ_MUv7hqa-p7yQEtFfYCkq1KKrUJZMRjmS4qda1rg-Em-dw9RFvQtPodRYF0DKT7A7qgmLUfIkuky3NnsQtvaUo8ZVtUiwIEfRdqw1oQIY4CSYz-wUl2xZa7n2QQBROE7wroot@minikube:/#
*/

// inClusterConfig returns a config object which uses the service account
// kubernetes gives to pods. It's intended for clients that expect to be
// running inside a pod running on kubernetes. It will return ErrNotInCluster
// if called from a process not running in a kubernetes environment.
func inClusterConfig() (*Config, error) {
	const (
		tokenFile  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
		rootCAFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	)
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if len(host) == 0 || len(port) == 0 {
		return nil, ErrNotInCluster
	}

	token, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		return nil, err
	}

	pool, err := certutil.NewCertPool(rootCAFile)
	if err != nil {
		return nil, err
	}
	return &Config{
		Host:            "https://" + net.JoinHostPort(host, port),
		CACertPool:      pool,
		BearerToken:     string(token),
		BearerTokenFile: tokenFile, // TODO should I re-check this periodically? Or lazily on a bad response?
	}, nil
}

type Config struct {
	Host            string
	BearerToken     string
	BearerTokenFile string
	CACertPool      *x509.CertPool
}
