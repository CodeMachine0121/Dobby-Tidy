import {
  CreateRule,
  ListRules,
  GetRule,
  DeleteRule,
  EnableRule,
  DisableRule,
  ListRecentLogs,
  ListLogsByRule,
  GetTodayCount,
  GetLicenseInfo,
  ActivateLicense,
} from '../../wailsjs/go/main/App'
import type { CreateRuleRequest, LicenseInfo, LogDTO, RuleDTO } from '../types'

export const api = {
  rules: {
    list: (): Promise<RuleDTO[]> => ListRules(),
    get: (id: string): Promise<RuleDTO> => GetRule(id),
    create: (req: CreateRuleRequest): Promise<RuleDTO> => CreateRule(req),
    delete: (id: string): Promise<void> => DeleteRule(id),
    enable: (id: string): Promise<void> => EnableRule(id),
    disable: (id: string): Promise<void> => DisableRule(id),
  },
  logs: {
    recent: (limit: number): Promise<LogDTO[]> => ListRecentLogs(limit),
    byRule: (ruleId: string, limit: number): Promise<LogDTO[]> =>
      ListLogsByRule(ruleId, limit),
    todayCount: (): Promise<number> => GetTodayCount(),
  },
  license: {
    info: (): Promise<LicenseInfo> => GetLicenseInfo() as Promise<LicenseInfo>,
    activate: (key: string): Promise<void> => ActivateLicense(key),
  },
}
