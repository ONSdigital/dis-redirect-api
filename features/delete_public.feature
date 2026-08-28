@DeletePublic
Feature: DELETE redirect endpoint for public mode

  Background: Service setup
    Given the redirect api is running

  Scenario: Return 405 when attempting to DELETE in public mode
    Given I am an admin user
    And redis is healthy
    When I DELETE "/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGg="
    Then the HTTP status code should be "405"
