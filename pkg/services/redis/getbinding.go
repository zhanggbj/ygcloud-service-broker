package redis

import (
	"context"

	brokerapi "github.com/pivotal-cf/brokerapi/domain"
)

func (r *RedisBroker) GetBinding(ctx context.Context, instanceID, bindingID string) (brokerapi.GetBindingSpec, error) {
	bs := brokerapi.GetBindingSpec{}
	return bs, nil
}
