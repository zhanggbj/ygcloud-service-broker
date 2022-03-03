package mysql

import (
	"context"

	brokerapi "github.com/pivotal-cf/brokerapi/domain"
)

func (m *MySqlBroker) GetInstance(ctx context.Context, instanceID string) (brokerapi.GetInstanceDetailsSpec, error) {

	return brokerapi.GetInstanceDetailsSpec{}, nil
}
