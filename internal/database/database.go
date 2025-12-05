package database

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/krateoplatformops/finops-prometheus-scraper/apis"
)

const maxBatchSize = 1000 // Number of records to send in each batch

func UploadMetrics(metrics []apis.MetricRecord, uploadServiceURL string, config apis.Config, usernamePassword *apis.UsernamePassword) error {
	// Split metrics into batches
	for i := 0; i < len(metrics); i += maxBatchSize {
		end := min(i+maxBatchSize, len(metrics))
		batch := metrics[i:end]

		// Convert batch to JSON
		jsonData, err := json.Marshal(batch)
		if err != nil {
			return fmt.Errorf("error marshaling data: %v", err)
		}

		// Create request
		url := ""
		if strings.ToLower(config.Exporter.MetricType) != "generic" {
			url = fmt.Sprintf("%s/upload?table=%s&type=%s", strings.Trim(uploadServiceURL, "/"), config.Exporter.TableName, config.Exporter.MetricType)
		} else if strings.ToLower(config.Exporter.MetricType) == "generic" {
			if config.Exporter.Generic != nil {
				url = fmt.Sprintf("%s/upload?table=%s&type=%s&value=%d", strings.Trim(uploadServiceURL, "/"), config.Exporter.TableName, config.Exporter.MetricType, config.Exporter.Generic.ValueColumnIndex)
			} else {
				return fmt.Errorf("generic object cannot be empty with metric type generic")
			}

		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("error creating request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth(usernamePassword.Username, usernamePassword.Password)

		// Send request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("error sending request: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		// Check response
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
		}

		log.Logger.Debug().Msg(string(body))

		log.Logger.Info().Msgf("Successfully uploaded batch of %d records, remaining %d", len(batch), len(metrics)-end)
	}
	return nil
}
