Feature: Redirect endpoint
    Scenario: Return the value when the key exists in redis
        Given the key "/economy/old-path" is already set to a value of "/economy/new-path" in the Redis store
        And the redirect api is running
        When I GET "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGg="
        Then I should receive the following JSON response with status "200":
            """
            {
                "from": "/economy/old-path",
                "to": "/economy/new-path"
            }
            """


    Scenario: Return 400 when the key is not base64
    Given redis is healthy
        And the redirect api is running
        When I GET "/v1/redirects/cheese"
        Then the HTTP status code should be "400"
        And I should receive the following response:
            """
                invalid base64 id
            """


    Scenario: Return 404 when the key is not found
        Given redis is healthy
        And the redirect api is running
        When I GET "/v1/redirects/b2xkLXBhdGg="
        Then the HTTP status code should be "404"
        And I should receive the following response:
            """
                key old-path not found
            """


    Scenario: Return 500 when redis returns an error
        Given redis stops running
        And the redirect api is running
        And I wait 4 seconds to pass the critical timeout
        When I GET "/v1/redirects/b2xkLXBhdGg="
        Then the HTTP status code should be "500"
        And I should receive the following response:
            """
                redis returned an error
            """

  Scenario: Upsert a redirect value via PUT if the key and value do not exist
    Given I am an admin user
    And the redirect api is running
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

  Scenario: Upsert a redirect value via PUT if the key and value already exist
    Given I am an admin user
    And the redirect api is running
    And redis is healthy
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
    Given I am an admin user
    And the redirect api is running
    And redis is healthy
    When I PUT "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGgg=="
      """
        {
          "from": "/economy/old-path",
          "to": "/economy/new-path"
        }
      """
    Then the HTTP status code should be "400"

  Scenario: Upsert a redirect value via PUT with invalid body
    Given I am an admin user
    And the redirect api is running
    And redis is healthy
    When I PUT "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGg=="
      """
        {
          "key": "/economy/old-path",
          "value": "/economy/new-path"
        }
      """
    Then the HTTP status code should be "400"

  Scenario: Upsert a redirect value via PUT without the correct permission
    Given the redirect api is running
    And redis is healthy
    When I PUT "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGg="
      """
        {
          "from": "/economy/old-path",
          "to": "/economy/new-path"
        }
      """
    Then the HTTP status code should be "401"

