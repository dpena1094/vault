package kubernetes

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
)

// CertificateBlockType is a possible value for pem.Block.Type.
const CertificateBlockType = "CERTIFICATE"

var ErrNotInCluster = errors.New("unable to load in-cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined")

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

	tlsClientConfig := TLSClientConfig{}

	if _, err := newPool(rootCAFile); err != nil {
			return nil, fmt.Errorf("expected to load root CA config from %s, but got err: %v", rootCAFile, err)
	} else {
		tlsClientConfig.CAFile = rootCAFile
	}

	return &Config{
		// TODO: switch to using cluster DNS.
		Host:            "https://" + net.JoinHostPort(host, port),
		TLSClientConfig: tlsClientConfig,
		BearerToken:     string(token),
		BearerTokenFile: tokenFile,
	}, nil
}

// newPool returns an x509.CertPool containing the certificates in the given PEM-encoded file.
// Returns an error if the file could not be read, a certificate could not be parsed, or if the file does not contain any certificates
func newPool(filename string) (*x509.CertPool, error) {
	certs, err := certsFromFile(filename)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	for _, cert := range certs {
		pool.AddCert(cert)
	}
	return pool, nil
}

// CertsFromFile returns the x509.Certificates contained in the given PEM-encoded file.
// Returns an error if the file could not be read, a certificate could not be parsed, or if the file does not contain any certificates
func certsFromFile(file string) ([]*x509.Certificate, error) {
	pemBlock, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	certs, err := parseCertsPEM(pemBlock)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %s", file, err)
	}
	return certs, nil
}

// Config holds the common attributes that can be passed to a Kubernetes client on
// initialization.
type Config struct {
	// Host must be a host string, a host:port pair, or a URL to the base of the apiserver.
	// If a URL is given then the (optional) Path of that URL represents a prefix that must
	// be appended to all request URIs used to access the apiserver. This allows a frontend
	// proxy to easily relocate all of the apiserver endpoints.
	Host string

	// Server requires Bearer authentication. This client will not attempt to use
	// refresh tokens for an OAuth2 flow.
	// TODO: demonstrate an OAuth2 compatible client.
	BearerToken string

	// Path to a file containing a BearerToken.
	// If set, the contents are periodically read.
	// The last successfully read value takes precedence over BearerToken.
	BearerTokenFile string

	// TLSClientConfig contains settings to enable transport layer security
	TLSClientConfig
}

// TLSClientConfig contains settings to enable transport layer security
type TLSClientConfig struct {
	// Trusted root certificates for server
	CAFile string
}

// TODO a lot of this cert stuff looks a lot like our own cert parsing stuff, can we replace it?
// parseCertsPEM returns the x509.Certificates contained in the given PEM-encoded byte array
// Returns an error if a certificate could not be parsed, or if the data does not contain any certificates
func parseCertsPEM(pemCerts []byte) ([]*x509.Certificate, error) {
	ok := false
	certs := []*x509.Certificate{}
	for len(pemCerts) > 0 {
		var block *pem.Block
		block, pemCerts = pem.Decode(pemCerts)
		if block == nil {
			break
		}
		// Only use PEM "CERTIFICATE" blocks without extra headers
		if block.Type != CertificateBlockType || len(block.Headers) != 0 {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return certs, err
		}

		certs = append(certs, cert)
		ok = true
	}

	if !ok {
		return certs, errors.New("data does not contain any valid RSA or ECDSA certificates")
	}
	return certs, nil
}
