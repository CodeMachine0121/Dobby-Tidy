# 檔案管家（File Manager）行為規格
# 涵蓋核心領域邏輯：規則管理、規則比對、命名樣版渲染、處理流程狀態機

Feature: 檔案管家 — 自動命名與歸檔
  As a 使用者
  I want 設定規則自動重新命名並移動新進檔案
  So that 不需手動整理工作資料夾

  # ─────────────────────────────────────────────────────────────
  # Rule 管理
  # ─────────────────────────────────────────────────────────────

  @S-1
  Scenario: 成功建立一條完整規則
    Given 系統中尚無任何規則
    When 使用者建立一條規則，監控資料夾為 "~/Downloads"，命名樣版為 "{project}-{type}-{YYYY}{MM}{DD}-{seq}.{ext}"，目標資料夾為 "~/Projects/design/"，專案名稱為 "my-app"，類型標籤為 "screenshot"，副檔名過濾 [".png", ".jpg"]
    Then 規則應被成功建立
    And 規則的 enabled 狀態為 true
    And 規則擁有唯一的 RuleId

  @S-2
  Scenario: 監控資料夾路徑重複時建立規則失敗
    Given 系統中已有一條規則監控資料夾 "~/Downloads"
    When 使用者嘗試再建立一條規則，監控資料夾同為 "~/Downloads"
    Then 應回傳錯誤，說明該監控資料夾路徑已被使用

  @S-3
  Scenario: 停用一條規則
    Given 系統中有一條 enabled 為 true 的規則，RuleId 為 "rule-001"
    When 使用者停用規則 "rule-001"
    Then 規則 "rule-001" 的 enabled 狀態應變為 false

  @S-4
  Scenario: 啟用一條已停用的規則
    Given 系統中有一條 enabled 為 false 的規則，RuleId 為 "rule-002"
    When 使用者啟用規則 "rule-002"
    Then 規則 "rule-002" 的 enabled 狀態應變為 true

  @S-5
  Scenario: 刪除一條規則
    Given 系統中有一條規則，RuleId 為 "rule-003"
    When 使用者刪除規則 "rule-003"
    Then 系統中應不再存在 RuleId 為 "rule-003" 的規則

  # ─────────────────────────────────────────────────────────────
  # RuleMatcher — 規則比對
  # ─────────────────────────────────────────────────────────────

  @S-6
  Scenario: 檔案副檔名符合過濾條件且無關鍵字限制時比對成功
    Given 一條規則的 FilterSpec 副檔名過濾為 [".png", ".jpg"]，無關鍵字條件
    When 偵測到一個新檔案，副檔名為 ".png"，檔名為 "screenshot-001.png"
    Then RuleMatcher 應回傳 true（比對成功）

  @S-7
  Scenario: 檔案副檔名不符合過濾條件時比對失敗
    Given 一條規則的 FilterSpec 副檔名過濾為 [".png", ".jpg"]
    When 偵測到一個新檔案，副檔名為 ".pdf"，檔名為 "report.pdf"
    Then RuleMatcher 應回傳 false（比對失敗）

  @S-8
  Scenario: 副檔名符合且關鍵字也符合時比對成功
    Given 一條規則的 FilterSpec 副檔名過濾為 [".png"]，關鍵字為 "export"
    When 偵測到一個新檔案，副檔名為 ".png"，檔名為 "figma-export-v2.png"
    Then RuleMatcher 應回傳 true（比對成功）

  @S-9
  Scenario: 副檔名符合但關鍵字不符合時比對失敗
    Given 一條規則的 FilterSpec 副檔名過濾為 [".png"]，關鍵字為 "export"
    When 偵測到一個新檔案，副檔名為 ".png"，檔名為 "screenshot.png"
    Then RuleMatcher 應回傳 false（比對失敗）

  @S-10
  Scenario: FilterSpec 未設定任何副檔名過濾（空清單）時比對任何副檔名均成功
    Given 一條規則的 FilterSpec 副檔名過濾為空清單，無關鍵字條件
    When 偵測到一個新檔案，副檔名為 ".pdf"
    Then RuleMatcher 應回傳 true（比對成功）

  # ─────────────────────────────────────────────────────────────
  # TemplateRenderer — 命名樣版渲染
  # ─────────────────────────────────────────────────────────────

  @S-11
  Scenario: 渲染包含所有支援變數的命名樣版
    Given 一個命名樣版為 "{project}-{type}-{YYYY}{MM}{DD}-{seq}.{ext}"
    And ProcessingContext 為：project="my-app"，typeLabel="screenshot"，date=2026-04-16，seq="001"，originalName="Untitled"，extension="png"
    When 執行樣版渲染
    Then 渲染結果應為 "my-app-screenshot-20260416-001.png"

  @S-12
  Scenario: 渲染包含 {original} 變數的命名樣版
    Given 一個命名樣版為 "{original}-{YYYY}{MM}{DD}.{ext}"
    And ProcessingContext 為：originalName="report-draft"，date=2026-04-16，extension="pdf"
    When 執行樣版渲染
    Then 渲染結果應為 "report-draft-20260416.pdf"

  @S-13
  Scenario: 渲染目標路徑樣版（含動態 {project} 路徑）
    Given 一個目標路徑樣版為 "~/Projects/{project}/assets/"
    And ProcessingContext 為：project="my-app"
    When 執行路徑樣版渲染
    Then 渲染結果應為 "~/Projects/my-app/assets/"

  # ─────────────────────────────────────────────────────────────
  # SequenceGenerator — 當日序號產生
  # ─────────────────────────────────────────────────────────────

  @S-14
  Scenario: 當日尚無任何成功紀錄時，序號從 001 開始
    Given 今天 2026-04-16 該規則尚無成功操作紀錄
    When 為規則 "rule-001" 在 2026-04-16 產生序號
    Then 序號應為 "001"

  @S-15
  Scenario: 當日已有 3 筆成功紀錄時，序號為 004
    Given 今天 2026-04-16 規則 "rule-001" 已有 3 筆成功操作紀錄
    When 為規則 "rule-001" 在 2026-04-16 產生序號
    Then 序號應為 "004"

  # ─────────────────────────────────────────────────────────────
  # ProcessingJob 狀態機
  # ─────────────────────────────────────────────────────────────

  @S-16
  Scenario: 新建立的 ProcessingJob 初始狀態為 pending
    Given 偵測到一個新檔案事件，路徑為 "~/Downloads/Untitled.png"
    When 系統建立一個新的 ProcessingJob
    Then Job 的狀態應為 "pending"
    And Job 擁有唯一的 JobId

  @S-17
  Scenario: RuleMatcher 比對成功後，Job 狀態轉移至 matched
    Given 一個狀態為 "pending" 的 ProcessingJob
    And 已成功比對到一條規則
    When 呼叫 Job.MarkMatched(ruleId, processingContext)
    Then Job 的狀態應為 "matched"

  @S-18
  Scenario: 開始執行檔案處理時，Job 狀態轉移至 processing
    Given 一個狀態為 "matched" 的 ProcessingJob
    When 呼叫 Job.MarkProcessing()
    Then Job 的狀態應為 "processing"

  @S-19
  Scenario: 檔案處理成功後，Job 狀態轉移至 succeeded
    Given 一個狀態為 "processing" 的 ProcessingJob
    When 呼叫 Job.MarkSucceeded(newPath="~/Projects/my-app/assets/my-app-screenshot-20260416-001.png")
    Then Job 的狀態應為 "succeeded"
    And Job 的 ProcessingResult.newPath 應為 "~/Projects/my-app/assets/my-app-screenshot-20260416-001.png"

  @S-20
  Scenario: 檔案處理失敗後，Job 狀態轉移至 failed
    Given 一個狀態為 "processing" 的 ProcessingJob
    When 呼叫 Job.MarkFailed(errorMessage="permission denied")
    Then Job 的狀態應為 "failed"
    And Job 的 ProcessingResult.errorMessage 應為 "permission denied"

  @S-21
  Scenario: 非法狀態轉移應回傳錯誤
    Given 一個狀態為 "pending" 的 ProcessingJob
    When 呼叫 Job.MarkSucceeded(newPath="~/some/path")
    Then 應回傳錯誤，說明狀態轉移不合法
