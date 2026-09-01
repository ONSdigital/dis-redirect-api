@UpsertPrivateBadAuth
Feature: Upsert redirect endpoint for private mode with invalid auth

  Background: Service setup
    Given private endpoints are enabled
    And the redirect api is running

  Scenario: Return 401 for PUT /v1/redirects/{id} with no auth
    Given redis is healthy
    And I am not authorised
    When I PUT "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGg="
      """
      {
        "from": "/economy/old-path",
        "to": "/economy/new-path"
      }
      """
    Then the HTTP status code should be "401"

  Scenario: Return 401 for PUT /v1/redirects/{id} with a bad JWT token
    Given redis is healthy
    And I set the "Authorization" header to "Bearer bad.jwt.token"
    When I PUT "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGg="
      """
      {
        "from": "/economy/old-path",
        "to": "/economy/new-path"
      }
      """
    Then the HTTP status code should be "401"
