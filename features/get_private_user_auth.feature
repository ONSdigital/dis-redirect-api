@GetRedirectPrivateUserAuth
Feature: GET redirect endpoint for private mode with user auth

  Background: Service setup
    Given an admin user has the "redirects:read" permission
    And private endpoints are enabled
    And the redirect api is running

  Scenario: Return 200 for GET /v1/redirects/{id} with valid user auth
    Given I am an admin user
    And the key "/economy/old-path" is already set to a value of "/economy/new-path" in the Redis store
    And redis is healthy
    When I GET "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGg="
    Then the HTTP status code should be "200"

  Scenario: Return 200 for GET /v1/redirects with valid user auth
    Given I am an admin user
    And the key "/economy/old-path" is already set to a value of "/economy/new-path" in the Redis store
    And redis is healthy
    When I GET "/v1/redirects"
    Then the HTTP status code should be "200"
