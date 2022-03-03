package redis

import (
	"context"

	brokerapi "github.com/pivotal-cf/brokerapi/domain"
)

func (r *RedisBroker) GetInstance(ctx context.Context, instanceID string) (brokerapi.GetInstanceDetailsSpec, error) {

	return brokerapi.GetInstanceDetailsSpec{}, nil
}
