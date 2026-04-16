export interface RuleDTO {
  id: string
  name: string
  enabled: boolean
  watchFolder: string
  recursive: boolean
  filterExts: string[]
  filterKeyword: string
  nameTemplate: string
  targetTemplate: string
  project: string
  typeLabel: string
  createdAt: string
  updatedAt: string
}

export interface CreateRuleRequest {
  name: string
  watchFolder: string
  recursive: boolean
  filterExts: string[]
  filterKeyword: string
  nameTemplate: string
  targetTemplate: string
  project: string
  typeLabel: string
}

export interface LogDTO {
  logId: string
  ruleId: string
  ruleName: string
  originalPath: string
  newPath: string
  status: string
  errorMessage: string
  processedAt: string
}

export interface LicenseInfo {
  status: 'active' | 'expired' | 'activated'
  daysRemaining: number
}
