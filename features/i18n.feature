Feature: i18n Language Switching
  As a Dobby desktop app user
  I want to switch between Traditional Chinese and English
  So that I can use the app in my preferred language

  Background:
    Given the app is running

  # ─── Language Persistence ─────────────────────────────────────────

  @S-1
  Scenario: Default language is Traditional Chinese on first launch
    Given no language preference is stored in localStorage
    When the app loads
    Then the UI is displayed in Traditional Chinese

  @S-2
  Scenario: Language preference persists after restart
    Given the user has selected "en" in Settings
    When the app is restarted
    Then the UI is displayed in English

  @S-3
  Scenario: Language falls back to zh-TW when localStorage value is invalid
    Given localStorage contains an unrecognized language code "fr"
    When the app loads
    Then the UI is displayed in Traditional Chinese

  # ─── Language Switcher UI ─────────────────────────────────────────

  @S-4
  Scenario: Language switcher appears in Settings above Notifications
    Given the app is on the Settings page
    Then a "Language" card section is visible above the "Notifications" section

  @S-5
  Scenario: Language switcher shows current language as selected
    Given localStorage contains language "zh-TW"
    When the user opens Settings
    Then the language select shows "繁體中文" as the selected option

  @S-6
  Scenario: Language switcher shows English as selected when en is active
    Given localStorage contains language "en"
    When the user opens Settings
    Then the language select shows "English" as the selected option

  # ─── Instant Language Switch ──────────────────────────────────────

  @S-7
  Scenario: Switching to English updates all UI text immediately
    Given the current language is "zh-TW"
    When the user selects "en" in the Settings language dropdown
    Then all navigation labels switch to English without page reload
    And localStorage stores "en"

  @S-8
  Scenario: Switching back to Traditional Chinese updates all UI text immediately
    Given the current language is "en"
    When the user selects "zh-TW" in the Settings language dropdown
    Then all navigation labels switch to Traditional Chinese without page reload
    And localStorage stores "zh-TW"

  # ─── Translation Coverage ─────────────────────────────────────────

  @S-9
  Scenario: All nav items are translated
    When the language is "en"
    Then the sidebar shows "Dashboard", "Rules", "Activity Logs", "Settings"

  @S-10
  Scenario: Dashboard page uses translated strings
    When the language is "en"
    Then Dashboard shows "Today's Processed", "Active Rules", "Recent Activity"

  @S-11
  Scenario: Rules page title and subtitle use interpolation
    Given there are 3 rules with 2 active
    When the language is "en"
    Then the subtitle reads "3 rules, 2 active"

  @S-12
  Scenario: Logs page subtitle uses interpolation
    Given there are 10 logs with 8 successful
    When the language is "en"
    Then the subtitle reads "10 records · 8 successful"

  @S-13
  Scenario: License section uses translated strings
    Given the license status is "active" with 7 days remaining
    When the language is "en"
    Then the license status reads "Trial — 7 days remaining"

  @S-14
  Scenario: Trial expired banner uses translated strings
    Given the license status is "expired"
    When the language is "en"
    Then the top banner reads "Trial period has ended"
    And the banner link reads "Enter License Key →"

  # ─── TypeScript Key Safety ────────────────────────────────────────

  @S-15
  Scenario: zh-TW and en translation files have identical keys
    Given the zh-TW and en locale files are loaded
    Then every key present in zh-TW exists in en
    And every key present in en exists in zh-TW
