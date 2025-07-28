Feature: Delete redirect endpoint

  Background: Service setup
    Given an admin user has the "legacy:delete" permission
    And the redirect api is running

  Scenario: Delete a redirect if the key does not exist
    Given I am an admin user
    And redis is healthy
    And redis contains no value for key "/economy/old-path"
    When I DELETE "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGg="
    Then the HTTP status code should be "404"
    And I should receive the following response:
      """
      redirect not found
      """

  Scenario: Delete a redirect if the key exists
    Given I am an admin user
    And redis is healthy
    And the key "/economy/old-path" is already set to a value of "/economy/new-path" in the Redis store
    When I DELETE "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGg="
    Then the HTTP status code should be "204"
    And redis contains no value for key "/economy/old-path"

  Scenario: Delete a redirect with invalid base64 id
    Given I am an admin user
    And redis is healthy
    When I DELETE "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGgg=="
    Then the HTTP status code should be "400"
    Then I should receive the following response:
      """
      invalid base64 id
      """

  Scenario: Delete a redirect without the correct permission
    Given redis is healthy
    And I am not authenticated
    When I DELETE "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGg="
    Then the HTTP status code should be "401"


  Scenario: Server error when attempting to delete a redirect
    Given I am an admin user
    And redis stops running
    When I DELETE "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGg="
    Then the HTTP status code should be "500"
    And I should receive the following response:
      """
      failed to check redirect existence
      """