package openshiftmanager

import (
	"context"
	"fmt"

	"github.com/openshift/hypershift/control-plane-operator/controllers/openshiftmanager/libraryopenshiftmanager"
	"github.com/openshift/library-go/pkg/manifestclient"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/dynamic"
)

var (
	coreConfigMapGVR = corev1.SchemeGroupVersion.WithResource("configmaps")
)

func applyConfigMapOpenshiftAuthenticationConfigSystemCliconfig(applyCtx libraryopenshiftmanager.ApplyOutputResourceContext) error {
	switch applyCtx.ActionType {
	case manifestclient.ActionCreate:
		return applyCreateConfigMapOpenshiftAuthenticationConfigSystemCliconfig(applyCtx.Ctx, applyCtx.MgmtKubeClient, applyCtx.Request, applyCtx.ControlPlaneNamespace)
	}
	return fmt.Errorf("applyConfigMapOpenshiftAuthenticationConfigSystemCliconfig action type: %v not supported", applyCtx.ActionType)
}

// openshift-authentication/v4-0-config-system-cliconfig
//
// more details:
// https://docs.google.com/document/d/1lerQtnLFofoXaO08SX2iYOv0b7pUA0Rn3Q3htxKjnrY/edit?tab=t.0
// Position in the document: 21
func applyCreateConfigMapOpenshiftAuthenticationConfigSystemCliconfig(ctx context.Context, mgmtKubeClient *dynamic.DynamicClient, requestToCreate manifestclient.SerializedRequestish, controlPlaneNamespace string) error {
	// on hcp the oauth-server configuration is stored in
	// controlPlaneNamespace/oauth-openshift configmap under "config.yaml" key
	//
	// xref: https://github.com/openshift/hypershift/blob/675f881923cfa312115ba9bd572f39c201bbe689/control-plane-operator/controllers/hostedcontrolplane/v2/oauth/config.go#L44
	//
	// for OM we are going to create a new configmap in controlPlaneNamespace
	// under "openshift-authentication--v4-0-config-system-cliconfig" name
	//
	// TODO: figure out which transformations are required, for example:
	// xref: https://github.com/openshift/hypershift/blob/675f881923cfa312115ba9bd572f39c201bbe689/control-plane-operator/controllers/hostedcontrolplane/v2/oauth/config.go#L63
	//
	// - namedCerts, xref: https://github.com/openshift/hypershift/blob/675f881923cfa312115ba9bd572f39c201bbe689/control-plane-operator/controllers/hostedcontrolplane/v2/oauth/config.go#L66
	//   in standalone custom certs come from "v4-0-config-system-custom-router-certs" or "v4-0-config-system-router-certs"" secrets
	//   xref: https://github.com/openshift/cluster-authentication-operator/blob/7c29d664bd571ce5f8e99456a206584651d200a7/pkg/controllers/configobservation/routersecret/observe_router_secret.go#L17
	//
	// - login-url-overrides, xref: https://github.com/openshift/hypershift/blob/675f881923cfa312115ba9bd572f39c201bbe689/control-plane-operator/controllers/hostedcontrolplane/v2/oauth/config.go#L79
	//
	// - verify masterURL because it seems to be matching masterPublicURL
	//   xref: https://github.com/openshift/hypershift/blob/675f881923cfa312115ba9bd572f39c201bbe689/control-plane-operator/controllers/hostedcontrolplane/v2/oauth/config.go#L73
	//   on standalone masterURL points to oauth-server service
	//   xref: https://github.com/openshift/cluster-authentication-operator/blob/817783a52d042f4ac3aa8faac7421ac013b42481/pkg/controllers/payload/payload_config_controller.go#L228
	//
	// - any other transformations?
	// TODO: fix me

	unstructuredConfigSystemCliConfig, err := decodeIndividualObj(requestToCreate.GetSerializedRequest().Body)
	if err != nil {
		return err
	}

	if err = applyTransformationsToConfigMapOpenshiftAuthenticationConfigSystemCliconfig(unstructuredConfigSystemCliConfig); err != nil {
		return err
	}

	opts, err := getCreateOptionsFromSerializedRequest(requestToCreate.GetSerializedRequest())
	if err != nil {
		return err
	}

	return createUnstructuredResource(ctx, coreConfigMapGVR, mgmtKubeClient, unstructuredConfigSystemCliConfig, opts, controlPlaneNamespace)
}
