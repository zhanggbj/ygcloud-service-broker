package redis

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/zhanggbj/ygcloud-service-broker/pkg/database"
	"github.com/zhanggbj/ygcloud-service-broker/pkg/models"
	"helm.sh/helm/v3/pkg/release"
)

func SyncStatusWithService(r *RedisBroker, instanceID, serviceID, planID string) (*release.Release, error, error) {
	dbInstance := database.InstanceDetails{}

	r.Logger.Debug(fmt.Sprintf("SyncStatusWithService start"))
	// Log opts
	r.Logger.Debug(fmt.Sprintf("SyncStatusWithService instance opts: instanceID: %s serviceID: %s "+
		"planID: %s", instanceID, serviceID, planID))

	// Init cloud client
	err := r.CloudCredentials.Initial()
	if err != nil {
		return &release.Release{}, err, errors.New(fmt.Sprintf("SyncStatusWithService failed to init cloud client %s", instanceID))
	}

	// Get status
	instance := &release.Release{}
	ReleaseName := fmt.Sprintf("postgres-%s", instanceID)
	if instance, err = r.CloudCredentials.ClientSet.GetRelease(ReleaseName); err != nil {
		r.Logger.Error(fmt.Sprintf("Failed to get instance %s", instanceID), err)
	}

	// TODO: Add code to handle different operational status, like updating in progress
	// get InstanceDetails in back database
	err = database.BackDBConnection.
		Where("instance_id = ? and service_id = ? and plan_id = ?", instanceID, serviceID, planID).
		First(&dbInstance).Error
	if err != nil {
		r.Logger.Debug(fmt.Sprintf("SyncStatusWithService get mysql instance in back database failed. Error: %s", err))
		return instance, fmt.Errorf("SyncStatusWithService get mysql instance failed. Error: %s", err), nil
	}
	// Log InstanceDetails
	r.Logger.Debug(fmt.Sprintf("SyncStatusWithService mysql instance in back database: %v", models.ToJson(dbInstance)))
	// update target info in back database
	targetInfo, err := json.Marshal(instance)
	if err != nil {
		return instance, fmt.Errorf("SyncStatusWithService marshal mysql instance failed. Error: %s", err), nil
	}

	dbInstance.TargetID = instanceID
	dbInstance.TargetName = ReleaseName
	dbInstance.TargetStatus = string(instance.Info.Status)
	dbInstance.TargetInfo = string(targetInfo)

	err = database.BackDBConnection.Save(&dbInstance).Error
	if err != nil {
		r.Logger.Debug(fmt.Sprintf("SyncStatusWithService update  instance target status in back database failed. "+
			"Error: %s", err))
		return instance, fmt.Errorf("SyncStatusWithService update mysql instance target status failed. Error: %s", err), nil
	}
	// Sync target status success
	r.Logger.Debug(fmt.Sprintf("SyncStatusWithService update mysql instance target status succeed: %s", instanceID))

	return instance, nil, nil
}
