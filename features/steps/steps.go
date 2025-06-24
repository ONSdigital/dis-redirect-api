package steps

import (
	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
)

func (c *RedirectComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the redirect api is running$`, c.theRedirectAPIIsRunning)
}

func (c *RedirectComponent) theRedirectAPIIsRunning() {
	assert.Equal(c, true, c.ServiceRunning)
}
