package mysql

import (
	"encoding/json"

	"code.cloudfoundry.org/lager"
	"github.com/zhanggbj/ygcloud-service-broker/pkg/config"
	"gopkg.in/mgo.v2/bson"
)

// DCSBroker define
type MySqlBroker struct {
	CloudCredentials config.CloudCredentials
	Catalog          config.Catalog
	Logger           lager.Logger
}

// BindingCredential represent dcs binding credential
type BindingCredential struct {
	IP       string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	UserName string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Name     string `json:"name,omitempty"`
	Type     string `json:"type,omitempty"`
}

// ProvisionParameters represent provision parameters
type ProvisionParameters struct {
	SpecCode                string                 `json:"speccode,omitempty" bson:"speccode,omitempty"`
	VolumeType              string                 `json:"volume_type,omitempty" bson:"volume_type,omitempty"`
	VolumeSize              int                    `json:"volume_size,omitempty" bson:"volume_size,omitempty"`
	AvailabilityZone        string                 `json:"availability_zone,omitempty" bson:"availability_zone,omitempty"`
	VPCID                   string                 `json:"vpc_id,omitempty" bson:"vpc_id,omitempty"`
	SubnetID                string                 `json:"subnet_id,omitempty" bson:"subnet_id,omitempty"`
	SecurityGroupID         string                 `json:"security_group_id,omitempty" bson:"security_group_id,omitempty"`
	Name                    string                 `json:"name,omitempty" bson:"name,omitempty"`
	DatabasePort            string                 `json:"database_port,omitempty" bson:"database_port,omitempty"`
	DatabasePassword        string                 `json:"database_password,omitempty" bson:"database_password,omitempty"`
	BackupStrategyStarttime string                 `json:"backup_strategy_starttime,omitempty" bson:"backup_strategy_starttime,omitempty"`
	BackupStrategyKeepdays  int                    `json:"backup_strategy_keepdays,omitempty" bson:"backup_strategy_keepdays,omitempty"`
	HAEnable                bool                   `json:"ha_enable,omitempty" bson:"ha_enable,omitempty"`
	HAReplicationMode       string                 `json:"ha_replicationmode,omitempty" bson:"ha_replicationmode,omitempty"`
	UnknownFields           map[string]interface{} `json:"-" bson:",inline"`
}

func (f *ProvisionParameters) MarshalJSON() ([]byte, error) {
	var j interface{}
	b, _ := bson.Marshal(f)
	bson.Unmarshal(b, &j)
	return json.Marshal(&j)
}

// Collect unknown fields into "UnknownFields"
func (f *ProvisionParameters) UnmarshalJSON(b []byte) error {
	var j map[string]interface{}
	json.Unmarshal(b, &j)
	b, _ = bson.Marshal(&j)
	return bson.Unmarshal(b, f)
}

const (
	// AddtionalParamUsername for username
	AddtionalParamUsername string = "username"
	// AddtionalParamPassword for password
	AddtionalParamPassword string = "password"
	// AddtionalParamRequest for request
	AddtionalParamRequest string = "request"
)
