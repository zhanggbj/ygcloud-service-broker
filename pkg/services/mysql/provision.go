package mysql

import (
	brokerapi "github.com/pivotal-cf/brokerapi/domain"
)

// Provision implematation
func (m *MySqlBroker) Provision(instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool) (brokerapi.ProvisionedServiceSpec, error) {
	svcSpec := brokerapi.ProvisionedServiceSpec{}
	return svcSpec, nil
}
