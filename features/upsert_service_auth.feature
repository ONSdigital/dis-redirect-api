@ServiceAuth
Feature: Upsert redirect endpoint with user auth

  Background: Service setup
    Given service "users/zebedee" has the "legacy:edit" permission
    And the redirect api is running

  Scenario: Upsert a redirect value via PUT if the key and value do not exist
    Given I am identified as "zebedee"
    And I use a service auth token "fake-service-auth-token"
    And redis is healthy
    And redis contains no value for key "/economy/old-path"
    When I PUT "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGg="
      """
      {
        "from": "/economy/old-path",
        "to": "/economy/new-path"
      }
      """
    Then the HTTP status code should be "201"
    And the key "/economy/old-path" has a value of "/economy/new-path" in the Redis store
