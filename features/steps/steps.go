package steps

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/ONSdigital/dis-redirect-api/api"
	"github.com/ONSdigital/dis-redirect-api/models"
	"github.com/cucumber/godog"
	"github.com/rdumont/assistdog"
	"github.com/stretchr/testify/assert"
)

func (c *RedirectComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the redirect api is running$`, c.theRedirectAPIIsRunning)
	ctx.Step(`^I would expect there to be three or more redirects returned in a list$`, c.iWouldExpectThereToBeThreeOrMoreRedirectsReturnedInAList)
	ctx.Step(`^in each redirect I would expect the response to contain values that have these structures$`, c.inEachRedirectIWouldExpectTheResponseToContainValuesThatHaveTheseStructures)
	ctx.Step(`^the list of redirects should also contain the following values:$`, c.theListOfRedirectsShouldAlsoContainTheFollowingValues)
	ctx.Step(`^I would expect there to be (\d+) redirects returned in a list$`, c.iWouldExpectThereToBeRedirectsReturnedInAList)
}

func (c *RedirectComponent) theRedirectAPIIsRunning() {
	assert.Equal(c, true, c.ServiceRunning)
}

func (c *RedirectComponent) iWouldExpectThereToBeThreeOrMoreRedirectsReturnedInAList() error {
	c.responseBody, _ = io.ReadAll(c.apiFeature.HTTPResponse.Body)

	var response models.Redirects
	err := json.Unmarshal(c.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}
	numRedirectsFound := len(response.RedirectList)
	assert.True(&c.ErrorFeature, numRedirectsFound >= 3, "The list should contain three or more redirects but it only contains "+strconv.Itoa(numRedirectsFound))

	return c.ErrorFeature.StepError()
}

func (c *RedirectComponent) inEachRedirectIWouldExpectTheResponseToContainValuesThatHaveTheseStructures(table *godog.Table) error {
	var response models.Redirects

	err := json.Unmarshal(c.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	for i := range response.RedirectList {
		redirect := response.RedirectList[i]
		err := c.checkStructure(&redirect)
		if err != nil {
			return fmt.Errorf("failed to check that the response has the expected structure: %w", err)
		}
	}
	return c.ErrorFeature.StepError()
}

func (c *RedirectComponent) checkStructure(responseRedirect *models.Redirect) error {
	from := responseRedirect.From
	assert.NotEmpty(&c.ErrorFeature, from)
	assert.NotEmpty(&c.ErrorFeature, responseRedirect.To)
	encodedFrom := api.EncodeBase64(from)
	assert.Equal(&c.ErrorFeature, encodedFrom, responseRedirect.Id)
	expectedSelfHref := "https://api.beta.ons.gov.uk/v1/redirects/" + encodedFrom
	assert.Equal(&c.ErrorFeature, expectedSelfHref, responseRedirect.Links.Self.Href)
	assert.Equal(&c.ErrorFeature, encodedFrom, responseRedirect.Links.Self.Id)
	return nil
}

func (c *RedirectComponent) theListOfRedirectsShouldAlsoContainTheFollowingValues(table *godog.Table) error {
	expectedResult, err := assistdog.NewDefault().ParseMap(table)
	if err != nil {
		return fmt.Errorf("failed to parse table: %w", err)
	}
	var response models.Redirects

	err = json.Unmarshal(c.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	c.checkValuesInRedirects(expectedResult, response)

	return c.ErrorFeature.StepError()
}

func (c *RedirectComponent) checkValuesInRedirects(expectedResult map[string]string, redirectsList models.Redirects) {
	assert.Equal(&c.ErrorFeature, expectedResult["count"], strconv.Itoa(redirectsList.Count))
	assert.Equal(&c.ErrorFeature, expectedResult["cursor"], redirectsList.Cursor)
	assert.Equal(&c.ErrorFeature, expectedResult["next_cursor"], redirectsList.NextCursor)
	assert.Equal(&c.ErrorFeature, expectedResult["total_count"], strconv.Itoa(redirectsList.TotalCount))
}

func (c *RedirectComponent) iWouldExpectThereToBeRedirectsReturnedInAList(expectedNumRedirects int) error {
	c.responseBody, _ = io.ReadAll(c.apiFeature.HTTPResponse.Body)

	var response models.Redirects
	err := json.Unmarshal(c.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}
	numRedirectsFound := len(response.RedirectList)
	assert.True(&c.ErrorFeature, numRedirectsFound == expectedNumRedirects, "The list should contain "+strconv.Itoa(expectedNumRedirects)+" redirects but it contains "+strconv.Itoa(numRedirectsFound))

	return c.ErrorFeature.StepError()
}
