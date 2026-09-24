@GetRedirectPrivate
Feature: GET redirect endpoint for private mode with to filter

  Background: Service setup
    Given an admin user has the "redirects:read" permission
    And reverse lookup is enabled
    And the redirect api is running

  Scenario: Return all the redirects that exist in redis using specific valid path parameters
    Given I am an admin user
    And the key "fwd:/economy/old-path1" is already set to a value of "/economy/new-path1" in the Redis store
    And the key "fwd:/economy/old-path2" is already set to a value of "/economy/new-path1" in the Redis store
    And the key "fwd:/economy/old-path3" is already set to a value of "/economy/new-path2" in the Redis store
    And the key "rev:/economy/new-path1" is already set to a value of a set of the following values in the Redis store
      | values             |
      | /economy/old-path1 |
      | /economy/old-path2 |
    When I GET "/v1/redirects?count=2&cursor=1&to=/economy/new-path1"
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
      | 2     | 1      | 0           | 2           |

  Scenario: Return 400 when the to value given is not a valid relative path
    Given I am an admin user
    And redis is healthy
    When I GET "/v1/redirects?to=invalid-path"
    Then the HTTP status code should be "400"
    And I should receive the following response:
        """
            the 'to' query parameter must be a valid relative path
        """
