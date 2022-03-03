package redis

import (
	brokerapi "github.com/pivotal-cf/brokerapi/domain"
)

// Deprovision implematation
func (r *RedisBroker) Deprovision(instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (brokerapi.DeprovisionServiceSpec, error) {
	svcSpec := brokerapi.DeprovisionServiceSpec{}

	return svcSpec, nil
}
