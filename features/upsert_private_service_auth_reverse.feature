@UpsertReversePrivateServiceAuth
Feature: Upsert redirect endpoint for private mode with service auth and reverse lookup

  Background: Service setup
    Given service "dis-other-service" has the "redirects:edit" permission
    And private endpoints are enabled
    And reverse lookup is enabled
    And the redirect api is running
    And I am identified as "dis-other-service"

  Scenario: Upsert a redirect value via PUT if the key and value do not exist
    Given redis is healthy
    And I am authorised
    And redis contains no value for key "fwd:/economy/old-path"
    When I PUT "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGg="
        """
          {
            "from": "/economy/old-path",
            "to": "/economy/new-path"
          }
        """
    Then the HTTP status code should be "201"
    And the key "fwd:/economy/old-path" has a value of "/economy/new-path" in the Redis store
    And the key "rev:/economy/new-path" has a value of a set of the following values in the Redis store
      | values            |
      | /economy/old-path |

  Scenario: Upsert a redirect value via PUT if the key and value already exist
    Given redis is healthy
    And I am authorised
    And the key "fwd:/economy/old-path" is already set to a value of "/economy/new-path" in the Redis store
    And the key "rev:/economy/new-path" is already set to a value of a set of the following values in the Redis store
      | values            |
      | /economy/old-path |
    When I PUT "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGg="
        """
          {
            "from": "/economy/old-path",
            "to": "/economy/a-really-new-path"
          }
        """
    Then the HTTP status code should be "201"
    And the key "fwd:/economy/old-path" has a value of "/economy/a-really-new-path" in the Redis store
    And the key "rev:/economy/a-really-new-path" has a value of a set of the following values in the Redis store
      | values            |
      | /economy/old-path |
    And redis contains no value for key "rev:/economy/new-path"



