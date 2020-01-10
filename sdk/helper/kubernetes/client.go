package kubernetes

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/go-cleanhttp"
)

var ErrNotFound = errors.New("not found")

type Tag struct {
	Key, Value string
}

func NewLightWeightClient() (LightWeightClient, error) {
	config, err := inClusterConfig()
	if err != nil {
		return nil, err
	}
	return &lightWeightClient{
		config: config,
	}, nil
}

type LightWeightClient interface {
	// GetPod merely verifies a pod's existence, returning an
	// error if the pod doesn't exist.
	GetPod(namespace, podName string) error

	// UpdatePodTags updates the pod's tags tags to the given ones,
	// overwriting previous values for a given tag key. It does so
	// non-destructively, or in other words, without tearing down
	// the pod.
	UpdatePodTags(namespace, podName string, tags ...*Tag) error
}

type lightWeightClient struct {
	config *Config
}

func (c *lightWeightClient) GetPod(namespace, podName string) error {
	endpoint := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s", namespace, podName)
	method := http.MethodGet

	req, err := http.NewRequest(method, c.config.Host+endpoint, nil)
	if err != nil {
		return err
	}
	if err := c.do(req, nil); err != nil {
		return err
	}
	return nil
}

func (c *lightWeightClient) UpdatePodTags(namespace, podName string, tags ...*Tag) error {
	endpoint := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s", namespace, podName)
	method := http.MethodPatch

	var patch []interface{}
	for _, tag := range tags {
		patch = append(patch, map[string]string{
			"op":    "add",
			"path":  "/metadata/labels/" + tag.Key,
			"value": tag.Value,
		})
	}
	body, err := json.Marshal(patch)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(method, c.config.Host+endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json-patch+json")
	return c.do(req, nil)
}

func (c *lightWeightClient) do(req *http.Request, ptrToReturnObj interface{}) error {
	// Finish setting up a valid request.
	req.Header.Set("Authorization", "Bearer "+c.config.BearerToken)
	req.Header.Set("Accept", "application/json")
	client := cleanhttp.DefaultClient()
	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: c.config.CACertPool,
		},
	}

	// Execute it.
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// Check for success.
	// Note - someday we may want to implement retrying and exponential backoff with jitter
	// here, but on the first iteration we're going with a naive approach because we don't
	// know whether it'll be needed.
	switch resp.StatusCode {
	case 200, 201, 202:
		// Pass.
	case 404:
		return ErrNotFound
	default:
		return fmt.Errorf("unexpected status code: %s", sanitizedDebuggingInfo(req, resp))
	}

	// If we're not supposed to read out the body, we have nothing further
	// to do here.
	if ptrToReturnObj == nil {
		return nil
	}

	// Attempt to read out the body into the given return object.
	if err := json.NewDecoder(resp.Body).Decode(ptrToReturnObj); err != nil {
		return fmt.Errorf("unable to read as %T: %s", ptrToReturnObj, sanitizedDebuggingInfo(req, resp))
	}
	return nil
}

// sanitizedDebuggingInfo converts an http response to a string without
// including its headers, to avoid leaking authorization
// headers.
func sanitizedDebuggingInfo(req *http.Request, resp *http.Response) string {
	// Ignore error here because if we're unable to read the body or
	// it doesn't exist, it'll just be "", which is fine.
	body, _ := ioutil.ReadAll(resp.Body)
	return fmt.Sprintf("method: %s, url: %s, statuscode: %d, body: %s", req.Method, req.URL, resp.StatusCode, body)
}
