package conditions

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	rblnv1beta1 "github.com/rebellions-sw/rbln-npu-operator/api/v1beta1"
)

type clusterPolicyConditionManager struct {
	client client.Client
}

func NewClusterPolicyConditionMgr(client client.Client) ConditionUpdater {
	return &clusterPolicyConditionManager{
		client: client,
	}
}

func (u *clusterPolicyConditionManager) SetConditionsReady(ctx context.Context, cr any, reason, message string) error {
	clusterPolicyCr, err := asClusterPolicy(cr)
	if err != nil {
		return err
	}
	return u.setConditions(ctx, clusterPolicyCr, ConditionReady, reason, message)
}

func (u *clusterPolicyConditionManager) SetConditionsError(ctx context.Context, cr any, reason, message string) error {
	clusterPolicyCr, err := asClusterPolicy(cr)
	if err != nil {
		return err
	}
	return u.setConditions(ctx, clusterPolicyCr, ConditionError, reason, message)
}

func asClusterPolicy(obj any) (*rblnv1beta1.RBLNClusterPolicy, error) {
	clusterPolicy, ok := obj.(*rblnv1beta1.RBLNClusterPolicy)
	if !ok {
		return nil, fmt.Errorf("provided object is not a *rblnv1beta1.RBLNClusterPolicy")
	}
	return clusterPolicy, nil
}

func (u *clusterPolicyConditionManager) setConditions(ctx context.Context, cr *rblnv1beta1.RBLNClusterPolicy, statusType, reason, message string) error {
	switch statusType {
	case ConditionReady:
		meta.SetStatusCondition(&cr.Status.Conditions, metav1.Condition{
			Type:    ConditionReady,
			Status:  metav1.ConditionTrue,
			Reason:  reason,
			Message: message,
		})

		meta.SetStatusCondition(&cr.Status.Conditions, metav1.Condition{
			Type:   ConditionError,
			Status: metav1.ConditionFalse,
			Reason: ConditionReady,
		})
	case ConditionError:
		meta.SetStatusCondition(&cr.Status.Conditions, metav1.Condition{
			Type:   ConditionReady,
			Status: metav1.ConditionFalse,
			Reason: ConditionError,
		})

		meta.SetStatusCondition(&cr.Status.Conditions, metav1.Condition{
			Type:    ConditionError,
			Status:  metav1.ConditionTrue,
			Reason:  reason,
			Message: message,
		})
	default:
		return fmt.Errorf("unknown status type provided: %s", statusType)
	}

	return u.client.Status().Update(ctx, cr)
}
