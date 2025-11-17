package openshiftmanager

import (
	"context"
	"fmt"

	"github.com/openshift/library-go/pkg/manifestclient"
	"github.com/openshift/multi-operator-manager/pkg/library/libraryinputresources"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
)

func getCreateOptionsFromSerializedRequest(request manifestclient.SerializedRequestish) (metav1.CreateOptions, error) {
	opts := &metav1.CreateOptions{}
	err := yaml.Unmarshal(request.GetSerializedRequest().Options, opts)
	if err != nil {
		return metav1.CreateOptions{}, fmt.Errorf("unable to decode options: %w", err)
	}
	return *opts, nil
}

func createUnstructuredResource(ctx context.Context, gvr schema.GroupVersionResource, mgmtKubeClient *dynamic.DynamicClient, unstructuredObject *unstructured.Unstructured, opts metav1.CreateOptions, controlPlaneNamespace string) error {
	unstructuredObject.SetName(hcpNameForNamespacedStandaloneResource(unstructuredObject.GetNamespace(), unstructuredObject.GetName()))
	unstructuredObject.SetNamespace(controlPlaneNamespace)

	_, err := mgmtKubeClient.Resource(gvr).Namespace(controlPlaneNamespace).Create(ctx, unstructuredObject, opts)
	return err
}

func runtimeObjectToInputResource(obj runtime.Object, gvr schema.GroupVersionResource) (*libraryinputresources.Resource, error) {
	rawObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	ret := &libraryinputresources.Resource{
		ResourceType: gvr,
		Content:      &unstructured.Unstructured{Object: rawObj},
	}
	return ret, nil
}

func getResourceToInputResources(ctx context.Context, gvr schema.GroupVersionResource, mgmtKubeClient *dynamic.DynamicClient, controlPlaneNamespace, controlPlaneResourceName, standaloneResourceNamespace, standaloneResourceName string) (*libraryinputresources.Resource, error) {
	unstructuredSecret, err := mgmtKubeClient.Resource(gvr).Namespace(controlPlaneNamespace).Get(ctx, controlPlaneResourceName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	unstructuredSecret.SetNamespace(standaloneResourceNamespace)
	unstructuredSecret.SetName(standaloneResourceName)
	return &libraryinputresources.Resource{
		ResourceType: gvr,
		Content:      unstructuredSecret,
	}, nil
}

// TODO:figure out the best naming scheme for namespaced resources
// operator-name--namespace--name
func hcpNameForNamespacedStandaloneResource(standaloneNamespace, standaloneName string) string {
	return standaloneNamespace + "--" + standaloneName
}

func extractConfigMapDataNoCopy(unstructuredConfigMap *unstructured.Unstructured) (map[string]string, error) {
	configMapRawData, found, err := unstructured.NestedFieldNoCopy(unstructuredConfigMap.Object, "data")
	if err != nil {
		return nil, fmt.Errorf("failed reading .Data field from %s/%s configmap, err: %w", unstructuredConfigMap.GetNamespace(), unstructuredConfigMap.GetName(), err)
	}

	configMapData := map[string]string{}
	if found && configMapRawData != nil {
		var ok bool
		configMapRawMap, ok := configMapRawData.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected type of data in .Data field for %s/%s configmap, expected map[string]interface{}, got: %T", unstructuredConfigMap.GetNamespace(), unstructuredConfigMap.GetName(), configMapRawData)
		}
		for k, v := range configMapRawMap {
			strVal, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("unexpected type stored in %s/%s configmap under %s key, expected string, got: %T", unstructuredConfigMap.GetNamespace(), unstructuredConfigMap.GetName(), k, v)
			}
			configMapData[k] = strVal
		}
	}

	return configMapData, nil
}
