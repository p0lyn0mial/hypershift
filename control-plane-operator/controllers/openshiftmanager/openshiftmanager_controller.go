package openshiftmanager

import (
	"github.com/go-logr/logr"

	"github.com/openshift/hypershift/control-plane-operator/controllers/openshiftmanager/libraryopenshiftmanager"
	"github.com/openshift/multi-operator-manager/pkg/library/libraryinputresources"

	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Reconciler struct {
	MgmtClusterCfg *rest.Config

	// ControlPlaneNamespace is the namespace on the management cluster where the control plane components run.
	ControlPlaneNamespace string

	// HostedControlPlaneName is the name of the hosted control plane that owns this controller instance.
	HostedControlPlaneName string

	// InputDirectory the directory where the input resources are stored
	InputDirectory string

	// OutputDirectory the directory where the output resources are stored
	OutputDirectory string

	log logr.Logger
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	omController, err := libraryopenshiftmanager.New(mgr.GetLogger(), r.MgmtClusterCfg, r.ControlPlaneNamespace, r.HostedControlPlaneName, r.InputDirectory, r.OutputDirectory)
	if err != nil {
		return err
	}

	omController.RegisterInputResourceGetterFuncOrDie(libraryinputresources.ExactLowLevelOperator("authentications"), projectOperatorAuthenticationCluster)
	omController.RegisterInputResourceGetterFuncOrDie(libraryinputresources.ExactConfigResource("authentications"), projectConfigAuthenticationCluster)
	omController.RegisterInputResourceGetterFuncOrDie(libraryinputresources.ExactResource("config.openshift.io", "v1", "clusterversions", "", "version"), projectConfigClusterVersionCluster)
	omController.RegisterInputResourceGetterFuncOrDie(libraryinputresources.ExactResource("route.openshift.io", "v1", "routes", "openshift-authentication", "oauth-openshift"), getRouteOpenshiftAuthenticationOauthOpenshift)
	omController.RegisterInputResourceGetterFuncOrDie(libraryinputresources.ExactResource("", "v1", "services", "openshift-authentication", "oauth-openshift"), getServiceOpenshiftAuthenticationOauthOpenshift)
	omController.RegisterInputResourceGetterFuncOrDie(libraryinputresources.ExactSecret("openshift-authentication", "v4-0-config-system-session"), getSecretOpenshiftAuthenticationConfigSystemSession)

	return omController.SetupWithManager(mgr)
}
