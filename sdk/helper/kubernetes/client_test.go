package kubernetes

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
)

const (
	testNamespace = "default"
	testPodname   = "shell-demo"
)

func TestClient(t *testing.T) {
	closeFunc := testServer(t)
	defer closeFunc()

	client, err := NewLightWeightClient()
	if err != nil {
		t.Fatal(err)
	}
	e := &env{
		client: client,
	}
	e.TestGetPod(t)
	e.TestGetPodNotFound(t)
	e.TestUpdatePodTags(t)
	e.TestUpdatePodTagsNotFound(t)
}

type env struct {
	client LightWeightClient
}

func (e *env) TestGetPod(t *testing.T) {
	if err := e.client.GetPod(testNamespace, testPodname); err != nil {
		t.Fatal(err)
	}
}

func (e *env) TestGetPodNotFound(t *testing.T) {
	err := e.client.GetPod(testNamespace, "no-exist")
	if err == nil {
		t.Fatal("expected error because pod is unfound")
	}
	if err != ErrNotFound {
		t.Fatalf("expected %q but received %q", ErrNotFound, err)
	}
}

func (e *env) TestUpdatePodTags(t *testing.T) {
	if err := e.client.UpdatePodTags(testNamespace, testPodname, &Tag{
		Key:   "fizz",
		Value: "buzz",
	}); err != nil {
		t.Fatal(err)
	}
}

func (e *env) TestUpdatePodTagsNotFound(t *testing.T) {
	err := e.client.UpdatePodTags(testNamespace, "no-exist", &Tag{
		Key:   "fizz",
		Value: "buzz",
	})
	if err == nil {
		t.Fatal("expected error because pod is unfound")
	}
	if err != ErrNotFound {
		t.Fatalf("expected %q but received %q", ErrNotFound, err)
	}
}

func testServer(t *testing.T) (closeFunc func()) {
	// Edit the url scheme for our test server, and use our
	// fixtures to supply the token and ca.crt.
	scheme = "http://"
	tokenFile = "fixtures/token"
	rootCAFile = "fixtures/ca.crt"

	// Read in the fixtures we'll use for responses.
	getPodSuccessResp, err := ioutil.ReadFile("fixtures/get-pod-response.json")
	if err != nil {
		t.Fatal(err)
	}
	updatePodTagsSuccessResp, err := ioutil.ReadFile("fixtures/update-pod-tags-response.json")
	if err != nil {
		t.Fatal(err)
	}
	notFoundResp, err := ioutil.ReadFile("fixtures/not-found-response.json")
	if err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		namespace, podName, err := parsePath(r.URL.Path)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte(fmt.Sprintf("unable to parse %s: %s", r.URL.Path, err.Error())))
			return
		}

		switch {
		case namespace != testNamespace, podName != testPodname:
			w.WriteHeader(404)
			w.Write(notFoundResp)
			return
		case r.Method == http.MethodGet:
			w.WriteHeader(200)
			w.Write(getPodSuccessResp)
			return
		case r.Method == http.MethodPatch:
			w.WriteHeader(200)
			w.Write(updatePodTagsSuccessResp)
		default:
			w.WriteHeader(400)
			w.Write([]byte(fmt.Sprintf("unexpected request method: %s", r.Method)))
		}
	}))

	// ts.URL example: http://127.0.0.1:35681
	urlFields := strings.Split(ts.URL, "://")
	if len(urlFields) != 2 {
		t.Fatal("received unexpected test url: " + ts.URL)
	}
	urlFields = strings.Split(urlFields[1], ":")
	if len(urlFields) != 2 {
		t.Fatal("received unexpected test url: " + ts.URL)
	}
	if err := os.Setenv(EnvVarKubernetesServiceHost, urlFields[0]); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv(EnvVarKubernetesServicePort, urlFields[1]); err != nil {
		t.Fatal(err)
	}
	return ts.Close
}

// The path should be formatted like this:
// fmt.Sprintf("/api/v1/namespaces/%s/pods/%s", namespace, podName)
func parsePath(urlPath string) (namespace, podName string, err error) {
	original := urlPath
	podName = path.Base(urlPath)
	urlPath = strings.TrimSuffix(urlPath, "/pods/"+podName)
	namespace = path.Base(urlPath)
	if original != fmt.Sprintf("/api/v1/namespaces/%s/pods/%s", namespace, podName) {
		return "", "", fmt.Errorf("received unexpected path: %s", original)
	}
	return namespace, podName, nil
}
