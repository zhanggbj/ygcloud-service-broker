package mysql

import (
	brokerapi "github.com/pivotal-cf/brokerapi/domain"
)

// Bind implematation
func (m *MySqlBroker) Bind(instanceID, bindingID string, details brokerapi.BindDetails, myBool bool) (brokerapi.Binding, error) {
	b := brokerapi.Binding{}
	return b, nil
}
