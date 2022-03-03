package redis

import (
	"context"

	brokerapi "github.com/pivotal-cf/brokerapi/domain"
)

func (r *RedisBroker) LastBindingOperation(ctx context.Context, instanceID, bindingID string, details brokerapi.PollDetails) (brokerapi.LastOperation, error) {
	return brokerapi.LastOperation{}, nil
}
