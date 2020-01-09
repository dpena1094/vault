package kubernetes

// TODO make every effort to avoid using client-go.
// TODO
// To put or to patch, that is the question.
// https://kubernetes.io/docs/tasks/run-application/update-api-object-kubectl-patch/

type Pod struct {
	Labels map[string]string
	// TODO more will need to be read in here
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
	// TODO
	/*
		pod, err := clientSet.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
	 */
	return nil, nil
}

func (c *lightWeightClient) UpdatePod(namespace string, pod *Pod) error {
	// TODO
	/*
		if _, err := clientSet.CoreV1().Pods(namespace).Update(pod); err != nil {
			return nil, err
		}
	 */
	return nil
}

func (c *lightWeightClient) PatchPodTag(namespace, podName, tagKey, tagValue string) error {
	// TODO
	/*
		patch := map[string]string{
			"op":    "add",
			"path":  "/spec/template/metadata/labels/" + key,
			"value": toString(value),
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
