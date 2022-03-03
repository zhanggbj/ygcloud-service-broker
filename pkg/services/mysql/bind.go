package mysql

import (
	"encoding/json"
	"fmt"

	brokerapi "github.com/pivotal-cf/brokerapi/domain"
	apiresponses "github.com/pivotal-cf/brokerapi/domain/apiresponses"
	"github.com/zhanggbj/ygcloud-service-broker/pkg/database"
	"github.com/zhanggbj/ygcloud-service-broker/pkg/models"
)

// Bind implematation
func (m *MySqlBroker) Bind(instanceID, bindingID string, details brokerapi.BindDetails, myBool bool) (brokerapi.Binding, error) {
	// Check mysql bind length in back database
	var length int
	err := database.BackDBConnection.
		Model(&database.BindDetails{}).
		Where("bind_id = ? and instance_id = ? and service_id = ? and plan_id = ?", bindingID, instanceID, details.ServiceID, details.PlanID).
		Count(&length).Error
	if err != nil {
		return brokerapi.Binding{}, fmt.Errorf("check mysql bind length in back database failed. Error: %s", err)
	}
	// ErrBindingAlreadyExistsSame
	if length > 0 {
		// Get BindDetails in back database
		bddetail := database.BindDetails{}
		err = database.BackDBConnection.
			Where("bind_id = ? and instance_id = ? and service_id = ? and plan_id = ?", bindingID, instanceID, details.ServiceID, details.PlanID).
			First(&bddetail).Error
		if err != nil {
			return brokerapi.Binding{}, fmt.Errorf("get binding in back database failed. Error: %s", err)
		}

		// Get additional info from InstanceDetails
		bindingdetail := brokerapi.Binding{}
		err = bddetail.GetBindInfo(&bindingdetail)
		if err != nil {
			return brokerapi.Binding{}, fmt.Errorf("get binding info failed. Error: %s", err)
		}
		return bindingdetail, apiresponses.ErrBindingAlreadyExists
	}

	// Check mysql instance length in back database
	err = database.BackDBConnection.
		Model(&database.InstanceDetails{}).
		Where("instance_id = ? and service_id = ? and plan_id = ?", instanceID, details.ServiceID, details.PlanID).
		Count(&length).Error
	if err != nil {
		return brokerapi.Binding{}, fmt.Errorf("check mysql instance length in back database failed. Error: %s", err)
	}
	// ErrInstanceDoesNotExist
	if length == 0 {
		return brokerapi.Binding{}, apiresponses.ErrInstanceDoesNotExist
	}

	// get InstanceDetails in back database
	ids := database.InstanceDetails{}
	err = database.BackDBConnection.
		Where("instance_id = ? and service_id = ? and plan_id = ?", instanceID, details.ServiceID, details.PlanID).
		First(&ids).Error
	if err != nil {
		return brokerapi.Binding{}, apiresponses.ErrInstanceDoesNotExist
	}

	// Log InstanceDetails
	m.Logger.Debug(fmt.Sprintf("mysql instance in back database: %v", models.ToJson(ids)))

	// Log opts
	m.Logger.Debug(fmt.Sprintf("bind mysql instance opts: instanceID: %s bindingID: %s", instanceID, bindingID))

	// Invoke sdk get
	releaseName := fmt.Sprintf("mysql-%s", instanceID)

	// Init cloud client
	err = m.CloudCredentials.Initial()
	if err != nil {
		return brokerapi.Binding{}, fmt.Errorf("create cloud client failed. Error: %s", err)
	}

	instance, err := m.CloudCredentials.ClientSet.GetRelease(releaseName)
	if err != nil {
		return brokerapi.Binding{}, fmt.Errorf("get mysql instance failed. Error: %s", err)
	}

	// Find service
	service, err := m.Catalog.FindService(details.ServiceID)
	if err != nil {
		return brokerapi.Binding{}, fmt.Errorf("find mysql service failed. Error: %s", err)
	}

	// Get additional info from InstanceDetails
	addtionalparam := map[string]string{}
	err = ids.GetAdditionalInfo(&addtionalparam)
	if err != nil {
		return brokerapi.Binding{}, fmt.Errorf("get mysql instance additional info failed. Error: %s", err)
	}

	// Get specified parameters
	dbname := addtionalparam[AddtionalParamDBname]
	dbpassword := addtionalparam[AddtionalParamDBPassword]

	// Build Binding Credential: Default database user name is root
	credential, err := BuildBindingCredential(instance.Name, 3306, dbname, "root", dbpassword, service.Name)
	if err != nil {
		return brokerapi.Binding{}, fmt.Errorf("build mysql instance binding credential failed. Error: %s", err)
	}

	// Log result
	m.Logger.Debug(fmt.Sprintf("bind mysql instance success: %v", models.ToJson(credential)))

	// Constuct result
	result := brokerapi.Binding{Credentials: credential}

	// Marshal bind info
	bindinfo, err := json.Marshal(result)
	if err != nil {
		return brokerapi.Binding{}, fmt.Errorf("marshal mysql bind info failed. Error: %s", err)
	}

	// create BindDetails in back database
	bdsOpts := database.BindDetails{
		ServiceID:      details.ServiceID,
		PlanID:         details.PlanID,
		InstanceID:     instanceID,
		BindID:         bindingID,
		BindInfo:       string(bindinfo),
		AdditionalInfo: "",
	}

	// log BindDetails opts
	m.Logger.Debug(fmt.Sprintf("create mysql bind in back database opts: %v", models.ToJson(bdsOpts)))

	err = database.BackDBConnection.Create(&bdsOpts).Error
	if err != nil {
		return brokerapi.Binding{}, fmt.Errorf("create mysql bind in back database failed. Error: %s", err)
	}

	// Log BindDetails result
	m.Logger.Debug(fmt.Sprintf("create mysql bind in back database succeed: %s", bindingID))

	// Return result
	return result, nil
}
