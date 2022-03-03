package mysql

import (
	brokerapi "github.com/pivotal-cf/brokerapi/domain"
)

// GetPlanSchemas implematation
func (m *MySqlBroker) GetPlanSchemas(serviceID string, planID string, metadata *brokerapi.ServicePlanMetadata) (*brokerapi.ServiceSchemas, error) {
	planSchema := brokerapi.ServiceSchemas{}
	return &planSchema, nil
}
