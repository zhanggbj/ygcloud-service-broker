package redis

import (
	"context"
	"fmt"

	brokerapi "github.com/pivotal-cf/brokerapi/domain"
)

// LastOperation implematation
func (r *RedisBroker) LastOperation(ctx context.Context, instanceID string, details brokerapi.PollDetails) (brokerapi.LastOperation, error) {
	// TODO: Handle different cases based on details.OperationData, like Provisioning, updating or deprovisioning
	// OperationProvisioning || OperationUpdating
	instance, err, serviceErr := SyncStatusWithService(r, instanceID, details.ServiceID,
		details.PlanID)

	r.Logger.Debug(fmt.Sprintf("SyncStatusWithService finished"))
	if err != nil {
		return brokerapi.LastOperation{}, err
	}
	if serviceErr != nil {
		return brokerapi.LastOperation{
			State:       brokerapi.Failed,
			Description: fmt.Sprintf("get mysql instance failed. Error: %s", serviceErr),
		}, nil
	}
	// Status
	if instance.Info.Status == "deployed" {
		return brokerapi.LastOperation{
			State:       brokerapi.Succeeded,
			Description: fmt.Sprintf("Status: %s", instance.Info.Status),
		}, nil
	} else if instance.Info.Status == "failed" {
		return brokerapi.LastOperation{
			State:       brokerapi.Failed,
			Description: fmt.Sprintf("Status: %s", instance.Info.Status),
		}, nil
	} else {
		return brokerapi.LastOperation{
			State:       brokerapi.InProgress,
			Description: fmt.Sprintf("Status: %s", instance.Info.Status),
		}, nil
	}
}
