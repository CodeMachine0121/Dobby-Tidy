Feature: Background File Processor
  As a user with file processing rules configured
  I want the application to automatically scan watched folders and process matching files
  So that files are renamed and moved without manual intervention

  Background:
    Given a rule "發票規則" watching folder "/Downloads" with extension filter [".pdf"] and keyword ""
    And the rule has name template "{project}_{type}_{YYYY}{MM}{DD}_{seq}{ext}"
    And the rule has target template "/Documents/{project}/{type}/{YYYY}/{MM}"
    And the rule is enabled

  @S-1
  Scenario: Scan discovers a matching file and processes it successfully
    Given the folder "/Downloads" contains a file "invoice_abc.pdf"
    When the background processor runs a scan
    Then a ProcessingJob is created in state "succeeded"
    And the file is moved to the rendered target path
    And an operation log entry is written with status "success"

  @S-2
  Scenario: Scan skips files that do not match the filter
    Given the folder "/Downloads" contains a file "photo.jpg"
    When the background processor runs a scan
    Then no ProcessingJob is created for "photo.jpg"
    And no operation log entry is written for "photo.jpg"

  @S-3
  Scenario: Scan skips files that do not match the keyword filter
    Given a rule "合約規則" watching folder "/Downloads" with extension filter [".pdf"] and keyword "contract"
    And the rule is enabled
    And the folder "/Downloads" contains a file "invoice.pdf"
    When the background processor runs a scan
    Then no ProcessingJob is created for "invoice.pdf" under "合約規則"

  @S-4
  Scenario: File move failure marks job as failed and writes error log
    Given the folder "/Downloads" contains a file "invoice_abc.pdf"
    And the file system will fail to move the file with error "permission denied"
    When the background processor runs a scan
    Then a ProcessingJob is created in state "failed"
    And an operation log entry is written with status "failed"
    And the error message contains "permission denied"

  @S-5
  Scenario: No enabled rules results in no processing
    Given all rules are disabled
    When the background processor runs a scan
    Then no ProcessingJob is created
    And no operation log entry is written

  @S-6
  Scenario: Scan is skipped when another scan is already running
    Given a scan is currently in progress
    When the background processor attempts to start another scan
    Then the second scan is skipped without error

  @S-7
  Scenario: Scan covers recursive subdirectories when rule is recursive
    Given the rule "發票規則" has recursive scanning enabled
    And the folder "/Downloads/2024" contains a file "invoice_sub.pdf"
    When the background processor runs a scan
    Then a ProcessingJob is created for "invoice_sub.pdf"

  @S-8
  Scenario: Scan ignores files in subdirectories when rule is non-recursive
    Given the rule "發票規則" has recursive scanning disabled
    And only the folder "/Downloads/2024" contains a file "invoice_sub.pdf"
    When the background processor runs a scan
    Then no ProcessingJob is created for "invoice_sub.pdf"

  @S-9
  Scenario: Processor builds correct ProcessingContext before rendering templates
    Given the folder "/Downloads" contains a file "invoice_abc.pdf"
    And today's date is 2024-03-15
    And the sequence for the rule today is 1
    When the background processor runs a scan
    Then the ProcessingJob context has project from the rule
    And the ProcessingJob context has date 2024-03-15
    And the ProcessingJob context has seq "001"
