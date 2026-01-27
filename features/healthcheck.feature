@Healthcheck
Feature: Health endpoint

  Rule: Redis is healthy
    Background:
      Given redis is healthy

    @HealthcheckOK
    Scenario: Returning a OK (200) status when health endpoint called
      Given the redirect api is running
      And I have a healthcheck interval of 1 second
      And I wait 2 seconds for the healthcheck to be available
      When I GET "/health"
      Then the health checks should have completed within 6 seconds
      And I should receive the following health JSON response:
        """
        {
          "status": "OK",
          "version": {
            "git_commit": "6584b786caac36b6214ffe04bf62f058d4021538",
            "language": "go",
            "language_version": "go1.24.2",
            "version": "v1.2.3"
          },
          "checks": [
            {
              "name": "Redis",
              "status": "OK",
              "status_code": 200,
              "message": "redis is healthy"
            }
          ]
        }
        """

  Rule: Redis is unhealthy
    Background:
      Given redis stops running

    @HealthcheckWarning
    Scenario: Returning a WARNING (429) status when health endpoint called
      Given the redirect api is running
      And I have a healthcheck interval of 1 second
      And I wait 4 seconds for the healthcheck to be available
      When I GET "/health"
      Then the HTTP status code should be "429"
      And the response header "Content-Type" should be "application/json; charset=utf-8"
      And the health checks should have completed within 5 seconds
      And I should receive the following health JSON response:
        """
        {
          "status": "WARNING",
          "version": {
            "git_commit": "6584b786caac36b6214ffe04bf62f058d4021538",
            "language": "go",
            "language_version": "go1.17.8",
            "version": "v1.2.3"
          },
          "checks": [
            {
              "name": "Redis",
              "status": "CRITICAL",
              "status_code": 500,
              "message": "connect: connection refused"
            }
          ]
        }
        """

    @HealthcheckCritical
    Scenario: Returning a CRITICAL (500) status when health endpoint called
      Given the redirect api is running
      And I have a healthcheck interval of 1 second
      And I wait 8 seconds to pass the critical timeout
      When I GET "/health"
      Then the HTTP status code should be "500"
      And the response header "Content-Type" should be "application/json; charset=utf-8"
      And the health checks should have completed within 9 seconds
      Then I should receive the following health JSON response:
        """
        {
          "status": "CRITICAL",
          "version": {
            "git_commit": "6584b786caac36b6214ffe04bf62f058d4021538",
            "language": "go",
            "language_version": "go1.17.8",
            "version": "v1.2.3"
          },
          "checks": [
            {
              "name": "Redis",
              "status": "CRITICAL",
              "status_code": 500,
              "message": "connect: connection refused"
            }
          ]
        }
        """
