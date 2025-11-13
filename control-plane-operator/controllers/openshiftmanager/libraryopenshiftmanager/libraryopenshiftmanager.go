package libraryopenshiftmanager

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	hypershiftv1beta1 "github.com/openshift/hypershift/api/hypershift/v1beta1"
	hyperclient "github.com/openshift/hypershift/client/clientset/clientset"

	"github.com/openshift/multi-operator-manager/pkg/library/libraryinputresources"
	"github.com/openshift/multi-operator-manager/pkg/library/libraryoutputresources"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/go-logr/logr"
)

type InputResourceGetterContext struct {
	Ctx                   context.Context
	MgmtKubeClient        *dynamic.DynamicClient
	ControlPlaneNamespace string
	HostedControlPlane    *hypershiftv1beta1.HostedControlPlane
}

type InputResourceGetterFunc func(inputCtx InputResourceGetterContext) (*libraryinputresources.Resource, error)

type ApplyOutputResourceContext struct {
}

type ApplyOutputResourceFunc func(outputCtx ApplyOutputResourceContext) error

type Controller struct {
	mgmtKubeClient *dynamic.DynamicClient

	mgmtHcpClient *hyperclient.Clientset

	// controlPlaneNamespace is the namespace on the management cluster where the control plane components run.
	controlPlaneNamespace string

	// hostedControlPlaneName is the name of the hosted control plane that owns this controller instance.
	hostedControlPlaneName string

	// inputDirectory the directory where the input resources are stored
	inputDirectory string

	// outputDirectory the directory where the output resources are stored
	outputDirectory string

	inputResourceRegistry  map[libraryinputresources.ExactResourceID]InputResourceGetterFunc
	outputResourceRegistry map[libraryoutputresources.ExactResourceID]ApplyOutputResourceFunc

	log logr.Logger
}

func New(log logr.Logger, mgmtClusterCfg *rest.Config, controlPlaneNamespace string, hostedControlPlaneName string, inputDirectory string, outputDirectory string) (*Controller, error) {
	mgmtHcpClient, err := hyperclient.NewForConfig(mgmtClusterCfg)
	if err != nil {
		return nil, err
	}

	// HCP uses controller-runtime but for experimentation
	// we are going to use the dynamic client
	//
	// TODO: use controller-runtime
	mgmtKubeClient, err := dynamic.NewForConfig(mgmtClusterCfg)
	if err != nil {
		return nil, err
	}

	return &Controller{
		mgmtKubeClient: mgmtKubeClient,
		mgmtHcpClient:  mgmtHcpClient,

		controlPlaneNamespace:  controlPlaneNamespace,
		hostedControlPlaneName: hostedControlPlaneName,
		inputDirectory:         inputDirectory,
		outputDirectory:        outputDirectory,

		inputResourceRegistry:  map[libraryinputresources.ExactResourceID]InputResourceGetterFunc{},
		outputResourceRegistry: map[libraryoutputresources.ExactResourceID]ApplyOutputResourceFunc{},

		log: log.WithName("library-openshift-manager"),
	}, nil
}

func (c *Controller) RegisterInputResourceGetterFuncOrDie(exactResourceID libraryinputresources.ExactResourceID, inputResourceGetter InputResourceGetterFunc) {
	if _, ok := c.inputResourceRegistry[exactResourceID]; ok {
		panic(fmt.Errorf("InputResourceGetterFunc for %q resource already registered", exactResourceID))
	}
	c.inputResourceRegistry[exactResourceID] = inputResourceGetter
}

func (c *Controller) RegisterApplyOutputResourceFuncOrDie(exactResourceID libraryoutputresources.ExactResourceID, applyOutputResource ApplyOutputResourceFunc) {
	if _, ok := c.outputResourceRegistry[exactResourceID]; ok {
		panic(fmt.Errorf("ApplyOutputResourceFunc for %q resource already registered", exactResourceID))
	}
	c.outputResourceRegistry[exactResourceID] = applyOutputResource
}

func (c *Controller) SetupWithManager(mgr ctrl.Manager) error {
	return mgr.Add(c)
}

func (c *Controller) Start(ctx context.Context) error {
	c.log.Info("Starting OpenshiftManager Controller")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			c.log.Info("Running OpenshiftManager Sync method")
			err := c.runInternal(ctx)
			if err != nil {
				c.log.Error(err, "Failed to run OpenshiftManager Sync method")
			}
		}
	}
}

func (c *Controller) runInternal(ctx context.Context) error {
	actualResources, err := c.getAuthOperatorRequiredInputResourcesFromCluster(ctx, getAuthOperatorStaticInputResources())
	if err != nil {
		return err
	}
	if err = c.writeRequiredInputResources(actualResources, c.inputDirectory); err != nil {
		return err
	}
	return nil

	// TODO
	/*outputResourcesGetter, err := c.execAuthOperatorApplyConfigurationCommand(ctx)
	if err != nil {
		return err
	}
	return c.applyAuthOperatorOutputResources(ctx, mgmtKubeClient, guestClusterKubeClient, outputResourcesGetter)*/
}

func (c *Controller) getAuthOperatorRequiredInputResourcesFromCluster(ctx context.Context, requiredInputResources *libraryinputresources.InputResources) ([]*libraryinputresources.Resource, error) {
	ret, err := c.getAuthOperatorRequiredInputResourcesForResourceList(ctx, requiredInputResources.ApplyConfigurationResources)
	if err != nil {
		return nil, err
	}

	return unstructuredToMustGatherFormat(ret)
}

// for POC we read data directly from the cluster, normally we would use the cache
// TODO: read resources from the cache
func (c *Controller) getAuthOperatorRequiredInputResourcesForResourceList(ctx context.Context, resourceList libraryinputresources.ResourceList) ([]*libraryinputresources.Resource, error) {
	hostedControlPlane, err := c.mgmtHcpClient.HypershiftV1beta1().HostedControlPlanes(c.controlPlaneNamespace).Get(ctx, c.hostedControlPlaneName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	ret := libraryinputresources.NewUniqueResourceSet()
	errs := []error{}

	handleResourceInstanceAndErrorFn := func(resourceInstance *libraryinputresources.Resource, err error) {
		if apierrors.IsNotFound(err) {
			return
		}
		if err != nil {
			errs = append(errs, err)
			return
		}
		ret.Insert(resourceInstance)
	}

	// for the POC, we only need to take into account the ExactResources
	// TODO:add support for other types
	for _, currResource := range resourceList.ExactResources {
		inputResourceGetterFunc, ok := c.inputResourceRegistry[currResource]
		if !ok {
			errs = append(errs, fmt.Errorf("unable to gather an unknown (not implemented ?) input resource %s", currResource))
			continue
		}
		inputCtx := InputResourceGetterContext{
			Ctx:                   ctx,
			MgmtKubeClient:        c.mgmtKubeClient,
			ControlPlaneNamespace: c.controlPlaneNamespace,
			HostedControlPlane:    hostedControlPlane,
		}
		handleResourceInstanceAndErrorFn(inputResourceGetterFunc(inputCtx))
	}

	for _, currResource := range resourceList.GeneratedNameResources {
		c.log.Info("WARNING: skipping reconciling GeneratedNameResources", "generatedResourceID", currResource)
	}
	for _, currResource := range resourceList.LabelSelectedResources {
		c.log.Info("WARNING: skipping reconciling LabelSelectedResources", "labelSelectedResource", currResource)
	}

	return ret.List(), errors.Join(errs...)
}

func (c *Controller) writeRequiredInputResources(actualResources []*libraryinputresources.Resource, targetDir string) error {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("unable to create %q: %w", targetDir, err)
	}

	errs := []error{}
	for _, currResource := range actualResources {
		if err := libraryinputresources.WriteResource(currResource, targetDir); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// normally, we would get that list by running the input-res command of the auth-operator.
// for now, just return a static list required to run controllers that manage the oauth-server.
//
// TODO: get the list directly from the auth-operator
func getAuthOperatorStaticInputResources() *libraryinputresources.InputResources {
	return &libraryinputresources.InputResources{
		ApplyConfigurationResources: libraryinputresources.ResourceList{
			ExactResources: []libraryinputresources.ExactResourceID{
				// operator.openshift.io
				libraryinputresources.ExactLowLevelOperator("authentications"),
				// config.openshift.io
				libraryinputresources.ExactConfigResource("authentications"),
				libraryinputresources.ExactResource("config.openshift.io", "v1", "clusterversions", "", "version"),
				// oauth-openshift (aka. oauth-server)
				libraryinputresources.ExactResource("route.openshift.io", "v1", "routes", "openshift-authentication", "oauth-openshift"),
				libraryinputresources.ExactResource("", "v1", "services", "openshift-authentication", "oauth-openshift"),
				libraryinputresources.ExactSecret("openshift-authentication", "v4-0-config-system-session"),
			},
		},
	}
}
