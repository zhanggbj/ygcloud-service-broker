package mysql

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	brokerapi "github.com/pivotal-cf/brokerapi/domain"
	apiresponses "github.com/pivotal-cf/brokerapi/domain/apiresponses"
	"helm.sh/helm/v3/pkg/release"

	helmClient "github.com/mittwald/go-helm-client"
	"github.com/zhanggbj/ygcloud-service-broker/pkg/database"
	"github.com/zhanggbj/ygcloud-service-broker/pkg/models"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// Init cloud client
	err = m.CloudCredentials.Initial()
	if err != nil {
		return brokerapi.ProvisionedServiceSpec{}, fmt.Errorf("create cloud client failed. Error: %s", err)
	}

	// Find service
	// service, err := m.Catalog.FindService(details.ServiceID)
	// if err != nil {
	// 	return brokerapi.ProvisionedServiceSpec{}, fmt.Errorf("find dcs service failed. Error: %s", err)
	// }

	// // Find service plan
	// servicePlan, err := m.Catalog.FindServicePlan(details.ServiceID, details.PlanID)
	// if err != nil {
	// 	return brokerapi.ProvisionedServiceSpec{}, fmt.Errorf("find service plan failed. Error: %s", err)
	// }

	// TODO: servicePlan.Metadata API changed, need to update
	// Get parameters from service plan metadata
	// metadataParameters := MetadataParameters{}
	// if servicePlan.Metadata != nil {
	// 	if len(servicePlan.Metadata.AdditionalMetadata) > 0 {
	// 		err := json.Unmarshal(servicePlan.Metadata.Parameters, &metadataParameters)
	// 		if err != nil {
	// 			return brokerapi.ProvisionedServiceSpec{},
	// 				fmt.Errorf("Error unmarshalling Parameters from service plan: %s", err)
	// 		}
	// 	}
	// }

	// Get parameters from details
	provisionParameters := ProvisionParameters{}
	if len(details.RawParameters) > 0 {
		err := json.Unmarshal(details.RawParameters, &provisionParameters)
		if err != nil {
			return brokerapi.ProvisionedServiceSpec{},
				apiresponses.NewFailureResponse(fmt.Errorf("Error unmarshalling rawParameters from details: %s", err),
					http.StatusBadRequest, "Error unmarshalling rawParameters")
		}
		// Exist other unknown fields,
		if len(provisionParameters.UnknownFields) > 0 {
			return brokerapi.ProvisionedServiceSpec{},
				apiresponses.NewFailureResponse(
					fmt.Errorf("Parameters are not following schema: %+v", provisionParameters.UnknownFields),
					http.StatusBadRequest, "Parameters are not following schema")
		}
	}

	instanceNamespace := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: instanceID},
	}

	instanceNamespace, err = m.CloudCredentials.KubeClient.CoreV1().Namespaces().Create(context.Background(), instanceNamespace, metav1.CreateOptions{})
	if err != nil {
		m.Logger.Error("Failed to create namespace", err)
		return brokerapi.ProvisionedServiceSpec{}, fmt.Errorf("Provision failed. Error: %s", err)
	}

	values := "auth:\n" +
		"  rootPassword: " + provisionParameters.DatabasePassword + "\n" +
		"  database: " + provisionParameters.Name

	chartSpec := helmClient.ChartSpec{
		ReleaseName: fmt.Sprintf("mysql-%s", instanceID),
		ChartName:   "https://charts.bitnami.com/bitnami/mysql-8.8.26.tgz",
		Namespace:   instanceID,
		UpgradeCRDs: false,
		Wait:        false,
		ValuesYaml:  values,
	}

	// Install a chart release.
	// Note that helmclient.Options.Namespace should ideally match the namespace in chartSpec.Namespace.
	instance := &release.Release{}
	if instance, err = m.CloudCredentials.ClientSet.InstallOrUpgradeChart(context.Background(), &chartSpec); err != nil {
		m.Logger.Error(fmt.Sprintf("Failed to provision %s", instanceID), err)
	}

	// Log result
	m.Logger.Debug(fmt.Sprintf("provision mysql instance result: %v", models.ToJson(instance.Info.Status)))

	// Invoke sdk get
	freshInstance, err := m.CloudCredentials.ClientSet.GetRelease(chartSpec.ReleaseName)
	if err != nil {
		return brokerapi.ProvisionedServiceSpec{}, fmt.Errorf("get mysql instance failed. Error: %s", err)
	}

	// Constuct addtional info
	addtionalparam := map[string]string{}
	addtionalparam[AddtionalParamDBUsername] = "root"
	addtionalparam[AddtionalParamDBPassword] = provisionParameters.DatabasePassword
	addtionalparam[AddtionalParamDBname] = provisionParameters.Name

	// Marshal addtional info
	addtionalinfo, err := json.Marshal(addtionalparam)
	if err != nil {
		return brokerapi.ProvisionedServiceSpec{}, fmt.Errorf("marshal mysql addtional info failed. Error: %s", err)
	}

	// create InstanceDetails in back database
	idsOpts := database.InstanceDetails{
		ServiceID:      details.ServiceID,
		PlanID:         details.PlanID,
		InstanceID:     instanceID,
		TargetID:       instanceID,
		TargetName:     freshInstance.Name,
		TargetStatus:   string(freshInstance.Info.Status),
		TargetInfo:     "fake-target-infor",
		AdditionalInfo: string(addtionalinfo),
	}

	// log InstanceDetails opts
	m.Logger.Debug(fmt.Sprintf("create mysql instance in back database opts: %v", models.ToJson(idsOpts)))

	err = database.BackDBConnection.Create(&idsOpts).Error
	if err != nil {
		return brokerapi.ProvisionedServiceSpec{}, fmt.Errorf("create mysql instance in back database failed. Error: %s", err)
	}

	// Log InstanceDetails result
	m.Logger.Debug(fmt.Sprintf("create mysql instance in back database succeed: %s", instanceID))

	// Return result
	if asyncAllowed {
		// OperationDatas for OperationProvisioning
		ods := database.OperationDetails{
			OperationType:  models.OperationProvisioning,
			ServiceID:      details.ServiceID,
			PlanID:         details.PlanID,
			InstanceID:     instanceID,
			TargetID:       instanceID,
			TargetName:     freshInstance.Name,
			TargetStatus:   string(freshInstance.Info.Status),
			TargetInfo:     "k8s pod",
			AdditionalInfo: string(addtionalinfo),
		}

		operationdata, err := ods.ToString()
		if err != nil {
			return brokerapi.ProvisionedServiceSpec{}, fmt.Errorf("convert mysql instance operation datas failed. Error: %s", err)
		}

		// log OperationDatas
		m.Logger.Debug(fmt.Sprintf("create mysql instance operation datas: %s", operationdata))

		// Create OperationDetails
		err = database.BackDBConnection.Create(&ods).Error
		if err != nil {
			return brokerapi.ProvisionedServiceSpec{}, fmt.Errorf("create operation in back database failed. Error: %s", err)
		}
	}
	// Log Provision
	m.Logger.Debug(fmt.Sprintf("Provision finished %s", instanceID))

	dashboardUrl := fmt.Sprintf("http://example.dashboard.com/mysql/%s", instanceID)

	return brokerapi.ProvisionedServiceSpec{IsAsync: true, DashboardURL: dashboardUrl, OperationData: fmt.Sprintf("task_%s", instanceID)}, nil
}
