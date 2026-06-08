package config

import (
	"time"

	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/kelseyhightower/envconfig"
)

const (
	RedisTLSProtocol = "TLS"

	defaultBindAddr                   = "localhost:29900"
	defaultRedirectAPIURL             = "http://localhost:29900"
	defaultGracefulShutdownTimeout    = 5 * time.Second
	defaultHealthCheckInterval        = 30 * time.Second
	defaultHealthCheckCriticalTimeout = 90 * time.Second
	defaultOTBatchTimeout             = 5 * time.Second
	defaultOTExporterOTLPEndpoint     = "localhost:4317"
	defaultOTServiceName              = "dis-redirect-api"
	defaultOtelEnabled                = false
	defaultRedisAddress               = "localhost:6379"
)

// Config represents service configuration for dis-redirect-api
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	OTBatchTimeout             time.Duration `encconfig:"OTEL_BATCH_TIMEOUT"`
	OTExporterOTLPEndpoint     string        `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	OTServiceName              string        `envconfig:"OTEL_SERVICE_NAME"`
	OtelEnabled                bool          `envconfig:"OTEL_ENABLED"`
	RedirectAPIURL             string        `envconfig:"REDIRECT_API_URL"`
	RedisAddress               string        `envconfig:"REDIS_ADDRESS"`
	RedisClusterName           string        `envconfig:"REDIS_CLUSTER_NAME"`
	RedisRegion                string        `envconfig:"REDIS_REGION"`
	RedisSecProtocol           string        `envconfig:"REDIS_SEC_PROTO"`
	RedisService               string        `envconfig:"REDIS_SERVICE"`
	RedisUsername              string        `envconfig:"REDIS_USERNAME"`
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
		BindAddr:                   defaultBindAddr,
		RedirectAPIURL:             defaultRedirectAPIURL,
		GracefulShutdownTimeout:    defaultGracefulShutdownTimeout,
		HealthCheckInterval:        defaultHealthCheckInterval,
		HealthCheckCriticalTimeout: defaultHealthCheckCriticalTimeout,
		OTBatchTimeout:             defaultOTBatchTimeout,
		OTExporterOTLPEndpoint:     defaultOTExporterOTLPEndpoint,
		OTServiceName:              defaultOTServiceName,
		OtelEnabled:                defaultOtelEnabled,
		RedisAddress:               defaultRedisAddress,
		RedisClusterName:           "",
		RedisRegion:                "",
		RedisSecProtocol:           "",
		RedisService:               "",
		RedisUsername:              "",
		AuthorisationConfig:        authorisation.NewDefaultConfig(),
	}

	return cfg, envconfig.Process("", cfg)
}
