package steps

import (
	"github.com/ONSdigital/dp-authorisation/v2/authorisationtest"
	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
)

func (c *RedirectComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	c.apiFeature.RegisterSteps(ctx)
	ctx.Step(`^I am an admin user$`, c.adminJWTToken)
	ctx.Step(`^I am not authenticated$`, c.iAmNotAuthenticated)
	ctx.Step(`^the redirect api is running$`, c.theRedirectAPIIsRunning)
}

func (c *RedirectComponent) theRedirectAPIIsRunning() {
	assert.Equal(c, true, c.ServiceRunning)
}

func (c *RedirectComponent) adminJWTToken() error {
	err := c.apiFeature.ISetTheHeaderTo("Authorization", authorisationtest.AdminJWTToken)
	return err
}

func (c *RedirectComponent) iAmNotAuthenticated() error {
	err := c.apiFeature.ISetTheHeaderTo("Authorization", "")
	return err
}
