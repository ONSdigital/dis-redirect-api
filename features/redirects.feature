Feature: Redirect endpoint
    Scenario: Return the value when the key exists in redis
        Given the key "/economy/old-path" is already set to a value of "/economy/new-path" in the Redis store
        And the redirect api is running
        When I GET "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGg="
        Then I should receive the following JSON response with status "200":
            """
            {
                "from": "/economy/old-path",
                "to": "/economy/new-path",
                "id": "",
                "links": {
                    "self": {
                        "href": "",
                        "id": ""
                    }
                  }
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

    Scenario: Return all the redirects that exist in redis using default path parameters
        Given the key "/economy/old-path1" is already set to a value of "/economy/new-path1" in the Redis store
        And the key "/economy/old-path2" is already set to a value of "/economy/new-path2" in the Redis store
        And the key "/economy/old-path3" is already set to a value of "/economy/new-path3" in the Redis store
        And the redirect api is running
        When I GET "/v1/redirects"
        Then the HTTP status code should be "200"
        And I would expect there to be three or more redirects returned in a list
        And in each redirect I would expect the response to contain values that have these structures
            | from                   | Not empty string                              |
            | to                     | Not empty string                              |
            | id                     | 'from' value encoded as Base64 string         |
            | links: self: href      | https://api.beta.ons.gov.uk/v1/redirects/{id} |
            | links: self: id        | {id}                                          |
        And the list of redirects should also contain the following values:
            | count                  | 10                           |
            | cursor                 | 0                            |
            | next_cursor            | 0                            |
            | total_count            | 3                            |

    Scenario: Return all the redirects that exist in redis using specific valid path parameters
        Given the key "/economy/old-path1" is already set to a value of "/economy/new-path1" in the Redis store
        And the key "/economy/old-path2" is already set to a value of "/economy/new-path2" in the Redis store
        And the key "/economy/old-path3" is already set to a value of "/economy/new-path3" in the Redis store
        And the redirect api is running
        When I GET "/v1/redirects?count=2&cursor=1"
        Then the HTTP status code should be "200"
#        And I would expect there to be two redirects returned in a list
#        And in each redirect I would expect the response to contain values that have these structures
#            | from                   | Not empty string                              |
#            | to                     | Not empty string                              |
#            | id                     | 'from' value encoded as Base64 string         |
#            | links: self: href      | https://api.beta.ons.gov.uk/v1/redirects/{id} |
#            | links: self: id        | {id}                                          |
#        And the list of redirects should also contain the following values:
#            | count                  | 2                            |
#            | cursor                 | 1                            |
#            | next_cursor            | 2                            |
#            | total_count            | 3                            |
