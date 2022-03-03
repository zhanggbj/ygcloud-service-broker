package mysql

import (
	brokerapi "github.com/pivotal-cf/brokerapi/domain"
)

// Update implematation
func (m *MySqlBroker) Update(instanceID string, details brokerapi.UpdateDetails, asyncAllowed bool) (brokerapi.UpdateServiceSpec, error) {
	spec := brokerapi.UpdateServiceSpec{}

	return spec, nil
}
