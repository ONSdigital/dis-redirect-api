@UpsertPublic
Feature: Upsert redirect endpoint for public mode

  Background: Service setup
    Given the redirect api is running

  Scenario: Return 405 when attempting to PUT in public mode
    Given I am an admin user
    And redis is healthy
    When I PUT "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGg="
      """
      {
        "from": "/economy/old-path",
        "to": "/economy/new-path"
      }
      """
    Then the HTTP status code should be "405"
