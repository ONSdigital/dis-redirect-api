@GetRedirectPrivate
Feature: GET redirect endpoint for private mode with reverse lookup enabled

  Background: Service setup
    Given an admin user has the "redirects:read" permission
    And reverse lookup is enabled
    And the redirect api is running

  # TODO Update the href value to be "http://localhost:29900/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGg=" when dp-net has been fixed
  Scenario: Return the redirect without the fwd prefix when the forward lookup key exists in redis
    Given I am an admin user
    And the key "fwd:/economy/old-path" is already set to a value of "/economy/new-path" in the Redis store
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

  Scenario: Return 404 when only an unprefixed key exists in redis
    Given I am an admin user
    And the key "/economy/only-legacy" is already set to a value of "/economy/new-path" in the Redis store
    And redis is healthy
    When I GET "/v1/redirects/L2Vjb25vbXkvb25seS1sZWdhY3k="
    Then the HTTP status code should be "404"
    And I should receive the following response:
        """
            not found
        """
