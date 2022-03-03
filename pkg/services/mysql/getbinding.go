package mysql

import (
	"context"

	brokerapi "github.com/pivotal-cf/brokerapi/domain"
)

func (m *MySqlBroker) GetBinding(ctx context.Context, instanceID, bindingID string) (brokerapi.GetBindingSpec, error) {
	bs := brokerapi.GetBindingSpec{}
	return bs, nil
}
