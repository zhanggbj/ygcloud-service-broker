package redis

import (
	brokerapi "github.com/pivotal-cf/brokerapi/domain"
)

// Unbind implematation
func (r *RedisBroker) Unbind(instanceID, bindingID string, details brokerapi.UnbindDetails, asyncAllowed bool) (brokerapi.UnbindSpec, error) {
	return brokerapi.UnbindSpec{}, nil
}
