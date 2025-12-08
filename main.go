package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"time"

	"github.com/krateoplatformops/finops-prometheus-scraper/apis"
	"github.com/krateoplatformops/finops-prometheus-scraper/internal/config"
	"github.com/krateoplatformops/finops-prometheus-scraper/internal/database"
	localendpoints "github.com/krateoplatformops/finops-prometheus-scraper/internal/helpers/kube/endpoints"
	localrequest "github.com/krateoplatformops/finops-prometheus-scraper/internal/helpers/kube/http/request"
	localstatus "github.com/krateoplatformops/finops-prometheus-scraper/internal/helpers/kube/http/response"
	"github.com/krateoplatformops/finops-prometheus-scraper/internal/helpers/kube/secrets"
	"github.com/krateoplatformops/plumbing/http/request"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"k8s.io/client-go/rest"

	"github.com/rs/zerolog/log"

	finopsdatatypes "github.com/krateoplatformops/finops-data-types/api/v1"
)

func parseMF(data []byte) (map[string]*dto.MetricFamily, error) {
	var parser expfmt.TextParser
	mf, err := parser.TextToMetricFamilies(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return mf, nil
}

func WriteProm(api finopsdatatypes.API) ([]byte, error) {
	time.Sleep(2 * time.Second)

	rc, _ := rest.InClusterConfig()
	endpoint, err := localendpoints.FromSecret(context.Background(), rc, api.EndpointRef)
	if err != nil {
		return nil, err
	}

	res := &localstatus.Status{Code: 500}
	var bodyData []byte

	for ok := true; ok; ok = (res.Code != 200) {
		opts := request.RequestOptions{
			Endpoint: &endpoint,
			RequestInfo: request.RequestInfo{
				Path:    api.Path,
				Verb:    &api.Verb,
				Headers: api.Headers,
				Payload: &api.Payload,
			},
			ResponseHandler: func(rc io.ReadCloser) error {
				bodyData, _ = io.ReadAll(rc)
				return nil
			},
		}
		// log.Info().Msgf("Parsed Endpoint awsAccessKey: %s", opts.Endpoint.AwsAccessKey)
		// log.Info().Msgf("Parsed Endpoint awsSecretKey: %s", opts.Endpoint.AwsSecretKey)
		// log.Info().Msgf("Parsed Endpoint awsRegion: %s", opts.Endpoint.AwsRegion)
		// log.Info().Msgf("Parsed Endpoint awsService: %s", opts.Endpoint.AwsService)

		// log.Info().Msgf("Endpoint HasAwsAuth: %t", opts.Endpoint.HasAwsAuth())

		res = localrequest.Do(context.Background(), opts)

		if res.Code != 200 {
			log.Warn().Msgf("Received status code %d", res.Code)
			log.Warn().Msgf("Body %s", string(bodyData))

			log.Logger.Warn().Msgf("Retrying connection in 5s...")
			time.Sleep(5 * time.Second)

			log.Logger.Info().Msgf("Parsing Endpoint again...")
			rc, _ := rest.InClusterConfig()
			endpoint, err = localendpoints.FromSecret(context.Background(), rc, api.EndpointRef)
			if err != nil {
				continue
			}
		}
	}

	return bodyData, nil
}

func main() {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Error().Err(err).Msg("error occured while retrieving InClusterConfig, halting...")
		return
	}

	cli, err := secrets.NewSecretsRESTClient(cfg)
	if err != nil {
		log.Error().Err(err).Msg("error occured while creating REST client, halting...")
		return
	}

	uploadServiceURL := os.Getenv("URL_DB_WEBSERVICE")
	time.Sleep(5 * time.Second)
	for {
		config, err := config.ParseConfigFile("/config/config.yaml")
		if err != nil {
			log.Error().Err(err).Msg("error occured while parsing scraper configuration, halting...")
			return
		}

		log.Logger.Info().Msg("Starting loop...")

		passwordSecret, err := secrets.GetSecret(context.Background(), secrets.ClientOptions{
			Cli:       cli,
			Name:      config.DatabaseConfig.PasswordSecretRef.Name,
			Namespace: config.DatabaseConfig.PasswordSecretRef.Namespace,
		})
		if err != nil {
			log.Error().Err(err).Msg("error occured while retrieving password secret, continuing to next cycle...")
			continue
		}
		usernamePassword := &apis.UsernamePassword{
			Username: string(config.DatabaseConfig.Username),
			Password: string(passwordSecret.Data[config.DatabaseConfig.PasswordSecretRef.Key]),
		}

		// Get and verify metrics data
		data, err := WriteProm(config.Exporter.API)
		if err != nil {
			log.Error().Err(err).Msg("Error while writing prometheus file")
		}

		second_file := []byte{}
		for len(data) != len(second_file) || len(data) == 0 {
			second_file = data
			data, err = WriteProm(config.Exporter.API)
			if err != nil {
				log.Error().Err(err).Msg("error while writing prometheus file (loop)")
			}
			seconds := 5 * time.Second
			log.Logger.Info().Msgf("Exporter is still updating or has not published anything yet, waiting %s...", seconds)
			time.Sleep(seconds)
		}

		// Parse metrics
		mf, err := parseMF(data)
		if err != nil {
			log.Error().Err(err).Msg("Error while reading prometheus metrics from file")
		}

		// Convert metrics to records
		var metrics []apis.MetricRecord
		timestamp := time.Now().Unix()

		for _, value := range mf {
			if config.Exporter.Generic != nil {
				if config.Exporter.Generic.MetricName != "" {
					if *value.Name != config.Exporter.Generic.MetricName {
						continue
					}
				}
			}
			for _, metric := range value.Metric {
				record := apis.MetricRecord{
					Labels:    make(map[string]string),
					Value:     metric.GetGauge().GetValue(),
					Timestamp: timestamp,
				}

				// Convert labels to map
				for _, label := range metric.Label {
					record.Labels[*label.Name] = *label.Value
				}

				metrics = append(metrics, record)
			}
		}

		// Upload metrics in batches
		uploadTime := time.Now()
		err = database.UploadMetrics(metrics, uploadServiceURL, config, usernamePassword)
		if err != nil {
			log.Logger.Warn().Msgf("Error uploading metrics: %v, continuing...", err)
		}
		log.Debug().Msgf("Upload took: %d ms", time.Since(uploadTime).Milliseconds())

		// Wait for next polling interval
		log.Debug().Msgf("Polling interval set to %s, starting sleep...", config.Exporter.PollingInterval.Duration.String())
		time.Sleep(config.Exporter.PollingInterval.Duration)
	}
}
