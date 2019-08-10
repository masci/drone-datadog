package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	// BuildCommit contains the git sha of the commit built, defaults to empty
	BuildCommit = "-"
	// BuildTag contains the tag of the commit built, defaults to 'snapshot'
	BuildTag = "snapshot"
)

// Config contains what's in the `settings` section of the config file
type Config struct {
	APIKey string
	DryRun bool
}

// Metric represents a point that'll be sent to Datadog
type Metric struct {
	Name  string
	Type  string
	Value float32
	Host  string
	Tags  []string
}

// Metrics is a type alias for a slice of Metric
type Metrics []Metric

// Event represents an event that'll be sent to Datadog
type Event struct {
	Title     string   `json:"title"`
	Text      string   `json:"text"`
	AlertType string   `json:"alert_type,omitempty"`
	Host      string   `json:"host,omitempty"`
	Tags      []string `json:"tags,omitempty"`
}

// Events is a type alias for a slice ov Event
type Events []Event

// MarshalJSON provides custom serialization for the Metric object
func (m Metric) MarshalJSON() ([]byte, error) {
	points := []float32{
		float32(time.Now().Unix()),
		m.Value,
	}

	return json.Marshal(&struct {
		Name   string    `json:"name"`
		Type   string    `json:"type,omitempty"`
		Host   string    `json:"host,omitempty"`
		Tags   []string  `json:"tags,omitempty"`
		Points []float32 `json:"points"`
	}{
		Name:   m.Name,
		Type:   m.Type,
		Host:   m.Host,
		Tags:   m.Tags,
		Points: points,
	})
}

func isValidMetricType(t string) bool {
	return map[string]bool{
		"gauge": true,
		"rate":  true,
		"count": true,
		"":      true, // will default to `gauge`
	}[t]
}

func isValidAlertType(t string) bool {
	return map[string]bool{
		"info":    true,
		"success": true,
		"warning": true,
		"error":   true,
		"":        true, // will default to `info`
	}[t]
}

func printVersion() {
	log.Printf("Drone-Datadog Plugin version: %s - git commit: %s", BuildTag, BuildCommit)
}

func parseConfig() (*Config, error) {
	cfg := &Config{
		APIKey: os.Getenv("PLUGIN_API_KEY"),
		DryRun: os.Getenv("PLUGIN_DRY_RUN") == "true",
	}

	if cfg.APIKey == "" && !cfg.DryRun {
		return nil, fmt.Errorf("Datadog API Key is missing")
	}

	return cfg, nil
}

func parseMetrics() (Metrics, error) {
	configData := []byte(os.Getenv("PLUGIN_METRICS"))
	data := Metrics{}
	if err := json.Unmarshal(configData, &data); err != nil {
		return nil, fmt.Errorf("metrics configuration error: %v", err)
	}

	metrics := Metrics{}
	for _, m := range data {
		if !isValidMetricType(m.Type) {
			log.Printf("invalid metric type: %s", m.Type)
			continue
		}

		metrics = append(metrics, m)
	}

	return metrics, nil
}

func parseEvents() (Events, error) {
	configData := []byte(os.Getenv("PLUGIN_EVENTS"))
	data := Events{}
	if err := json.Unmarshal(configData, &data); err != nil {
		return nil, fmt.Errorf("events configuration error: %v", err)
	}

	events := Events{}
	for _, ev := range data {
		if !isValidAlertType(ev.AlertType) {
			log.Printf("invalid alert type: %s", ev.AlertType)
			continue
		}

		events = append(events, ev)
	}

	return events, nil
}

func send(url string, payload []byte, dryRun bool) error {
	if dryRun {
		log.Println("Dry run, logging payload:")
		log.Println(string(payload))
		return nil
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	if res, err := http.DefaultClient.Do(req); err == nil {
		if res.StatusCode >= 300 {
			return fmt.Errorf("server responded with: %s", res.Status)
		}
	} else {
		return fmt.Errorf("unable to send data: %s", err)
	}

	return nil
}

func main() {
	showVersion := flag.Bool("v", false, "print plugin version")
	flag.Parse()

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	cfg, err := parseConfig()
	if err != nil {
		log.Fatalf("parseConfig() err: %s", err)
	}

	if metrics, err := parseMetrics(); err == nil {
		payload, err := json.Marshal(struct {
			Series Metrics `json:"series"`
		}{
			Series: metrics,
		})

		if err != nil {
			log.Printf("error encoding metrics: %v", err)
		} else {
			url := fmt.Sprintf("https://api.datadoghq.com/api/v1/series?api_key=%s", cfg.APIKey)
			if err := send(url, payload, cfg.DryRun); err != nil {
				log.Fatalf("unable to send metrics: %v", err)
			}

			log.Printf("%d metric(s) sent successfully", len(metrics))
		}
	} else {
		log.Println(err)
	}

	if events, err := parseEvents(); err == nil {
		url := fmt.Sprintf("https://api.datadoghq.com/api/v1/events?api_key=%s", cfg.APIKey)
		successCount := 0
		// events must be posted one at a time
		for _, ev := range events {
			payload, err := json.Marshal(ev)
			if err != nil {
				log.Printf("error encoding event: %v", err)
				continue
			}

			if err := send(url, payload, cfg.DryRun); err != nil {
				log.Printf("unable to send event: %v", err)
				continue
			}

			successCount++
		}

		log.Printf("%d event(s) sent successfully", successCount)
	}
}
