package mysql

import (
	"context"

	brokerapi "github.com/pivotal-cf/brokerapi/domain"
)

// LastOperation implematation
func (m *MySqlBroker) LastOperation(ctx context.Context, instanceID string, details brokerapi.PollDetails) (brokerapi.LastOperation, error) {
	op := brokerapi.LastOperation{}
	return op, nil
}
