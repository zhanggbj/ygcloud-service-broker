package mysql

import (
	brokerapi "github.com/pivotal-cf/brokerapi/domain"
)

// Deprovision implematation
func (m *MySqlBroker) Deprovision(instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (brokerapi.DeprovisionServiceSpec, error) {
	svcSpec := brokerapi.DeprovisionServiceSpec{}

	return svcSpec, nil
}
