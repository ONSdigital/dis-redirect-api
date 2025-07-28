package steps

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/ONSdigital/dis-redirect-api/config"
	"github.com/ONSdigital/dis-redirect-api/service"
	"github.com/ONSdigital/dis-redirect-api/store"
	disRedis "github.com/ONSdigital/dis-redis"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	componentTest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

const (
	gitCommitHash = "6584b786caac36b6214ffe04bf62f058d4021538"
	appVersion    = "v1.2.3"
)

type RedirectComponent struct {
	responseBody []byte
	componentTest.ErrorFeature
	svcList                 *service.ExternalServiceList
	svc                     *service.Service
	errorChan               chan error
	Config                  *config.Config
	HTTPServer              *http.Server
	ServiceRunning          bool
	apiFeature              *componentTest.APIFeature
	redisFeature            *componentTest.RedisFeature
	authFeature             *componentTest.AuthorizationFeature
	StartTime               time.Time
	AuthorisationMiddleware authorisation.Middleware
}

func NewRedirectComponent(redisFeat *componentTest.RedisFeature, authFeat *componentTest.AuthorizationFeature) (*RedirectComponent, error) {
	return &RedirectComponent{
		redisFeature:   redisFeat,
		authFeature:    authFeat, // keep this if needed
		errorChan:      make(chan error),
		ServiceRunning: false,
	}, nil
}

func (c *RedirectComponent) InitAPIFeature() *componentTest.APIFeature {
	c.apiFeature = componentTest.NewAPIFeature(c.InitialiseService)

	return c.apiFeature
}

func (c *RedirectComponent) Reset() *RedirectComponent {
	c.apiFeature.Reset()
	return c
}

func (c *RedirectComponent) Close() error {
	if c.svc != nil && c.ServiceRunning {
		c.redisFeature.Close()
		if err := c.svc.Close(context.Background()); err != nil {
			return err
		}
		c.ServiceRunning = false
	}

	return nil
}

func (c *RedirectComponent) InitialiseService() (http.Handler, error) {
	return c.HTTPServer.Handler, nil
}

func (c *RedirectComponent) DoGetHealthcheckOk(cfg *config.Config, _, _, _ string) (service.HealthChecker, error) {
	componentBuildTime := strconv.Itoa(int(time.Now().Unix()))
	versionInfo, err := healthcheck.NewVersionInfo(componentBuildTime, gitCommitHash, appVersion)
	if err != nil {
		return nil, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	return &hc, nil
}

func (c *RedirectComponent) DoGetHTTPServer(bindAddr string, router http.Handler) service.HTTPServer {
	c.HTTPServer = &http.Server{
		ReadHeaderTimeout: 3 * time.Second,
		Addr:              bindAddr,
		Handler:           router,
	}
	return c.HTTPServer
}

func (c *RedirectComponent) DoGetRedisClientOk(ctx context.Context, cfg *config.Config) (store.Redis, error) {
	redisCli, err := disRedis.NewClient(ctx, &disRedis.ClientConfig{
		Address: cfg.RedisAddress,
	})

	return redisCli, err
}

func (c *RedirectComponent) DoGetAuthorisationMiddlewareOk(ctx context.Context, cfg *authorisation.Config) (authorisation.Middleware, error) {
	middleware, err := authorisation.NewMiddlewareFromConfig(ctx, cfg, cfg.JWTVerificationPublicKeys)
	if err != nil {
		return nil, err
	}

	c.AuthorisationMiddleware = middleware
	return c.AuthorisationMiddleware, nil
}
