package models

import (
	"context"

	brokerapi "github.com/pivotal-cf/brokerapi/domain"
)

// ServiceBrokerProxy is used to implement details
type ServiceBrokerProxy interface {
	Provision(instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool) (brokerapi.ProvisionedServiceSpec, error)

	Deprovision(instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (brokerapi.DeprovisionServiceSpec, error)

	Bind(instanceID, bindingID string, details brokerapi.BindDetails, myBool bool) (brokerapi.Binding, error)

	Unbind(instanceID, bindingID string, details brokerapi.UnbindDetails, asyncAllowed bool) (brokerapi.UnbindSpec, error)

	Update(instanceID string, details brokerapi.UpdateDetails, asyncAllowed bool) (brokerapi.UpdateServiceSpec, error)

	LastOperation(ctx context.Context, instanceID string, details brokerapi.PollDetails) (brokerapi.LastOperation, error)

	GetPlanSchemas(serviceID string, planID string, metadata *brokerapi.ServicePlanMetadata) (*brokerapi.ServiceSchemas, error)

	GetBinding(ctx context.Context, instanceID, bindingID string) (brokerapi.GetBindingSpec, error)

	GetInstance(ctx context.Context, instanceID string) (brokerapi.GetInstanceDetailsSpec, error)

	LastBindingOperation(ctx context.Context, instanceID, bindingID string, details brokerapi.PollDetails) (brokerapi.LastOperation, error)
}
