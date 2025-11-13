package openshiftmanager

import (
	"fmt"
	"github.com/openshift/hypershift/control-plane-operator/controllers/openshiftmanager/libraryopenshiftmanager"

	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/openshift/multi-operator-manager/pkg/library/libraryinputresources"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	// TODO: figure out the best naming scheme
	// for cluster-wide resources
	// operator.openshift.io--authentications--cluster
	operatorAuthenticationConfigMapName = "operator.openshift.io--authentications--cluster"
)

// operator.openshift.io/authentications/cluster
func projectOperatorAuthenticationCluster(inputCtx libraryopenshiftmanager.InputResourceGetterContext) (*libraryinputresources.Resource, error) {
	gvr := corev1.SchemeGroupVersion.WithResource("configmaps")
	authOperatorConfigMap, err := inputCtx.MgmtKubeClient.Resource(gvr).Namespace(inputCtx.ControlPlaneNamespace).Get(inputCtx.Ctx, operatorAuthenticationConfigMapName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	authOperatorYaml, found, err := unstructured.NestedString(authOperatorConfigMap.Object, "data", "cluster.yaml")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("missing cluster.yaml field in %s/%s configmap", inputCtx.ControlPlaneNamespace, operatorAuthenticationConfigMapName)
	}
	unstructuredAuthOperator, err := decodeIndividualObj([]byte(authOperatorYaml))
	if err != nil {
		return nil, err
	}

	gvr = schema.GroupVersionResource{Group: operatorv1.SchemeGroupVersion.Group, Version: operatorv1.SchemeGroupVersion.Version, Resource: "authentications"}
	ret := &libraryinputresources.Resource{
		ResourceType: gvr,
		Content:      unstructuredAuthOperator,
	}
	return ret, nil
}
