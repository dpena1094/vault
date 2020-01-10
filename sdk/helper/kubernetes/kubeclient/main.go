package main

// This code builds a minimal binary of the lightweight kubernetes
// client and exposes it for manual testing.
// The intention is that the binary can be built and dropped into
// a Kube environment like this:
// https://kubernetes.io/docs/tasks/debug-application-cluster/get-shell-running-container/
// Then, commands can be run to test its API calls.
// The above commands are intended to be run inside an instance of
// minikube that has been started.
// After building this binary, place it in the container like this:
// $ kubectl cp kubeclient /shell-demo:/
// At first you may get 403's, which can be resolved using this:
// https://github.com/fabric8io/fabric8/issues/6840#issuecomment-307560275
//
// Example calls:
// 		./kubeclient -call='get-pod' -namespace='default' -pod-name='shell-demo'
// 		./kubeclient -call='update-pod-tags' -namespace='default' -pod-name='shell-demo' -tags='fizz:buzz,foo:bar'

import (
	"flag"
	"fmt"
	"strings"

	"github.com/hashicorp/vault/sdk/helper/kubernetes"
)

var callToMake string
var tagsToAdd string
var namespace string
var podName string

func init() {
	flag.StringVar(&callToMake, "call", "", `the call to make: 'get-pod' or 'update-pod-tags'`)
	flag.StringVar(&tagsToAdd, "tags", "", `if call is "update-pod-tags", that tags to update like so: "fizz:buzz,foo:bar"`)
	flag.StringVar(&namespace, "namespace", "", "the namespace to use")
	flag.StringVar(&podName, "pod-name", "", "the pod name to use")
}

func main() {
	flag.Parse()

	client, err := kubernetes.NewLightWeightClient()
	if err != nil {
		panic(err)
	}

	switch callToMake {
	case "get-pod":
		if err := client.GetPod(namespace, podName); err != nil {
			panic(err)
		}
		return
	case "update-pod-tags":
		tagPairs := strings.Split(tagsToAdd, ",")
		var tags []*kubernetes.Tag
		for _, tagPair := range tagPairs {
			fields := strings.Split(tagPair, ":")
			if len(fields) != 2 {
				panic(fmt.Errorf("unable to split %s from tags provided of %s", fields, tagsToAdd))
			}
			tags = append(tags, &kubernetes.Tag{
				Key:   fields[0],
				Value: fields[1],
			})
		}
		if err := client.UpdatePodTags(namespace, podName, tags...); err != nil {
			panic(err)
		}
		return
	default:
		panic(fmt.Errorf(`unsupported call provided: %q`, callToMake))
	}
}
