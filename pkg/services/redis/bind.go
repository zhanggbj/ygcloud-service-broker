package redis

import (
	brokerapi "github.com/pivotal-cf/brokerapi/domain"
)

// Bind implematation
func (r *RedisBroker) Bind(instanceID, bindingID string, details brokerapi.BindDetails, myBool bool) (brokerapi.Binding, error) {
	b := brokerapi.Binding{}
	return b, nil
}
