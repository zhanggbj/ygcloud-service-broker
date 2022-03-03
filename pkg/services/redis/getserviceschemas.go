package redis

import (
	brokerapi "github.com/pivotal-cf/brokerapi/domain"
)

// GetPlanSchemas implematation
func (r *RedisBroker) GetPlanSchemas(serviceID string, planID string, metadata *brokerapi.ServicePlanMetadata) (*brokerapi.ServiceSchemas, error) {
	planSchema := brokerapi.ServiceSchemas{}
	return &planSchema, nil
}
