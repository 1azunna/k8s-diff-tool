package cluster

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps the dynamic client and REST mapper for interacting with the cluster.
type Client struct {
	dynamicClient dynamic.Interface
	mapper        meta.RESTMapper
	namespace     string // Default namespace from kubeconfig
}

// NewClient creates a new Client using the default kubeconfig loading rules.
func NewClient(kubeContext string) (*Client, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	if kubeContext != "" {
		configOverrides.CurrentContext = kubeContext
	}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("error building kubeconfig: %w", err)
	}

	namespace, _, err := kubeConfig.Namespace()
	if err != nil {
		namespace = "default"
	}

	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating dynamic client: %w", err)
	}

	dc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating discovery client: %w", err)
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	return &Client{
		dynamicClient: dynClient,
		mapper:        mapper,
		namespace:     namespace,
	}, nil
}

// GetResource fetches a resource from the cluster given its GVK, name, and namespace.
// If namespace is empty, it uses the client's default namespace (from context).
func (c *Client) GetResource(apiVersion, kind, name, namespace string) (*unstructured.Unstructured, error) {
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid apiVersion %s: %w", apiVersion, err)
	}
	gvk := gv.WithKind(kind)

	mapping, err := c.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to find GVR for %s: %w", gvk, err)
	}

	targetNamespace := namespace
	if targetNamespace == "" {
		targetNamespace = c.namespace
	}

	var resource *unstructured.Unstructured
	if mapping.Scope.Name() == meta.RESTScopeNameRoot {
		resource, err = c.dynamicClient.Resource(mapping.Resource).Get(context.TODO(), name, metav1.GetOptions{})
	} else {
		resource, err = c.dynamicClient.Resource(mapping.Resource).Namespace(targetNamespace).Get(context.TODO(), name, metav1.GetOptions{})
	}

	if err != nil {
		return nil, err
	}

	return resource, nil
}

// ParseResources parses YAML bytes into a slice of Unstructured objects.
func ParseResources(data []byte) ([]*unstructured.Unstructured, error) {
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 4096)
	var objs []*unstructured.Unstructured

	for {
		var u unstructured.Unstructured
		err := decoder.Decode(&u)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		// Skip empty documents
		if len(u.Object) == 0 {
			continue
		}
		objs = append(objs, &u)
	}
	return objs, nil
}

// ServerSideApplyDryRun performs a server-side apply in dry-run mode to calculate the "future state" of the resource.
func (c *Client) ServerSideApplyDryRun(local *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	gvk := local.GroupVersionKind()
	mapping, err := c.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to map GVK %s: %w", gvk, err)
	}

	namespace := local.GetNamespace()
	if namespace == "" {
		namespace = c.namespace
	}
	name := local.GetName()

	// SSA requires the object to be JSON encoded
	data, err := json.Marshal(local.Object)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal local object: %w", err)
	}

	var drClient dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameRoot {
		drClient = c.dynamicClient.Resource(mapping.Resource)
	} else {
		drClient = c.dynamicClient.Resource(mapping.Resource).Namespace(namespace)
	}

	// Perform the Patch with DryRun
	patchOptions := metav1.PatchOptions{
		FieldManager: "kdiff",
		DryRun:       []string{metav1.DryRunAll},
		Force:        nil, // Set true if needed, but for diffing we usually want to see conflicts? No, we want to see result.
	}
	// Force ownership to simulate "If I applied this, what happens?"
	force := true
	patchOptions.Force = &force

	applied, err := drClient.Patch(context.TODO(), name, types.ApplyPatchType, data, patchOptions)
	if err != nil {
		return nil, fmt.Errorf("dry-run apply failed: %w", err)
	}

	return applied, nil
}
