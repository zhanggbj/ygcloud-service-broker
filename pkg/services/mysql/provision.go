package mysql

import (
	"fmt"

	brokerapi "github.com/pivotal-cf/brokerapi/domain"
	apiresponses "github.com/pivotal-cf/brokerapi/domain/apiresponses"

	"github.com/zhanggbj/ygcloud-service-broker/pkg/database"
	"github.com/zhanggbj/ygcloud-service-broker/pkg/models"
)

// Provision implematation
func (m *MySqlBroker) Provision(instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool) (brokerapi.ProvisionedServiceSpec, error) {
	// Check accepts_incomplete if this service support async
	if models.OperationAsyncMYSQL {
		err := m.Catalog.ValidateAcceptsIncomplete(asyncAllowed)
		if err != nil {
			return brokerapi.ProvisionedServiceSpec{}, err
		}
	}
	svcSpec := brokerapi.ProvisionedServiceSpec{}
	return svcSpec, nil

	// Check if service instance alreay exists in backend database
	var length int
	err := database.BackDBConnection.
		Model(&database.InstanceDetails{}).
		Where("instance_id = ? and service_id = ? and plan_id = ?", instanceID, details.ServiceID, details.PlanID).
		Count(&length).Error
	if err != nil {
		return brokerapi.ProvisionedServiceSpec{}, fmt.Errorf("check mysql instance count in database failed. Error: %s", err)
	}
	// ErrInstanceAlreadyExists
	if length > 0 {
		// Get InstanceDetails in back database
		iddetail := database.InstanceDetails{}
		err = database.BackDBConnection.
			Where("instance_id = ? and service_id = ? and plan_id = ?", instanceID, details.ServiceID, details.PlanID).
			First(&iddetail).Error
		if err != nil {
			return brokerapi.ProvisionedServiceSpec{}, fmt.Errorf("get instance in database failed. Error: %s", err)
		}

		// Get additional info from InstanceDetails
		addtionalparamdetail := map[string]string{}
		err = iddetail.GetAdditionalInfo(&addtionalparamdetail)
		if err != nil {
			return brokerapi.ProvisionedServiceSpec{}, fmt.Errorf("get instance additional info failed. Error: %s", err)
		}

		// TODO: double confirm if this is still valid
		// Check AddtionalParamRequest exist
		if _, ok := addtionalparamdetail[AddtionalParamRequest]; ok {
			if (addtionalparamdetail[AddtionalParamRequest] != "") &&
				(addtionalparamdetail[AddtionalParamRequest] == string(details.RawParameters)) {
				return brokerapi.ProvisionedServiceSpec{}, apiresponses.ErrInstanceAlreadyExists
			}
		}
		return brokerapi.ProvisionedServiceSpec{}, apiresponses.ErrInstanceAlreadyExists
	}

	// Init dcs client
	mysqlClient, err := m.CloudCredentials.Initial()
	if err != nil {
		return brokerapi.ProvisionedServiceSpec{}, fmt.Errorf("create dcs client failed. Error: %s", err)
	}

	// Find service
	service, err := m.Catalog.FindService(details.ServiceID)
	if err != nil {
		return brokerapi.ProvisionedServiceSpec{}, fmt.Errorf("find dcs service failed. Error: %s", err)
	}

	// Find service plan
	servicePlan, err := m.Catalog.FindServicePlan(details.ServiceID, details.PlanID)
	if err != nil {
		return brokerapi.ProvisionedServiceSpec{}, fmt.Errorf("find service plan failed. Error: %s", err)
	}

	return brokerapi.ProvisionedServiceSpec{}, nil
}
