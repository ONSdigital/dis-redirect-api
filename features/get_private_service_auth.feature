@GetRedirectPrivateServiceAuth
Feature: GET redirect endpoint for private mode with service auth

  Background: Service setup
    Given service "dis-other-service" has the "redirects:read" permission
    And private endpoints are enabled
    And the redirect api is running
    And I am identified as "dis-other-service"

  Scenario: Return 200 for GET /v1/redirects/{id} with valid service auth
    Given I am authorised
    And the key "/economy/old-path" is already set to a value of "/economy/new-path" in the Redis store
    And redis is healthy
    When I GET "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGg="
    Then the HTTP status code should be "200"

  Scenario: Return 200 for GET /v1/redirects with valid service auth
    Given I am authorised
    And the key "/economy/old-path" is already set to a value of "/economy/new-path" in the Redis store
    And redis is healthy
    When I GET "/v1/redirects"
    Then the HTTP status code should be "200"
