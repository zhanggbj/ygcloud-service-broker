package mysql

import (
	"code.cloudfoundry.org/lager"
	"github.com/zhanggbj/ygcloud-service-broker/pkg/config"
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

// MetadataParameters represent plan metadata parameters in config
type MetadataParameters struct {
	Engine            string   `json:"engine,omitempty"`
	EngineVersion     string   `json:"engine_version,omitempty"`
	SpecCode          string   `json:"speccode,omitempty"`
	ChargingType      string   `json:"charging_type,omitempty"`
	Capacity          int      `json:"capacity,omitempty"`
	VPCID             string   `json:"vpc_id,omitempty"`
	SubnetID          string   `json:"subnet_id,omitempty"`
	SecurityGroupID   string   `json:"security_group_id,omitempty"`
	AvailabilityZones []string `json:"availability_zones,omitempty"`
}
