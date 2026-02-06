package scope

import (
	"context"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	rebellionsaiv1alpha1 "github.com/rebellions-sw/rbln-npu-operator/api/v1alpha1"
	rblnv1beta1 "github.com/rebellions-sw/rbln-npu-operator/api/v1beta1"
	"github.com/rebellions-sw/rbln-npu-operator/internal/scope/patch"
)

type RBLNDriverScope struct {
	client client.Client

	ctx              context.Context
	log              logr.Logger
	scheme           *runtime.Scheme
	singleton        *rebellionsaiv1alpha1.RBLNDriver
	namespace        string
	openshiftVersion string

	patcher []patch.DriverPatcher
}

func NewRBLNDriverScope(
	ctx context.Context,
	client client.Client,
	log logr.Logger,
	scheme *runtime.Scheme,
	driver *rebellionsaiv1alpha1.RBLNDriver,
	clusterPolicy *rblnv1beta1.RBLNClusterPolicy,
	openshiftVersion string,
) (*RBLNDriverScope, error) {
	s := &RBLNDriverScope{
		client:           client,
		ctx:              ctx,
		log:              log,
		scheme:           scheme,
		singleton:        driver,
		openshiftVersion: openshiftVersion,
	}

	if clusterPolicy != nil && clusterPolicy.Spec.Namespace != "" {
		s.namespace = clusterPolicy.Spec.Namespace
	} else {
		s.namespace = os.Getenv("OPERATOR_NAMESPACE")
	}
	if s.namespace == "" {
		s.namespace = driver.Namespace
	}
	if s.namespace == "" {
		err := fmt.Errorf("namespace is not configured. Set OPERATOR_NAMESPACE env variable or namespace spec")
		s.log.Error(err, "namespace configuration error")
		return nil, err
	}

	dmp, err := patch.NewDriverManagerPatcher(client, log, s.namespace, driver, scheme, s.openshiftVersion)
	if err != nil {
		return s, err
	}
	s.patcher = append(s.patcher, dmp)

	return s, nil
}

func (s *RBLNDriverScope) PatchComponents(ctx context.Context) error {
	for _, p := range s.patcher {
		if p.IsEnabled() {
			if err := p.Patch(ctx, s.singleton); err != nil {
				return fmt.Errorf("failed to patch component: %v", err)
			}
		} else {
			if err := p.CleanUp(ctx, s.singleton); err != nil {
				return fmt.Errorf("failed to clean up component: %v", err)
			}
		}
	}
	return nil
}
