export interface TranslationKeys {
  nav: {
    dashboard: string
    rules: string
    logs: string
    settings: string
    trialExpired: string
    trialExpiredDesc: string
    enterLicenseKey: string
  }
  dashboard: {
    title: string
    subtitle: string
    todayProcessed: string
    filesUnit: string
    activeRules: string
    rulesUnit: string
    recentLogs: string
    viewAll: string
    noLogs: string
    noLogsHint: string
  }
  rules: {
    title: string
    subtitle: string
    addRule: string
    noRules: string
    noRulesHint: string
    add: string
    modal: {
      title: string
      name: string
      namePlaceholder: string
      watchFolder: string
      recursive: string
      filterExts: string
      filterExtsHint: string
      filterKeyword: string
      filterKeywordPlaceholder: string
      project: string
      typeLabel: string
      nameTemplate: string
      nameTemplatePreview: string
      targetFolder: string
      targetFolderHint: string
      cancel: string
      create: string
      saving: string
    }
    card: {
      enable: string
      disable: string
      delete: string
      confirmDelete: string
      cancelDelete: string
      nameTemplate: string
      targetFolder: string
      filterExts: string
      filterExtsAll: string
      filterKeyword: string
      project: string
      typeLabel: string
    }
    error: {
      nameRequired: string
      folderRequired: string
      templateRequired: string
    }
  }
  logs: {
    title: string
    subtitle: string
    subtitleWithError: string
    refresh: string
    filterByRule: string
    allRules: string
    noLogs: string
    noLogsForRule: string
    col: {
      original: string
      newPath: string
      rule: string
      time: string
    }
  }
  settings: {
    title: string
    subtitle: string
    language: string
    interfaceLanguage: string
    notifications: string
    desktopNotifications: string
    desktopNotificationsDesc: string
    about: string
    version: string
    appDesc: string
    templateRef: string
    templateCol: {
      var: string
      desc: string
      example: string
    }
    templateVars: {
      project: string
      type: string
      YYYY: string
      MM: string
      DD: string
      seq: string
      original: string
      ext: string
    }
  }
  license: {
    title: string
    activated: string
    activatedDesc: string
    trial: string
    trialDesc: string
    expired: string
    expiredDesc: string
    enterKey: string
    activate: string
    activating: string
    activateSuccess: string
    buyPrompt: string
    buyLink: string
    defaultError: string
  }
  common: {
    loading: string
  }
}
