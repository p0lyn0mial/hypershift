package openshiftmanager

import (
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	hyperv1 "github.com/openshift/hypershift/api/hypershift/v1beta1"
	"github.com/openshift/hypershift/control-plane-operator/controllers/openshiftmanager/libraryopenshiftmanager"
	"github.com/openshift/hypershift/support/capabilities"
	"github.com/openshift/hypershift/support/util"
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

var (
	coreSecretGVR = corev1.SchemeGroupVersion.WithResource("secrets")
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

// config.openshift.io/authentications/cluster resource doesn't exist in HCP we need to project it from HostedControlPlane
func projectConfigAuthenticationCluster(inputCtx libraryopenshiftmanager.InputResourceGetterContext) (*libraryinputresources.Resource, error) {
	cfg := &configv1.Authentication{
		TypeMeta: metav1.TypeMeta{
			APIVersion: configv1.SchemeGroupVersion.String(),
			Kind:       "Authentication",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
	}
	if inputCtx.HostedControlPlane != nil && inputCtx.HostedControlPlane.Spec.Configuration != nil && inputCtx.HostedControlPlane.Spec.Configuration.Authentication != nil {
		cfg.Spec = *inputCtx.HostedControlPlane.Spec.Configuration.Authentication
	}

	return runtimeObjectToInputResource(cfg, configv1.SchemeGroupVersion.WithResource("authentications"))
}

// config.openshift.io/clusterversions/cluster resource doesn't exist in HCP we need to project it from HostedControlPlane
func projectConfigClusterVersionCluster(inputCtx libraryopenshiftmanager.InputResourceGetterContext) (*libraryinputresources.Resource, error) {
	clusterVersion := &configv1.ClusterVersion{
		TypeMeta: metav1.TypeMeta{
			APIVersion: configv1.SchemeGroupVersion.String(),
			Kind:       "ClusterVersion",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "version"},
		Spec: configv1.ClusterVersionSpec{
			ClusterID: configv1.ClusterID(inputCtx.HostedControlPlane.Spec.ClusterID),
			Capabilities: &configv1.ClusterVersionCapabilitiesSpec{
				BaselineCapabilitySet:         configv1.ClusterVersionCapabilitySetNone,
				AdditionalEnabledCapabilities: capabilities.CalculateEnabledCapabilities(inputCtx.HostedControlPlane.Spec.Capabilities),
			},
			Upstream: inputCtx.HostedControlPlane.Spec.UpdateService,
			Channel:  inputCtx.HostedControlPlane.Spec.Channel,
		},
	}

	return runtimeObjectToInputResource(clusterVersion, configv1.SchemeGroupVersion.WithResource("clusterversions"))
}

// openshift-authentication/v4-0-config-system-cliconfig
func getConfigMapOpenshiftAuthenticationConfigSystemCliconfig(inputCtx libraryopenshiftmanager.InputResourceGetterContext) (*libraryinputresources.Resource, error) {
	standaloneResourceNamespace := "openshift-authentication"
	standaloneResourceName := "v4-0-config-system-cliconfig"
	res, err := getResourceToInputResources(inputCtx.Ctx, coreConfigMapGVR, inputCtx.MgmtKubeClient, inputCtx.ControlPlaneNamespace, hcpNameForNamespacedStandaloneResource(standaloneResourceNamespace, standaloneResourceName), standaloneResourceNamespace, standaloneResourceName)
	if err != nil {
		return nil, err
	}
	if err = revertTransformationsToConfigMapOpenshiftAuthenticationConfigSystemCliconfig(res.Content); err != nil {
		return nil, err
	}
	return res, nil
}

// openshift-authentication/oauth-openshift
func getRouteOpenshiftAuthenticationOauthOpenshift(inputCtx libraryopenshiftmanager.InputResourceGetterContext) (*libraryinputresources.Resource, error) {
	// TODO: figure out how to reconcile route on HCP
	//
	// atm reconciled in https://github.com/openshift/hypershift/blob/6b4d6324de66b9aabdbe7be434b28a17c900074b/control-plane-operator/controllers/hostedcontrolplane/hostedcontrolplane_controller.go#L1305
	//
	// route.Spec.Host is used to populate osinv1.OAuthConfig.MasterPublicURL
	// xref: https://github.com/openshift/cluster-authentication-operator/blob/817783a52d042f4ac3aa8faac7421ac013b42481/pkg/controllers/payload/payload_config_controller.go#L178
	//
	// For the POC we wil keep it simple and assume
	// ony one type of route
	//
	// TODO: production code will need cover all cases
	// https://github.com/openshift/hypershift/blob/6b4d6324de66b9aabdbe7be434b28a17c900074b/control-plane-operator/controllers/hostedcontrolplane/hostedcontrolplane_controller.go#L1646
	//
	// I think that on standalone route is managed by the openshift-router operator
	serviceStrategy := util.ServicePublishingStrategyByTypeForHCP(inputCtx.HostedControlPlane, hyperv1.OAuthServer)
	if serviceStrategy == nil {
		return nil, fmt.Errorf("OAuth strategy not specified")
	}
	if serviceStrategy.Type != hyperv1.Route {
		return nil, fmt.Errorf("unsupported (not implemented) service publishing strategy type: %v", serviceStrategy.Type)
	}
	if !util.IsPublicHCP(inputCtx.HostedControlPlane) {
		return nil, fmt.Errorf("unsupported (not implemented) publishing scope of cluster endpoints for: %s", inputCtx.HostedControlPlane.Name)
	}
	gvr := schema.GroupVersionResource{Group: "route.openshift.io", Version: "v1", Resource: "routes"}

	// TODO: export the route name
	// xref: https://github.com/openshift/hypershift/blob/8be1d9c6f8f79106444e48f2b7d0069b942ba0d7/control-plane-operator/controllers/hostedcontrolplane/manifests/infra.go#L104
	route, err := inputCtx.MgmtKubeClient.Resource(gvr).Namespace(inputCtx.HostedControlPlane.Namespace).Get(inputCtx.Ctx, "oauth", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// we need to change the name and the namespace
	// so that the operator can find the resource
	//
	// TODO: should we record the orig name and namespace ?
	route.SetNamespace("openshift-authentication")
	route.SetName("oauth-openshift")
	return &libraryinputresources.Resource{
		ResourceType: gvr,
		Content:      route,
	}, nil
}

// openshift-authentication/oauth-openshift
func getServiceOpenshiftAuthenticationOauthOpenshift(inputCtx libraryopenshiftmanager.InputResourceGetterContext) (*libraryinputresources.Resource, error) {
	// openshift-authentication/oauth-openshift service
	// is reconciled in https://github.com/openshift/hypershift/blob/6b4d6324de66b9aabdbe7be434b28a17c900074b/control-plane-operator/controllers/hostedcontrolplane/hostedcontrolplane_controller.go#L1305
	//
	// for the POC we simply assume the reconciler runs,
	// and we can read the service manifest
	// TODO: fix me (figure out how to reconcile service on HCP)

	serviceStrategy := util.ServicePublishingStrategyByTypeForHCP(inputCtx.HostedControlPlane, hyperv1.OAuthServer)
	if serviceStrategy == nil {
		return nil, fmt.Errorf("OAuth strategy not specified")
	}

	gvr := corev1.SchemeGroupVersion.WithResource("services")
	svc, err := inputCtx.MgmtKubeClient.Resource(gvr).Namespace(inputCtx.HostedControlPlane.Namespace).Get(inputCtx.Ctx, "oauth-openshift", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// we need to change the name and the namespace
	// so that the operator can find the resource
	//
	// TODO: should we record the orig name and namespace ?
	svc.SetNamespace("openshift-authentication")
	svc.SetName("oauth-openshift")
	return &libraryinputresources.Resource{
		ResourceType: gvr,
		Content:      svc,
	}, nil
}

// openshift-authentication/v4-0-config-system-session
func getSecretOpenshiftAuthenticationConfigSystemSession(inputCtx libraryopenshiftmanager.InputResourceGetterContext) (*libraryinputresources.Resource, error) {
	standaloneResourceNamespace := "openshift-authentication"
	standaloneResourceName := "v4-0-config-system-session"
	return getResourceToInputResources(inputCtx.Ctx, coreSecretGVR, inputCtx.MgmtKubeClient, inputCtx.ControlPlaneNamespace, hcpNameForNamespacedStandaloneResource(standaloneResourceNamespace, standaloneResourceName), standaloneResourceNamespace, standaloneResourceName)
}
