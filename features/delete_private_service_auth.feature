@DeletePrivateServiceAuth
Feature: DELETE redirect endpoint for private mode with service auth

  Background: Service setup
    Given service "dis-other-service" has the "redirects:delete" permission
    And private endpoints are enabled
    And the redirect api is running
    And I am identified as "dis-other-service"

  Scenario: Delete a redirect if the key does not exist
    Given I am authorised
    And redis is healthy
    And redis contains no value for key "/economy/old-path"
    When I DELETE "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGg="
    Then the HTTP status code should be "404"
    And I should receive the following response:
      """
      not found
      """

  Scenario: Delete a redirect if the key exists
    Given I am authorised
    And redis is healthy
    And the key "/economy/old-path" is already set to a value of "/economy/new-path" in the Redis store
    When I DELETE "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGg="
    Then the HTTP status code should be "204"
    And redis contains no value for key "/economy/old-path"
