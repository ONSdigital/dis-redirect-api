@DeleteReversePrivateServiceAuth
Feature: DELETE redirect endpoint for private mode with service auth and reverse lookup

  Background: Service setup
    Given service "dis-other-service" has the "redirects:delete" permission
    And private endpoints are enabled
    And reverse lookup is enabled
    And the redirect api is running
    And I am identified as "dis-other-service"

  Scenario: Delete a redirect if the forward key does not exist
    Given I am authorised
    And redis is healthy
    And redis contains no value for key "fwd:/economy/old-path"
    When I DELETE "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGg="
    Then the HTTP status code should be "404"
    And I should receive the following response:
      """
      not found
      """

  Scenario: Delete a redirect and its reverse lookup entry
    Given I am authorised
    And redis is healthy
    And the key "fwd:/economy/old-path" is already set to a value of "/economy/new-path" in the Redis store
    And the key "rev:/economy/new-path" is already set to a value of a set of the following values in the Redis store
      | values            |
      | /economy/old-path |
    When I DELETE "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGg="
    Then the HTTP status code should be "204"
    And redis contains no value for key "fwd:/economy/old-path"
    And redis contains no value for key "rev:/economy/new-path"
