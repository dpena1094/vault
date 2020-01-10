package kubernetes

import (
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net"
	"os"

	"github.com/hashicorp/vault/sdk/helper/certutil"
)

const (
	EnvVarKubernetesServiceHost = "KUBERNETES_SERVICE_HOST"
	EnvVarKubernetesServicePort = "KUBERNETES_SERVICE_PORT"
)

var (
	ErrNotInCluster = errors.New("unable to load in-cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined")

	// These are presented as variables so they can be updated
	// to point at test fixtures if needed.
	scheme     = "https://"
	tokenFile  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	rootCAFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
)

// inClusterConfig returns a config object which uses the service account
// kubernetes gives to pods. It's intended for clients that expect to be
// running inside a pod running on kubernetes. It will return ErrNotInCluster
// if called from a process not running in a kubernetes environment.
func inClusterConfig() (*Config, error) {
	host, port := os.Getenv(EnvVarKubernetesServiceHost), os.Getenv(EnvVarKubernetesServicePort)
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
		Host:            scheme + net.JoinHostPort(host, port),
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
