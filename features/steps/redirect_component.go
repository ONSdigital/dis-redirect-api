package steps

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/ONSdigital/dis-redirect-api/config"
	"github.com/ONSdigital/dis-redirect-api/service"
	"github.com/ONSdigital/dis-redirect-api/service/mock"
	"github.com/ONSdigital/dis-redirect-api/store"
	disRedis "github.com/ONSdigital/dis-redis"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-authorisation/v2/authorisationtest"
	componentTest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	permissionsSDK "github.com/ONSdigital/dp-permissions-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
)

const (
	gitCommitHash = "6584b786caac36b6214ffe04bf62f058d4021538"
	appVersion    = "v1.2.3"
)

type RedirectComponent struct {
	componentTest.ErrorFeature
	svcList                 *service.ExternalServiceList
	svc                     *service.Service
	errorChan               chan error
	Config                  *config.Config
	HTTPServer              *http.Server
	ServiceRunning          bool
	apiFeature              *componentTest.APIFeature
	redisFeature            *componentTest.RedisFeature
	StartTime               time.Time
	AuthorisationMiddleware authorisation.Middleware
}

func NewRedirectComponent(redisFeat *componentTest.RedisFeature) (*RedirectComponent, error) {
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

	fakePermissionsAPI := setupFakePermissionsAPI()
	c.Config.AuthorisationConfig.PermissionsAPIURL = fakePermissionsAPI.URL()

	initMock := &mock.InitialiserMock{
		DoGetHealthCheckFunc:             c.DoGetHealthcheckOk,
		DoGetHTTPServerFunc:              c.DoGetHTTPServer,
		DoGetRedisClientFunc:             c.DoGetRedisClientOk,
		DoGetAuthorisationMiddlewareFunc: c.DoGetAuthorisationMiddlewareOk,
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

func setupFakePermissionsAPI() *authorisationtest.FakePermissionsAPI {
	fakePermissionsAPI := authorisationtest.NewFakePermissionsAPI()
	bundle := getPermissionsBundle()
	fakePermissionsAPI.Reset()
	if err := fakePermissionsAPI.UpdatePermissionsBundleResponse(bundle); err != nil {
		log.Error(context.Background(), "failed to update permissions bundle response", err)
	}
	return fakePermissionsAPI
}

func getPermissionsBundle() *permissionsSDK.Bundle {
	return &permissionsSDK.Bundle{
		"legacy:edit": {
			"groups/role-admin": {
				{
					ID: "1",
				},
			},
		},
		"legacy:delete": {
			"groups/role-admin": {
				{
					ID: "1",
				},
			},
		},
	}
}
