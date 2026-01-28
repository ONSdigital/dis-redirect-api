@GetRedirect
Feature: Redirect endpoint

  Background: Service setup
    Given an admin user has the "redirects:read" permission
    And the redirect api is running

  # TODO Update the href value to be "http://localhost:29900/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGg=" when dp-net has been fixed
  Scenario: Return the value when the key exists in redis
    Given I am an admin user
    And the key "/economy/old-path" is already set to a value of "/economy/new-path" in the Redis store
    And redis is healthy
    When I GET "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGg="
    Then I should receive the following JSON response with status "200":
        """
        {
            "from": "/economy/old-path",
            "to": "/economy/new-path",
            "id": "L2Vjb25vbXkvb2xkLXBhdGg=",
            "links": {
                "self": {
                    "href": "http://localhost:29900/redirects/L2Vjb25vbXkvb2xkLXBhdGg=",
                    "id": "L2Vjb25vbXkvb2xkLXBhdGg="
                }
              }
            }
        """

  Scenario: Return 400 when the key is not base64
    Given I am an admin user
    And redis is healthy
    When I GET "/v1/redirects/cheese"
    Then the HTTP status code should be "400"
    And I should receive the following response:
        """
            the base64 id provided is invalid
        """

  Scenario: Return 404 when the key is not found
    Given I am an admin user
    And redis is healthy
    When I GET "/v1/redirects/b2xkLXBhdGg="
    Then the HTTP status code should be "404"
    And I should receive the following response:
        """
            not found
        """

  Scenario: Return 500 when redis returns an error
    Given I am an admin user
    And redis stops running
    And I wait 4 seconds to pass the critical timeout
    When I GET "/v1/redirects/b2xkLXBhdGg="
    Then the HTTP status code should be "500"
    And I should receive the following response:
        """
            internal error
        """

  Scenario: Return all the redirects that exist in redis using default path parameters
    Given I am an admin user
    And the key "/economy/old-path1" is already set to a value of "/economy/new-path1" in the Redis store
    And the key "/economy/old-path2" is already set to a value of "/economy/new-path2" in the Redis store
    And the key "/economy/old-path3" is already set to a value of "/economy/new-path3" in the Redis store
    When I GET "/v1/redirects"
    Then the HTTP status code should be "200"
    And I would expect there to be three or more redirects returned in a list
    And in each redirect I would expect the response to contain values that have these structures
      | from              | Not empty string                              |
      | to                | Not empty string                              |
      | id                | 'from' value encoded as Base64 string         |
      | links: self: href | https://api.beta.ons.gov.uk/v1/redirects/{id} |
      | links: self: id   | {id}                                          |
    And the list of redirects should also contain the following values:
      | count | cursor | next_cursor | total_count |
      | 10    | 0      | 0           | 3           |

  Scenario: Return all the redirects that exist in redis using specific valid path parameters
    Given I am an admin user
    And the key "/economy/old-path1" is already set to a value of "/economy/new-path1" in the Redis store
    And the key "/economy/old-path2" is already set to a value of "/economy/new-path2" in the Redis store
    And the key "/economy/old-path3" is already set to a value of "/economy/new-path3" in the Redis store
    When I GET "/v1/redirects?count=2&cursor=1"
    Then the HTTP status code should be "200"
    And I would expect there to be 2 redirects returned in a list
    And in each redirect I would expect the response to contain values that have these structures
      | from              | Not empty string                              |
      | to                | Not empty string                              |
      | id                | 'from' value encoded as Base64 string         |
      | links: self: href | https://api.beta.ons.gov.uk/v1/redirects/{id} |
      | links: self: id   | {id}                                          |
    And the list of redirects should also contain the following values:
      | count | cursor | next_cursor | total_count |
      | 2     | 1      | 0           | 3           |

  Scenario: Return 400 when the count value given is not an integer
    Given I am an admin user
    And redis is healthy
    When I GET "/v1/redirects?count=not-a-number"
    Then the HTTP status code should be "400"
    And I should receive the following response:
        """
            the count must be an integer giving the requested number of redirects
        """

  Scenario: Return 400 when the count value given is negative
    Given I am an admin user
    And redis is healthy
    When I GET "/v1/redirects?count=-5"
    Then the HTTP status code should be "400"
    And I should receive the following response:
            """
                the count must be a positive integer
            """

  Scenario: Return 400 when the cursor value given is not an integer
    Given I am an admin user
    And redis is healthy
    When I GET "/v1/redirects?cursor=not-a-number"
    Then the HTTP status code should be "400"
    And I should receive the following response:
        """
            the redirects cursor was invalid. It must be a positive integer
        """

  Scenario: Return 400 when the cursor value given is negative
    Given I am an admin user
    And redis is healthy
    When I GET "/v1/redirects?cursor=-6"
    Then the HTTP status code should be "400"
    And I should receive the following response:
        """
            the redirects cursor was invalid. It must be a positive integer
        """

  Scenario: Return 500 when calling get redirects with redis not running
    Given I am an admin user
    And redis stops running
    And I wait 4 seconds to pass the critical timeout
    When I GET "/v1/redirects"
    Then the HTTP status code should be "500"
    And I should receive the following response:
        """
            internal error
        """
