Feature: Gumroad License Verification — 以 Gumroad API 驗證並啟用 license

  Background:
    Given app 尚未啟用 license

  @S-1
  Scenario: 有效 key 首次啟用成功
    Given Gumroad API 回傳 success=true 且 uses=1
    When 使用者輸入有效的 Gumroad license key 並送出
    Then license 狀態變為 activated
    And key 被儲存至本機

  @S-2
  Scenario: 無效 key 被拒絕
    Given Gumroad API 回傳 success=false
    When 使用者輸入無效的 license key 並送出
    Then 回傳「無效的 license key」錯誤
    And license 狀態維持不變

  @S-3
  Scenario: 已在其他機器使用的 key 被拒絕
    Given Gumroad API 回傳 success=true 且 uses=2
    When 使用者輸入該 license key 並送出
    Then 回傳「此 key 已在另一台機器上使用」錯誤
    And license 狀態維持不變

  @S-4
  Scenario: 網路失敗時回傳明確錯誤
    Given Gumroad API 無法連線
    When 使用者輸入 license key 並送出
    Then 回傳「無法連線至驗證伺服器，請確認網路連線」錯誤
    And license 狀態維持不變

  @S-5
  Scenario: 已啟用的 license 不可重複啟用
    Given license 已在本機啟用
    When 使用者再次輸入任意 license key 並送出
    Then 回傳「license 已啟用」錯誤
    And 不呼叫 Gumroad API

  @S-6
  Scenario: 啟用後再次啟動 app 不需網路
    Given license 已在本機啟用
    When app 啟動並讀取 license 狀態
    Then license 狀態為 activated
    And 不呼叫 Gumroad API
