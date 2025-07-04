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

    Scenario: Return all the redirects that exist in redis
        Given the key "/economy/old-path1" is already set to a value of "/economy/new-path1" in the Redis store
        And the key "/economy/old-path2" is already set to a value of "/economy/new-path2" in the Redis store
        And the key "/economy/old-path3" is already set to a value of "/economy/new-path3" in the Redis store
        And the redirect api is running
        When I GET "/v1/redirects"
        Then I should receive the following JSON response with status "200":
            """
            {
              "count": 10,
              "items": [
                {
                  "from": "/economy/old-path1",
                  "to": "/economy/new-path1",
                  "id": "L2Vjb25vbXkvb2xkLXBhdGgx",
                  "links": {
                    "self": {
                      "href": "https://api.beta.ons.gov.uk/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGgx",
                      "id": "L2Vjb25vbXkvb2xkLXBhdGgx"
                    }
                  }
                },
                {
                  "from": "/economy/old-path2",
                  "to": "/economy/new-path2",
                  "id": "L2Vjb25vbXkvb2xkLXBhdGgy",
                  "links": {
                    "self": {
                      "href": "https://api.beta.ons.gov.uk/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGgy",
                      "id": "L2Vjb25vbXkvb2xkLXBhdGgy"
                    }
                  }
                },
                {
                  "from": "/economy/old-path3",
                  "to": "/economy/new-path3",
                  "id": "L2Vjb25vbXkvb2xkLXBhdGgz",
                  "links": {
                    "self": {
                      "href": "https://api.beta.ons.gov.uk/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGgz",
                      "id": "L2Vjb25vbXkvb2xkLXBhdGgz"
                    }
                  }
                }
              ],
              "cursor": "0",
              "next_cursor": "0",
              "total_count": 3
            }
            """
