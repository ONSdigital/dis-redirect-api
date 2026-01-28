package service

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/ONSdigital/dis-redirect-api/config"
	"github.com/ONSdigital/dis-redirect-api/store"
	disRedis "github.com/ONSdigital/dis-redis"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v3/http"
	"github.com/ONSdigital/log.go/v2/log"
)

// ExternalServiceList holds the initialiser and initialisation state of external services.
type ExternalServiceList struct {
	AuthorisationMiddleware bool
	HealthCheck             bool
	Init                    Initialiser
	Redis                   bool
}

// NewServiceList creates a new service list with the provided initialiser
func NewServiceList(initialiser Initialiser) *ExternalServiceList {
	return &ExternalServiceList{
		Init: initialiser,
	}
}

// Init implements the Initialiser interface to initialise dependencies
type Init struct{}

// GetHTTPServer creates an http server
func (e *ExternalServiceList) GetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	s := e.Init.DoGetHTTPServer(bindAddr, router)
	return s
}

// GetHealthCheck creates a healthcheck with versionInfo and sets teh HealthCheck flag to true
func (e *ExternalServiceList) GetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	hc, err := e.Init.DoGetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	e.HealthCheck = true
	return hc, nil
}

// DoGetHTTPServer creates an HTTP Server with the provided bind address and router
func (e *Init) DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	s := dphttp.NewServer(bindAddr, router)
	s.HandleOSSignals = false
	return s
}

// DoGetHealthCheck creates a healthcheck with versionInfo
func (e *Init) DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	return &hc, nil
}

// GetRedisClient creates a Redis client and sets the Redis flag to true
func (e *ExternalServiceList) GetRedisClient(ctx context.Context, cfg *config.Config) (store.Redis, error) {
	redis, err := e.Init.DoGetRedisClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	e.Redis = true
	return redis, nil
}

// DoGetRedisClient initialises a dis-redis client
func (e *Init) DoGetRedisClient(ctx context.Context, cfg *config.Config) (store.Redis, error) {
	clientCfg := &disRedis.ClientConfig{
		Address:     cfg.RedisAddress,
		ClusterName: cfg.RedisClusterName,
		Region:      cfg.RedisRegion,
		Service:     cfg.RedisService,
		Username:    cfg.RedisUsername,
	}

	if cfg.RedisSecProtocol == config.RedisTLSProtocol {
		log.Info(ctx, "redis TLS protocol specified, initializing dis-redis client with TLS")
		clientCfg.TLSConfig = &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: false,
		}
	}

	var redisClient store.Redis
	var err error

	if cfg.RedisRegion != "" && cfg.RedisService != "" && cfg.RedisClusterName != "" {
		redisClient, err = disRedis.NewClusterClient(ctx, clientCfg)
		if err != nil {
			log.Error(ctx, "failed to create dis-redis cluster client", err)
			return nil, err
		}
	} else {
		redisClient, err = disRedis.NewClient(ctx, clientCfg)
		if err != nil {
			log.Error(ctx, "failed to create dis-redis client", err)
			return nil, err
		}
	}

	return redisClient, nil
}

// DoGetAuthorisationMiddleware creates authorisation middleware for the given config
func (e *Init) DoGetAuthorisationMiddleware(ctx context.Context, authorisationConfig *authorisation.Config) (authorisation.Middleware, error) {
	return authorisation.NewFeatureFlaggedMiddleware(ctx, authorisationConfig, authorisationConfig.JWTVerificationPublicKeys)
}

// GetAuthorisationMiddleware creates a new instance of authorisation.Middlware
func (e *ExternalServiceList) GetAuthorisationMiddleware(ctx context.Context, authorisationConfig *authorisation.Config) (authorisation.Middleware, error) {
	e.AuthorisationMiddleware = true
	return e.Init.DoGetAuthorisationMiddleware(ctx, authorisationConfig)
}
