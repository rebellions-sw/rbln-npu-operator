package conditions

import (
	"context"
)

type ConditionUpdater interface {
	SetConditionsReady(ctx context.Context, cr any, reason, message string) error
	SetConditionsError(ctx context.Context, cr any, reason, message string) error
}
