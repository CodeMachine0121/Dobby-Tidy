Feature: Monetization — Trial Period and License Activation
  As a Dobby user
  I want a 14-day free trial with the option to purchase a license
  So that I can evaluate the product before paying

  @S-1
  Scenario: First launch initializes a 14-day trial
    Given the app has never been launched before
    When the app initializes on first launch
    Then the trial start date is recorded locally
    And the license status is "active"
    And 14 days remain in the trial

  @S-2
  Scenario: Trial is still active within the 14-day window
    Given the trial started 7 days ago
    When the license status is checked
    Then the license status is "active"
    And the background processor is allowed to run

  @S-3
  Scenario: Trial expires after 14 days
    Given the trial started 15 days ago
    When the license status is checked
    Then the license status is "expired"
    And the background processor is blocked

  @S-4
  Scenario: Background processor is blocked when trial is expired
    Given the trial has expired
    When the background processor attempts to scan
    Then the scan is skipped without processing any files

  @S-5
  Scenario: Valid license key activates the app
    Given the trial is still active
    When the user activates a valid license key
    Then the license status is "activated"
    And the background processor is allowed to run

  @S-6
  Scenario: Invalid license key checksum is rejected
    Given the user wants to activate a license
    When the user activates the license key "DOBBY-ABCD-EFGH-ZZZZ"
    Then activation fails with error "invalid license key"

  @S-7
  Scenario: Malformed license key format is rejected
    Given the user wants to activate a license
    When the user activates the license key "INVALID-FORMAT"
    Then activation fails with error "invalid license key format"

  @S-8
  Scenario: Activated license allows background processor even after trial expires
    Given the trial started 20 days ago
    And the license has been activated with a valid key
    When the license status is checked
    Then the license status is "activated"
    And the background processor is allowed to run
