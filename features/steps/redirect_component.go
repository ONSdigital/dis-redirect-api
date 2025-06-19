package steps

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/ONSdigital/dis-redirect-api/api"
	"github.com/ONSdigital/dis-redirect-api/config"
	"github.com/ONSdigital/dis-redirect-api/service"
	"github.com/ONSdigital/dis-redirect-api/service/mock"
	disRedis "github.com/ONSdigital/dis-redis"
	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

const (
	gitCommitHash = "6584b786caac36b6214ffe04bf62f058d4021538"
	appVersion    = "v1.2.3"
)

type RedirectComponent struct {
	componenttest.ErrorFeature
	svcList        *service.ExternalServiceList
	svc            *service.Service
	errorChan      chan error
	Config         *config.Config
	HTTPServer     *http.Server
	ServiceRunning bool
	apiFeature     *componenttest.APIFeature
	redisFeature   *componenttest.RedisFeature
	StartTime      time.Time
}

func NewRedirectComponent(redisFeat *componenttest.RedisFeature) (*RedirectComponent, error) {
	c := &RedirectComponent{
		HTTPServer:     &http.Server{ReadHeaderTimeout: 3 * time.Second},
		errorChan:      make(chan error),
		ServiceRunning: false,
	}

	var err error

	c.Config, err = config.Get()
	if err != nil {
		return nil, err
	}

	c.redisFeature = redisFeat
	c.Config.RedisAddress = c.redisFeature.Server.Addr()

	initMock := &mock.InitialiserMock{
		DoGetHealthCheckFunc: c.DoGetHealthcheckOk,
		DoGetHTTPServerFunc:  c.DoGetHTTPServer,
		DoGetRedisClientFunc: c.DoGetRedisClientOk,
	}

	c.Config.HealthCheckInterval = 1 * time.Second
	c.Config.HealthCheckCriticalTimeout = 3 * time.Second
	c.svcList = service.NewServiceList(initMock)

	c.Config.BindAddr = "localhost:0"
	c.StartTime = time.Now()
	c.svc, err = service.Run(context.Background(), c.Config, c.svcList, "1", "", "", c.errorChan)
	if err != nil {
		return nil, err
	}
	c.ServiceRunning = true

	return c, nil
}

func (c *RedirectComponent) InitAPIFeature() *componenttest.APIFeature {
	c.apiFeature = componenttest.NewAPIFeature(c.InitialiseService)

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

func (c *RedirectComponent) DoGetHealthcheckOk(cfg *config.Config, buildTime, gitCommit, version string) (service.HealthChecker, error) {
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

func (c *RedirectComponent) DoGetRedisClientOk(ctx context.Context, cfg *config.Config) (api.RedisClient, error) {
	redisCli, err := disRedis.NewClient(ctx, &disRedis.ClientConfig{
		Address: cfg.RedisAddress,
	})

	return redisCli, err
}
