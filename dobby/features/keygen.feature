Feature: Keygen CLI — 批次產生 DOBBY license keys

  Background:
    Given 工具可存取 license key 驗證器

  @S-1
  Scenario: 以預設數量產生 keys 並輸出到 stdout
    Given 使用者未指定 -n 與 -o 旗標
    When 執行 keygen 工具
    Then 輸出 10000 行到 stdout
    And 每行格式符合 DOBBY-XXXX-XXXX-XXXX

  @S-2
  Scenario: 以指定數量產生 keys 並輸出到 stdout
    Given 使用者指定 -n 50
    When 執行 keygen 工具
    Then 輸出恰好 50 行到 stdout
    And 每行格式符合 DOBBY-XXXX-XXXX-XXXX

  @S-3
  Scenario: 將 keys 輸出到指定檔案
    Given 使用者指定 -n 100 與 -o keys.txt
    When 執行 keygen 工具
    Then keys.txt 包含恰好 100 行
    And 每行格式符合 DOBBY-XXXX-XXXX-XXXX

  @S-4
  Scenario: 同一批次產生的 keys 不重複
    Given 使用者指定 -n 1000
    When 執行 keygen 工具
    Then 輸出的 1000 個 keys 全部唯一

  @S-5
  Scenario: 每個 key 通過 HMAC 驗證
    Given 使用者指定 -n 10
    When 執行 keygen 工具
    Then 每個輸出的 key 皆可被 LicenseKeyValidator.Validate 接受

  @S-6
  Scenario: 指定 -n 為 0 時輸出空檔案
    Given 使用者指定 -n 0
    When 執行 keygen 工具
    Then 輸出 0 行（空輸出）

  @S-7
  Scenario: 指定 -n 為負數時回傳錯誤
    Given 使用者指定 -n -1
    When 執行 keygen 工具
    Then 工具以非零狀態碼結束
    And stderr 包含錯誤訊息
