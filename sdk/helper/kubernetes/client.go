package kubernetes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/go-cleanhttp"
)

const (
	baseURL = "" // TODO
	versionedAPIPath = "" // TODO
)

// Pod is a collection of containers that can run on a host. This resource is created
// by clients and scheduled onto hosts.
type Pod struct {
	ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
}

type ObjectMeta struct {
	Labels map[string]string `json:"labels,omitempty" protobuf:"bytes,11,rep,name=labels"`
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

type LightWeightClient interface{
	GetPod(namespace, podName string) (*Pod, error)
	UpdatePod(namespace string, pod *Pod) error
	PatchPodTag(namespace, podName, tagKey, tagValue string) error
}

type lightWeightClient struct{
	config *Config
}

func (c *lightWeightClient) GetPod(namespace, podName string) (*Pod, error) {
	endpoint := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s", namespace, podName)
	method := http.MethodGet

	req, err := http.NewRequest(method, baseURL+versionedAPIPath+endpoint, nil)
	if err != nil {
		return nil, err
	}
	pod := &Pod{}
	if err := c.do(req, pod); err != nil {
		return nil, err
	}
	return pod, nil
}

// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#patch-pod-v1-core
func (c *lightWeightClient) UpdatePod(namespace string, pod *Pod) error {
	// TODO
	/*
		if _, err := clientSet.CoreV1().Pods(namespace).Update(pod); err != nil {
			return nil, err
		}
	*/
	return nil
}

// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#patch-pod-v1-core
func (c *lightWeightClient) PatchPodTag(namespace, podName, tagKey, tagValue string) error {
	// TODO
	/*
		patch := map[string]string{
			"op":    "add",
			"path":  "/spec/template/metadata/labels/" + key,
			"value": sanitizedDebuggingInfo(value),
		}
		data, err := json.Marshal([]interface{}{patch})
		if err != nil {
			return err
		}
		if _, err := r.clientSet.CoreV1().Pods(r.namespace).Patch(r.podName, types.JSONPatchType, data); err != nil {
			return err
		}
	 */
	return nil
}

func (c *lightWeightClient) do(req *http.Request, ptrToReturnObj interface{}) error {
	// Finish setting up a valid request.
	req.Header.Set("Authorization", "Bearer " + c.config.BearerToken)
	req.Header.Set("Accept", "application/json")
	// TODO what about the TLS CA certificate that was set on the config? And other config opts?

	// Execute it.
	resp, err := cleanhttp.DefaultClient().Do(req)
	if err != nil {
		return err
	}

	// Check for success.
	switch resp.StatusCode {
	case 200, 201, 202:
		// Pass.
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