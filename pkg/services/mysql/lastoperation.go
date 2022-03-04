package mysql

import (
	"context"
	"fmt"

	brokerapi "github.com/pivotal-cf/brokerapi/domain"
)

// LastOperation implematation
func (m *MySqlBroker) LastOperation(ctx context.Context, instanceID string, details brokerapi.PollDetails) (brokerapi.LastOperation, error) {
	// TODO: Handle different cases based on details.OperationData, like Provisioning, updating or deprovisioning
	// OperationProvisioning || OperationUpdating
	instance, err, serviceErr := SyncStatusWithService(m, instanceID, details.ServiceID,
		details.PlanID)

	m.Logger.Debug(fmt.Sprintf("SyncStatusWithService finished"))
	if err != nil {
		return brokerapi.LastOperation{}, err
	}
	if serviceErr != nil {
		return brokerapi.LastOperation{
			State:       brokerapi.Failed,
			Description: fmt.Sprintf("get mysql instance failed. Error: %s", serviceErr),
		}, nil
	}
	// Status, ContainerCreating
	if instance.Status.Phase == "Running" {
		return brokerapi.LastOperation{
			State:       brokerapi.Succeeded,
			Description: fmt.Sprintf("Status: %s", fmt.Sprintf("%v", instance.Status.Phase)),
		}, nil
	} else if instance.Status.Phase == "s" {
		return brokerapi.LastOperation{
			State:       brokerapi.Failed,
			Description: fmt.Sprintf("Status: %s", fmt.Sprintf("%v", instance.Status.Phase)),
		}, nil
	} else {
		return brokerapi.LastOperation{
			State:       brokerapi.InProgress,
			Description: fmt.Sprintf("Status: %s", fmt.Sprintf("%v", instance.Status.Phase)),
		}, nil
	}
}
 