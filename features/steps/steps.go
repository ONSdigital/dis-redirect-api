package steps

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/ONSdigital/dis-redirect-api/models"
	"github.com/cucumber/godog"
	"github.com/cucumber/messages/go/v21"
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

	return nil
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
	return nil
}

func (c *RedirectComponent) checkStructure(responseRedirect *models.Redirect) error {
	from := responseRedirect.From
	assert.NotEmpty(&c.ErrorFeature, from)
	assert.NotEmpty(&c.ErrorFeature, responseRedirect.To)
	encodedFrom := base64.StdEncoding.EncodeToString([]byte(from))
	assert.Equal(&c.ErrorFeature, encodedFrom, responseRedirect.Id)
	expectedSelfHref := "https://api.beta.ons.gov.uk/v1/redirects/" + encodedFrom
	assert.Equal(&c.ErrorFeature, expectedSelfHref, responseRedirect.Links.Self.Href)
	assert.Equal(&c.ErrorFeature, encodedFrom, responseRedirect.Links.Self.Id)
	return nil
}

func (c *RedirectComponent) theListOfRedirectsShouldAlsoContainTheFollowingValues(table *godog.Table) error {
	var response models.Redirects

	err := json.Unmarshal(c.responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	for i, row := range table.Rows {
		if i == 0 { // Skip header row
			continue
		}
		c.checkValuesInRedirects(row, response)
	}

	return nil
}

func (c *RedirectComponent) checkValuesInRedirects(row *messages.PickleTableRow, redirectsList models.Redirects) {
	strExpectedCount := row.Cells[0].Value
	intExpectedCount, _ := strconv.Atoi(strExpectedCount)
	intObservedCount := redirectsList.Count
	assert.True(&c.ErrorFeature, intExpectedCount == intObservedCount, "expected count to equal "+strExpectedCount+"but it is "+strconv.Itoa(intObservedCount))
	strExpectedCursor := row.Cells[1].Value
	intExpectedCursor, _ := strconv.Atoi(strExpectedCursor)
	intObservedCursor, _ := strconv.Atoi(redirectsList.Cursor)
	assert.True(&c.ErrorFeature, intExpectedCursor == intObservedCursor, "expected cursor to equal "+strExpectedCursor+" but it is "+redirectsList.Cursor)
	strExpectedNextCursor := row.Cells[2].Value
	intExpectedNextCursor, _ := strconv.Atoi(strExpectedNextCursor)
	intObservedNextCursor, _ := strconv.Atoi(redirectsList.NextCursor)
	assert.True(&c.ErrorFeature, intExpectedNextCursor == intObservedNextCursor, "expected next cursor to equal "+strExpectedNextCursor+" but it is "+redirectsList.NextCursor)
	strExpectedTotalCount := row.Cells[3].Value
	intExpectedTotalCount, _ := strconv.Atoi(strExpectedTotalCount)
	intObservedTotalCount := redirectsList.TotalCount
	assert.True(&c.ErrorFeature, intExpectedTotalCount == intObservedTotalCount, "expected total count to equal "+strExpectedTotalCount+"but it is "+strconv.Itoa(intObservedTotalCount))
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

	return nil
}
