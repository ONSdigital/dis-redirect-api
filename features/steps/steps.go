package steps

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ONSdigital/dis-redirect-api/config"
	"github.com/ONSdigital/dis-redirect-api/service"
	"github.com/ONSdigital/dis-redirect-api/service/mock"
	"github.com/cucumber/godog"
)

func (c *RedirectComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	c.apiFeature.RegisterSteps(ctx)
	ctx.Step(`^the redirect api is running$`, c.theRedirectAPIIsRunning)
}

func (c *RedirectComponent) theRedirectAPIIsRunning() error {
	if c.ServiceRunning {
		return nil // already started
	}

	var err error

	// Register permissions bundle handler before starting the service
	if err := c.authFeature.RegisterDefaultPermissionsBundle(); err != nil {
		return fmt.Errorf("failed to register permissions bundle: %w", err)
	}

	c.Config, err = config.Get()
	if err != nil {
		return err
	}

	c.Config.RedisAddress = c.redisFeature.Server.Addr()
	c.Config.AuthorisationConfig.ZebedeeURL = c.authFeature.FakeAuthService.ResolveURL("")
	c.Config.AuthorisationConfig.PermissionsAPIURL = c.authFeature.FakePermissionsAPI.ResolveURL("")

	initMock := &mock.InitialiserMock{
		DoGetHealthCheckFunc:             c.DoGetHealthcheckOk,
		DoGetHTTPServerFunc:              c.DoGetHTTPServer,
		DoGetRedisClientFunc:             c.DoGetRedisClientOk,
		DoGetAuthorisationMiddlewareFunc: c.DoGetAuthorisationMiddlewareOk,
	}

	c.Config.HealthCheckInterval = 1 * time.Second
	c.Config.HealthCheckCriticalTimeout = 3 * time.Second
	c.Config.BindAddr = "localhost:0"
	c.StartTime = time.Now()
	c.svcList = service.NewServiceList(initMock)

	c.HTTPServer = &http.Server{ReadHeaderTimeout: 3 * time.Second}
	c.svc, err = service.Run(context.Background(), c.Config, c.svcList, "1", "", "", c.errorChan)
	if err != nil {
		return err
	}

	c.ServiceRunning = true

	return nil
}
