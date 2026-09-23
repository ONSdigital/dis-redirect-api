@Reconcile
Feature: Database Reconciliation

  Background:
    Given redis is healthy

  Scenario: Rewrite to allow reverse lookup
    Given reverse lookup is enabled
    And the key "/economy/mypage" is already set to a value of "/newpage" in the Redis store
    And the redirect api is running
    Then the key "fwd:/economy/mypage" has a value of "/newpage" in the Redis store   
    And the key "rev:/newpage" has a value of a set of the following values in the Redis store
      | values          |
      | /economy/mypage |

  Scenario: Rewrite to allow forward lookup only
    Given reverse lookup is disabled
    And the key "fwd:/economy/mypage" is already set to a value of "/newpage" in the Redis store
    And the key "rev:/newpage" is already set to a value of a set of the following values in the Redis store
      | values          |
      | /economy/mypage |
    And the redirect api is running
    Then the key "/economy/mypage" has a value of "/newpage" in the Redis store   
    And redis contains no value for key "rev:/newpage"
