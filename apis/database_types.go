package apis

import (
	finopsdatatypes "github.com/krateoplatformops/finops-data-types/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigFromFile struct {
	DatabaseConfigRef finopsdatatypes.ObjectRef `yaml:"databaseConfigRef"`
	Exporter          Exporter                  `yaml:"exporter"`
}

type Config struct {
	DatabaseConfig DatabaseConfig `yaml:"databaseConfig"`
	Exporter       Exporter       `yaml:"exporter"`
}

type DatabaseConfig struct {
	Username          string                            `yaml:"username" json:"username"`
	PasswordSecretRef finopsdatatypes.SecretKeySelector `yaml:"passwordSecretRef" json:"passwordSecretRef"`
}

type Exporter struct {
	API             finopsdatatypes.API      `yaml:"api"`
	PollingInterval metav1.Duration          `yaml:"pollingInterval"`
	TableName       string                   `yaml:"tableName"`
	MetricType      string                   `yaml:"metricType"`
	Generic         *finopsdatatypes.Generic `yaml:"generic"`
}

type MetricRecord struct {
	Labels    map[string]string `json:"labels"`
	Value     float64           `json:"value"`
	Timestamp int64             `json:"timestamp"`
}

type UsernamePassword struct {
	Username string
	Password string
}
