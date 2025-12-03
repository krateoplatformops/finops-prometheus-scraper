package config

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/rs/zerolog/log"

	"gopkg.in/yaml.v3"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	types "github.com/krateoplatformops/finops-prometheus-scraper/apis"
)

func ParseConfigFile(file string) (types.Config, error) {
	fileReader, err := os.OpenFile(file, os.O_RDONLY, 0600)
	if err != nil {
		return types.Config{}, err
	}
	defer fileReader.Close()
	data, err := io.ReadAll(fileReader)
	if err != nil {
		return types.Config{}, err
	}

	parse := types.ConfigFromFile{}

	err = yaml.Unmarshal(data, &parse)
	if err != nil {
		return types.Config{}, err
	}

	configuration := types.Config{}
	configuration.Exporter = parse.Exporter

	// Get databaseconfig CR from Kubernetes
	inClusterConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Error().Err(err).Msg("error occured while retrieving InClusterConfig in parsing config")
		return types.Config{}, err
	}
	clientset, err := kubernetes.NewForConfig(inClusterConfig)
	if err != nil {
		log.Error().Err(err).Msg("error occured while requesting clientset in parsing config")
		return types.Config{}, err
	}

	jsonData, err := clientset.RESTClient().
		Get().
		AbsPath("/apis/finops.krateo.io/v1").
		Namespace(parse.DatabaseConfigRef.Namespace).
		Resource("databaseconfigs").
		Name(parse.DatabaseConfigRef.Name).
		DoRaw(context.TODO())
	if err != nil {
		log.Error().Err(err).Msg("error occured while reuqesting database config")
		return types.Config{}, err
	}

	var crdResponse CRDResponse
	err = json.Unmarshal(jsonData, &crdResponse)
	if err != nil {
		log.Error().Err(err).Msg("error occured while unmarshaling databaseconfig")
		return types.Config{}, err
	}

	configuration.DatabaseConfig = crdResponse.Spec

	return configuration, nil
}

type CRDResponse struct {
	Spec types.DatabaseConfig `json:"spec"`
}
