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

// fake the content of the operator res for the POC
var operatorAuthenticationsClusterConfigMapJSON = `{
    "apiVersion": "v1",
    "data": {
        "cluster.yaml": "apiVersion: operator.openshift.io/v1\nkind: Authentication\nmetadata:\n  creationTimestamp: \"2025-09-12T08:02:07Z\"\n  generation: 7\n  managedFields:\n  - apiVersion: operator.openshift.io/v1\n    fieldsType: FieldsV1\n    fieldsV1:\n      f:status:\n        f:conditions:\n          k:{\"type\":\"OAuthServerConfigObservationDegraded\"}:\n            .: {}\n            f:lastTransitionTime: {}\n            f:status: {}\n            f:type: {}\n    manager: oauth-server-ConfigObserver\n    operation: Apply\n    subresource: status\n    time: \"2025-09-12T08:02:09Z\"\n  - apiVersion: operator.openshift.io/v1\n    fieldsType: FieldsV1\n    fieldsV1:\n      f:status:\n        f:conditions:\n          k:{\"type\":\"OAuthConfigDegraded\"}:\n            .: {}\n            f:lastTransitionTime: {}\n            f:message: {}\n            f:reason: {}\n            f:status: {}\n            f:type: {}\n          k:{\"type\":\"OAuthConfigIngressDegraded\"}:\n            .: {}\n            f:lastTransitionTime: {}\n            f:message: {}\n            f:reason: {}\n            f:status: {}\n            f:type: {}\n          k:{\"type\":\"OAuthConfigRouteDegraded\"}:\n            .: {}\n            f:lastTransitionTime: {}\n            f:message: {}\n            f:reason: {}\n            f:status: {}\n            f:type: {}\n          k:{\"type\":\"OAuthConfigServiceDegraded\"}:\n            .: {}\n            f:lastTransitionTime: {}\n            f:message: {}\n            f:reason: {}\n            f:status: {}\n            f:type: {}\n          k:{\"type\":\"OAuthSessionSecretDegraded\"}:\n            .: {}\n            f:lastTransitionTime: {}\n            f:message: {}\n            f:reason: {}\n            f:status: {}\n            f:type: {}\n    manager: openshift-authentication-PayloadConfig\n    operation: Apply\n    subresource: status\n    time: \"2025-09-12T08:02:09Z\"\n  - apiVersion: operator.openshift.io/v1\n    fieldsType: FieldsV1\n    fieldsV1:\n      f:status:\n        f:conditions:\n          k:{\"type\":\"OAuthServerDeploymentAvailable\"}:\n            .: {}\n            f:lastTransitionTime: {}\n            f:message: {}\n            f:reason: {}\n            f:status: {}\n            f:type: {}\n          k:{\"type\":\"OAuthServerDeploymentDegraded\"}:\n            .: {}\n            f:lastTransitionTime: {}\n            f:message: {}\n            f:reason: {}\n            f:status: {}\n            f:type: {}\n          k:{\"type\":\"OAuthServerDeploymentProgressing\"}:\n            .: {}\n            f:lastTransitionTime: {}\n            f:message: {}\n            f:reason: {}\n            f:status: {}\n            f:type: {}\n          k:{\"type\":\"OAuthServerWorkloadDegraded\"}:\n            .: {}\n            f:lastTransitionTime: {}\n            f:status: {}\n            f:type: {}\n        f:generations:\n          k:{\"group\":\"apps\",\"name\":\"oauth-openshift\",\"namespace\":\"openshift-authentication\",\"resource\":\"deployments\"}:\n            .: {}\n            f:group: {}\n            f:lastGeneration: {}\n            f:name: {}\n            f:namespace: {}\n            f:resource: {}\n    manager: OAuthServer-Workload\n    operation: Apply\n    subresource: status\n    time: \"2025-09-12T09:28:27Z\"\n  - apiVersion: operator.openshift.io/v1\n    fieldsType: FieldsV1\n    fieldsV1:\n      f:status:\n        f:conditions:\n          k:{\"type\":\"OpenshiftAuthenticationStaticResourcesDegraded\"}:\n            .: {}\n            f:lastTransitionTime: {}\n            f:message: {}\n            f:reason: {}\n            f:status: {}\n            f:type: {}\n    manager: OpenshiftAuthenticationStaticResources-StaticResources\n    operation: Apply\n    subresource: status\n    time: \"2025-09-17T09:15:37Z\"\n  - apiVersion: operator.openshift.io/v1\n    fieldsType: FieldsV1\n    fieldsV1:\n      f:spec:\n        .: {}\n        f:logLevel: {}\n        f:managementState: {}\n        f:observedConfig:\n          .: {}\n          f:oauthServer:\n            .: {}\n            f:corsAllowedOrigins: {}\n            f:oauthConfig:\n              .: {}\n              f:loginURL: {}\n              f:templates:\n                .: {}\n                f:error: {}\n                f:login: {}\n                f:providerSelection: {}\n              f:tokenConfig:\n                .: {}\n                f:accessTokenMaxAgeSeconds: {}\n                f:authorizeTokenMaxAgeSeconds: {}\n            f:serverArguments:\n              .: {}\n              f:audit-log-format: {}\n              f:audit-log-maxbackup: {}\n              f:audit-log-maxsize: {}\n              f:audit-log-path: {}\n              f:audit-policy-file: {}\n            f:servingInfo:\n              .: {}\n              f:cipherSuites: {}\n              f:minTLSVersion: {}\n              f:namedCertificates: {}\n            f:volumesToMount:\n              .: {}\n              f:identityProviders: {}\n        f:operatorLogLevel: {}\n        f:unsupportedConfigOverrides: {}\n    manager: control-plane-operator\n    operation: Update\n    time: \"2025-10-14T11:03:57Z\"\n  name: cluster\n  resourceVersion: \"21523314\"\n  uid: dd62253f-90f4-4780-8d11-d2cabb73479a\nspec:\n  logLevel: Normal\n  managementState: Managed\n  observedConfig:\n    oauthServer:\n      corsAllowedOrigins:\n      - //127\\.0\\.0\\.1(:|$)\n      - //localhost(:|$)\n      oauthConfig:\n        loginURL: https://aeae417d4380a44c89494ab10c571737-29b48f7ab72a7ca9.elb.us-east-1.amazonaws.com:6443\n        templates:\n          error: /var/config/system/secrets/v4-0-config-system-ocp-branding-template/errors.html\n          login: /var/config/system/secrets/v4-0-config-system-ocp-branding-template/login.html\n          providerSelection: /var/config/system/secrets/v4-0-config-system-ocp-branding-template/providers.html\n        tokenConfig:\n          accessTokenMaxAgeSeconds: 86400\n          authorizeTokenMaxAgeSeconds: 300\n      serverArguments:\n        audit-log-format:\n        - json\n        audit-log-maxbackup:\n        - \"10\"\n        audit-log-maxsize:\n        - \"100\"\n        audit-log-path:\n        - /var/log/oauth-server/audit.log\n        audit-policy-file:\n        - /var/run/configmaps/audit/audit.yaml\n      servingInfo:\n        cipherSuites:\n        - TLS_AES_128_GCM_SHA256\n        - TLS_AES_256_GCM_SHA384\n        - TLS_CHACHA20_POLY1305_SHA256\n        - TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256\n        - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256\n        - TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384\n        - TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384\n        - TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256\n        - TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256\n        minTLSVersion: VersionTLS12\n        namedCertificates: []\n      volumesToMount:\n        identityProviders: '{}'\n  operatorLogLevel: Normal\n  unsupportedConfigOverrides: null\nstatus:\n  conditions:\n  - lastTransitionTime: \"2025-09-12T08:02:09Z\"\n    message: \"\"\n    reason: \"\"\n    status: \"False\"\n    type: OAuthConfigDegraded\n  - lastTransitionTime: \"2025-09-12T08:02:09Z\"\n    message: \"\"\n    reason: \"\"\n    status: \"False\"\n    type: OAuthConfigIngressDegraded\n  - lastTransitionTime: \"2025-09-12T08:02:09Z\"\n    message: \"\"\n    reason: \"\"\n    status: \"False\"\n    type: OAuthConfigRouteDegraded\n  - lastTransitionTime: \"2025-09-12T08:02:09Z\"\n    message: \"\"\n    reason: \"\"\n    status: \"False\"\n    type: OAuthConfigServiceDegraded\n  - lastTransitionTime: \"2025-09-12T08:02:09Z\"\n    message: \"\"\n    reason: \"\"\n    status: \"False\"\n    type: OAuthSessionSecretDegraded\n  - lastTransitionTime: \"2025-09-12T08:02:09Z\"\n    status: \"False\"\n    type: OAuthServerConfigObservationDegraded\n  - lastTransitionTime: \"2025-09-12T08:07:11Z\"\n    message: no oauth-openshift.openshift-authentication pods available on any node.\n    reason: NoPod\n    status: \"False\"\n    type: OAuthServerDeploymentAvailable\n  - lastTransitionTime: \"2025-09-12T09:28:26Z\"\n    message: 1 of 1 requested instances are unavailable for oauth-openshift.openshift-authentication\n      (no pods found with labels \"\")\n    reason: UnavailablePod\n    status: \"True\"\n    type: OAuthServerDeploymentDegraded\n  - lastTransitionTime: \"2025-09-12T08:07:11Z\"\n    message: 'deployment/oauth-openshift.openshift-authentication: 0/1 pods have been\n      updated to the latest generation and 0/1 pods are available'\n    reason: PodsUpdating\n    status: \"True\"\n    type: OAuthServerDeploymentProgressing\n  - lastTransitionTime: \"2025-09-12T08:07:11Z\"\n    status: \"False\"\n    type: OAuthServerWorkloadDegraded\n  - lastTransitionTime: \"2025-09-17T09:15:36Z\"\n    message: \"\"\n    reason: AsExpected\n    status: \"False\"\n    type: OpenshiftAuthenticationStaticResourcesDegraded\n  generations:\n  - group: apps\n    lastGeneration: 0\n    name: oauth-openshift\n    namespace: openshift-authentication\n    resource: deployments\n"
    },
    "kind": "ConfigMap",
    "metadata": {
        "creationTimestamp": "2025-09-15T12:20:29Z",
        "name": "operator.openshift.io--authentications--cluster",
        "resourceVersion": "19800004",
        "uid": "30120d6d-3866-4bfa-8264-83e9fa226760"
    }
}
`

func projectOperatorAuthenticationCluster(inputCtx libraryopenshiftmanager.InputResourceGetterContext) (*libraryinputresources.Resource, error) {
	gvr := corev1.SchemeGroupVersion.WithResource("configmaps")
	authOperatorConfigMap, err := decodeIndividualObj([]byte(operatorAuthenticationsClusterConfigMapJSON))
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
