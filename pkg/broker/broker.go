package broker

import (
	"context"
	"encoding/json"
	"fmt"

	"code.cloudfoundry.org/lager"
	"github.com/pivotal-cf/brokerapi/domain"
	brokerapi "github.com/pivotal-cf/brokerapi/domain"
	"github.com/zhanggbj/ygcloud-service-broker/pkg/config"
	"github.com/zhanggbj/ygcloud-service-broker/pkg/database"
	"github.com/zhanggbj/ygcloud-service-broker/pkg/models"
	"github.com/zhanggbj/ygcloud-service-broker/pkg/services/mysql"
	"github.com/zhanggbj/ygcloud-service-broker/pkg/services/redis"
)

// CloudServiceBroker define
type CloudServiceBroker struct {
	CloudCredentials config.CloudCredentials
	Catalog          config.Catalog
	ServiceBrokerMap map[string]models.ServiceBrokerProxy
	Logger           lager.Logger
}

// New returns a composed service broker object
func New(logger lager.Logger, config config.Config) (*CloudServiceBroker, error) {

	self := CloudServiceBroker{}
	self.CloudCredentials = config.CloudCredentials
	self.Catalog = config.Catalog
	self.Logger = logger

	// map service specific brokers to general broker
	self.ServiceBrokerMap = map[string]models.ServiceBrokerProxy{
		// mysql
		models.MysqlServiceName: &mysql.MySqlBroker{
			CloudCredentials: self.CloudCredentials,
			Catalog:          self.Catalog,
			Logger:           self.Logger,
		},
		// redis
		models.RedisServiceName: &redis.RedisBroker{
			CloudCredentials: self.CloudCredentials,
			Catalog:          self.Catalog,
			Logger:           self.Logger,
		},
	}

	// replace the mapping from name to a mapping from id
	for _, service := range self.Catalog.Services {
		self.ServiceBrokerMap[service.ID] = self.ServiceBrokerMap[service.Name]
		delete(self.ServiceBrokerMap, service.Name)
	}

	// fill with plan schemas
	for i := range self.Catalog.Services {
		// get detail service broker proxy from ServiceBrokerMap
		self.Logger.Debug(fmt.Sprintf("Catalog Services ID: %s. Name: %s",
			self.Catalog.Services[i].ID, self.Catalog.Services[i].Name))
		servicebrokerproxy := self.ServiceBrokerMap[self.Catalog.Services[i].ID]
		if servicebrokerproxy == nil {
			return &self, fmt.Errorf("could not fill with service broker: %s", self.Catalog.Services[i].ID)
		}
		for j := range self.Catalog.Services[i].Plans {
			schemas, err := servicebrokerproxy.GetPlanSchemas(
				self.Catalog.Services[i].ID,
				self.Catalog.Services[i].Plans[j].ID,
				self.Catalog.Services[i].Plans[j].Metadata)
			if err != nil {
				return &self, err
			}
			self.Catalog.Services[i].Plans[j].Schemas = schemas
		}
	}

	return &self, nil
}

// Services lists services in this cloud broker
func (cloudBroker *CloudServiceBroker) Services(
	ctx context.Context) ([]brokerapi.Service, error) {

	cloudBroker.Logger.Debug("Services received.")

	// reuturn service in catalog
	return cloudBroker.Catalog.Services, nil
}

// Provision creates a service instance
func (cloudBroker *CloudServiceBroker) Provision(
	ctx context.Context,
	instanceID string,
	details brokerapi.ProvisionDetails,
	asyncAllowed bool) (brokerapi.ProvisionedServiceSpec, error) {

	cloudBroker.Logger.Debug(fmt.Sprintf("Provision received. instanceID: %s", instanceID))

	error := cloudBroker.Catalog.ValidateOrgSpecGUID(details.OrganizationGUID, details.SpaceGUID)
	if error != nil {
		return brokerapi.ProvisionedServiceSpec{}, error
	}

	if details.ServiceID == "" || details.PlanID == "" {
		e := cloudBroker.Catalog.ValidateAcceptsIncomplete(asyncAllowed)
		if e != nil {
			return brokerapi.ProvisionedServiceSpec{}, e
		}
	}

	// find service plan
	_, err := cloudBroker.Catalog.FindServicePlan(details.ServiceID, details.PlanID)
	if err != nil {
		return brokerapi.ProvisionedServiceSpec{}, err
	}

	// get detail service broker proxy from ServiceBrokerMap
	servicebrokerproxy := cloudBroker.ServiceBrokerMap[details.ServiceID]
	if servicebrokerproxy == nil {
		return brokerapi.ProvisionedServiceSpec{}, fmt.Errorf("could not find service broker: %s", details.ServiceID)
	}

	// detail service broker proxy provision
	return servicebrokerproxy.Provision(instanceID, details, asyncAllowed)
}

// Deprovision deletes the given instance
func (cloudBroker *CloudServiceBroker) Deprovision(
	ctx context.Context,
	instanceID string,
	details brokerapi.DeprovisionDetails,
	asyncAllowed bool) (brokerapi.DeprovisionServiceSpec, error) {

	cloudBroker.Logger.Debug(fmt.Sprintf("Deprovision received. instanceID: %s", instanceID))

	if details.ServiceID == "" || details.PlanID == "" {
		e := cloudBroker.Catalog.ValidateAcceptsIncomplete(asyncAllowed)
		if e != nil {
			return brokerapi.DeprovisionServiceSpec{}, e
		}
	}

	// find service plan
	_, err := cloudBroker.Catalog.FindServicePlan(details.ServiceID, details.PlanID)
	if err != nil {
		return brokerapi.DeprovisionServiceSpec{}, err
	}

	// get detail service broker proxy from ServiceBrokerMap
	servicebrokerproxy := cloudBroker.ServiceBrokerMap[details.ServiceID]
	if servicebrokerproxy == nil {
		return brokerapi.DeprovisionServiceSpec{}, fmt.Errorf("could not find service broker: %s", details.ServiceID)
	}

	// detail service broker proxy deprovision
	return servicebrokerproxy.Deprovision(instanceID, details, asyncAllowed)
}

// Bind adds and returns the associated credentials
func (cloudBroker *CloudServiceBroker) Bind(
	ctx context.Context,
	instanceID string,
	bindingID string,
	details brokerapi.BindDetails,
	myBool bool) (brokerapi.Binding, error) {

	cloudBroker.Logger.Debug(fmt.Sprintf("Bind received. instanceID: %s. bindingID: %s", instanceID, bindingID))

	// find service plan
	_, err := cloudBroker.Catalog.FindServicePlan(details.ServiceID, details.PlanID)
	if err != nil {
		return brokerapi.Binding{}, err
	}

	// get detail service broker proxy from ServiceBrokerMap
	servicebrokerproxy := cloudBroker.ServiceBrokerMap[details.ServiceID]
	if servicebrokerproxy == nil {
		return brokerapi.Binding{}, fmt.Errorf("could not find service broker: %s", details.ServiceID)
	}

	// detail service broker proxy bind
	return servicebrokerproxy.Bind(instanceID, bindingID, details, true)
}

// Unbind removes the associated credentials
func (cloudBroker *CloudServiceBroker) Unbind(
	ctx context.Context,
	instanceID string,
	bindingID string,
	details brokerapi.UnbindDetails, asyncAllowed bool) (brokerapi.UnbindSpec, error) {

	cloudBroker.Logger.Debug(fmt.Sprintf("Unbind received. instanceID: %s. bindingID: %s", instanceID, bindingID))

	// find service plan
	_, err := cloudBroker.Catalog.FindServicePlan(details.ServiceID, details.PlanID)
	if err != nil {
		return brokerapi.UnbindSpec{}, err
	}

	// get detail service broker proxy from ServiceBrokerMap
	servicebrokerproxy := cloudBroker.ServiceBrokerMap[details.ServiceID]
	if servicebrokerproxy == nil {
		return brokerapi.UnbindSpec{}, fmt.Errorf("could not find service broker: %s", details.ServiceID)
	}

	// detail service broker proxy bind
	return servicebrokerproxy.Unbind(instanceID, bindingID, details, true)
}

// Update updates the given instance
func (cloudBroker *CloudServiceBroker) Update(
	ctx context.Context,
	instanceID string,
	details brokerapi.UpdateDetails,
	asyncAllowed bool) (brokerapi.UpdateServiceSpec, error) {

	cloudBroker.Logger.Debug(fmt.Sprintf("Update received. instanceID: %s", instanceID))

	if details.ServiceID == "" || details.PlanID == "" {
		e := cloudBroker.Catalog.ValidateAcceptsIncomplete(asyncAllowed)
		if e != nil {
			return brokerapi.UpdateServiceSpec{}, e
		}

		if instanceID != "" {
			// Check instance length in back database
			var length int
			err := database.BackDBConnection.
				Model(&database.InstanceDetails{}).
				Where("instance_id = ?", instanceID).
				Count(&length).Error
			if err != nil {
				return brokerapi.UpdateServiceSpec{}, err
			}
			// instance is existing
			if length > 0 {
				// Get InstanceDetails in back database
				iddetail := database.InstanceDetails{}
				err = database.BackDBConnection.
					Where("instance_id = ?", instanceID).
					First(&iddetail).Error
				if err != nil {
					return brokerapi.UpdateServiceSpec{}, err
				}

				details.ServiceID = iddetail.ServiceID
				details.PlanID = iddetail.PlanID

				cloudBroker.Logger.Debug(
					fmt.Sprintf("Auto discovery ServiceID: %s and PlanID: %s for instanceID: %s",
						details.ServiceID, details.PlanID, instanceID))
			}
		}
	}

	// find service plan
	_, err := cloudBroker.Catalog.FindServicePlan(details.ServiceID, details.PlanID)
	if err != nil {
		return brokerapi.UpdateServiceSpec{}, err
	}

	// get detail service broker proxy from ServiceBrokerMap
	servicebrokerproxy := cloudBroker.ServiceBrokerMap[details.ServiceID]
	if servicebrokerproxy == nil {
		return brokerapi.UpdateServiceSpec{}, fmt.Errorf("could not find service broker: %s", details.ServiceID)
	}

	// detail service broker proxy bind
	return servicebrokerproxy.Update(instanceID, details, asyncAllowed)
}

// LastOperation is called until the attempt times out or success or failure is returned
// if a service is provisioned or deprovision or update asynchronously
func (cloudBroker *CloudServiceBroker) LastOperation(
	ctx context.Context,
	instanceID string,
	details brokerapi.PollDetails) (brokerapi.LastOperation, error) {

	cloudBroker.Logger.Debug(fmt.Sprintf("LastOperation received. instanceID: %s", instanceID))

	operationData := details.OperationData
	if operationData != "" {
		// operationData is existing
		ods := database.OperationDetails{}
		err := json.Unmarshal([]byte(operationData), &ods)
		if err != nil {
			return brokerapi.LastOperation{}, err
		}
		// find service plan
		_, err = cloudBroker.Catalog.FindServicePlan(ods.ServiceID, ods.PlanID)
		if err != nil {
			return brokerapi.LastOperation{}, err
		}

		// get detail service broker proxy from ServiceBrokerMap
		servicebrokerproxy := cloudBroker.ServiceBrokerMap[ods.ServiceID]
		if servicebrokerproxy == nil {
			return brokerapi.LastOperation{}, fmt.Errorf("could not find service broker: %s", ods.ServiceID)
		}

		// detail service broker proxy bind
		return servicebrokerproxy.LastOperation(context.Background(), instanceID, domain.PollDetails{})
	} else {
		// Check OperationDetails length in back database
		var length int
		err := database.BackDBConnection.
			Model(&database.OperationDetails{}).
			Where("instance_id = ?", instanceID).
			Count(&length).Error
		if err != nil {
			return brokerapi.LastOperation{}, err
		}

		// last OperationDetails is exist
		if length > 0 {
			// get last OperationDetails in back database
			ods := database.OperationDetails{}
			err := database.BackDBConnection.
				Where("instance_id = ?", instanceID).
				Last(&ods).Error
			if err != nil {
				return brokerapi.LastOperation{}, err
			}
			// find service plan
			_, err = cloudBroker.Catalog.FindServicePlan(ods.ServiceID, ods.PlanID)
			if err != nil {
				return brokerapi.LastOperation{}, err
			}

			// get detail service broker proxy from ServiceBrokerMap
			servicebrokerproxy := cloudBroker.ServiceBrokerMap[ods.ServiceID]
			if servicebrokerproxy == nil {
				return brokerapi.LastOperation{}, fmt.Errorf("could not find service broker: %s", ods.ServiceID)
			}

			// detail service broker proxy bind
			lo, err := servicebrokerproxy.LastOperation(context.Background(), instanceID, domain.PollDetails{})
			if (lo.State == brokerapi.Succeeded) || (lo.State == brokerapi.Failed) {
				// Delete OperationDetails in back database
				err = database.BackDBConnection.Delete(&ods).Error
				if err != nil {
					return brokerapi.LastOperation{}, err
				}
			}
			return lo, err
		}
	}

	return brokerapi.LastOperation{State: brokerapi.Succeeded}, nil
}

//
func (cloudBroker *CloudServiceBroker) GetBinding(ctx context.Context, instanceID, bindingID string) (brokerapi.GetBindingSpec, error) {
	// get detail service broker proxy from ServiceBrokerMap
	servicebrokerproxy := cloudBroker.ServiceBrokerMap[""]
	if servicebrokerproxy == nil {
		return brokerapi.GetBindingSpec{}, fmt.Errorf("could not find service broker: %s", "")
	}
	return servicebrokerproxy.GetBinding(ctx, instanceID, bindingID)
}

func (cloudBroker *CloudServiceBroker) GetInstance(ctx context.Context, instanceID string) (brokerapi.GetInstanceDetailsSpec, error) {
	// get detail service broker proxy from ServiceBrokerMap
	servicebrokerproxy := cloudBroker.ServiceBrokerMap[""]
	if servicebrokerproxy == nil {
		return brokerapi.GetInstanceDetailsSpec{}, fmt.Errorf("could not find service broker: %s", "")
	}
	return servicebrokerproxy.GetInstance(ctx, instanceID)
}

func (cloudBroker *CloudServiceBroker) LastBindingOperation(ctx context.Context, instanceID, bindingID string, details brokerapi.PollDetails) (brokerapi.LastOperation, error) {
	// get detail service broker proxy from ServiceBrokerMap
	servicebrokerproxy := cloudBroker.ServiceBrokerMap[""]
	if servicebrokerproxy == nil {
		return brokerapi.LastOperation{}, fmt.Errorf("could not find service broker: %s", "")
	}
	return servicebrokerproxy.LastBindingOperation(ctx, instanceID, bindingID, details)
}
