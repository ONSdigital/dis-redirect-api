@UpsertRedirect @UpsertRedirectService
Feature: Upsert redirect endpoint with service auth

  Background: Service setup
    Given service "dis-other-service" has the "redirects:edit" permission
    And the redirect api is running
    And I am identified as "dis-other-service"

  Scenario: Upsert a redirect value via PUT if the key and value do not exist
    Given redis is healthy
    And I am authorised
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

  Scenario: Upsert a redirect value via PUT if the key and value already exist
    Given redis is healthy
    And I am authorised
    And the key "/economy/old-path" is already set to a value of "/economy/new-path" in the Redis store
    When I PUT "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGg="
        """
          {
            "from": "/economy/old-path",
            "to": "/economy/new-path"
          }
        """
    Then the HTTP status code should be "200"
    And the key "/economy/old-path" has a value of "/economy/new-path" in the Redis store

  Scenario: Upsert a redirect value via PUT with invalid base64 id
    Given redis is healthy
    And I am authorised
    When I PUT "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGgg=="
        """
          {
            "from": "/economy/old-path",
            "to": "/economy/new-path"
          }
        """
    Then the HTTP status code should be "400"

  Scenario: Upsert a redirect value via PUT with invalid body
    Given redis is healthy
    And I am authorised
    When I PUT "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGg=="
        """
          {
            "key": "/economy/old-path",
            "value": "/economy/new-path"
          }
        """
    Then the HTTP status code should be "400"

  Scenario: Upsert a redirect value via PUT without the correct permission
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

