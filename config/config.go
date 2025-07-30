package config

import (
	"time"

	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/kelseyhightower/envconfig"
)

// Config represents service configuration for dis-redirect-api
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	RedirectAPIURL             string        `envconfig:"REDIRECT_API_URL"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	OTBatchTimeout             time.Duration `encconfig:"OTEL_BATCH_TIMEOUT"`
	OTExporterOTLPEndpoint     string        `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	OTServiceName              string        `envconfig:"OTEL_SERVICE_NAME"`
	OtelEnabled                bool          `envconfig:"OTEL_ENABLED"`
	RedisAddress               string        `envconfig:"REDIS_ADDRESS"`
	EnableURLRewriting         bool          `envconfig:"ENABLE_URL_REWRITING"`
	AuthorisationConfig        *authorisation.Config
}

var cfg *Config

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		BindAddr:                   "localhost:29900",
		RedirectAPIURL:             "http://localhost:29900",
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		OTBatchTimeout:             5 * time.Second,
		OTExporterOTLPEndpoint:     "localhost:4317",
		OTServiceName:              "dis-redirect-api",
		OtelEnabled:                false,
		RedisAddress:               "localhost:6379",
		EnableURLRewriting:         false,
		AuthorisationConfig:        authorisation.NewDefaultConfig(),
	}

	return cfg, envconfig.Process("", cfg)
}
