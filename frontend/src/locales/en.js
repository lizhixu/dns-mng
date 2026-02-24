export const en = {
  // Common
  common: {
    loading: 'Loading...',
    error: 'Error',
    cancel: 'Cancel',
    save: 'Save',
    delete: 'Delete',
    edit: 'Edit',
    add: 'Add',
    search: 'Search',
    refresh: 'Refresh',
    actions: 'Actions',
    confirm: 'Confirm',
    active: 'Active',
    inactive: 'Inactive',
    all: 'All',
    seconds: 'seconds',
  },

  // Login page
  login: {
    title: 'Login',
    username: 'Username',
    password: 'Password',
    loginBtn: 'Login',
    noAccount: "Don't have an account?",
    registerLink: 'Register',
  },

  // Register page
  register: {
    title: 'Register',
    username: 'Username',
    password: 'Password',
    registerBtn: 'Register',
    hasAccount: 'Already have an account?',
    loginLink: 'Login',
  },

  // Layout
  layout: {
    title: 'DNS Manager',
    accounts: 'Accounts',
    domains: 'Domains',
    logout: 'Logout',
    themeLight: 'Light Mode',
    themeDark: 'Dark Mode',
    themeSystem: 'System Theme',
  },

  // Accounts
  accounts: {
    title: 'Accounts',
    subtitle: 'Manage your DNS provider accounts',
    addAccount: 'Add Account',
    accountName: 'Account Name',
    accountNamePlaceholder: 'e.g. My Personal DNS',
    provider: 'Provider',
    selectProvider: 'Select a provider',
    providerCannotChange: 'Provider cannot be changed',
    apiKey: 'API Key',
    apiKeyPlaceholder: 'Enter your provider API key',
    apiKeyKeepBlank: 'Leave blank to keep current key',
    addedOn: 'Added on',
    viewDomains: 'View Domains',
    editAccount: 'Edit Account',
    deleteAccount: 'Delete Account',
    deleteConfirm: 'Are you sure you want to delete this account?',
    linkNewAccount: 'Link New Account',
    linkAccount: 'Link Account',
    saveChanges: 'Save Changes',
    noAccounts: 'No accounts linked yet',
    linkFirst: 'Link your first account',
  },

  // All Domains
  allDomains: {
    title: 'All Domains',
    subtitle: 'View domains across all accounts',
    searchPlaceholder: 'Search domains or accounts...',
    noDomains: 'No domains found',
    noSearchResults: 'No domains match your search',
  },

  // Domains
  domains: {
    title: 'Domains',
    backToAccounts: 'Back to Accounts',
    searchPlaceholder: 'Search domains...',
    ttl: 'TTL',
    manageRecords: 'Manage Records',
    noDomains: 'No domains found for this account',
    noSearchResults: 'No domains match your search',
  },

  // DNS Records
  records: {
    title: 'DNS Records',
    backToDomains: 'Back to Domains',
    addRecord: 'Add Record',
    type: 'Type',
    nodeName: 'Node Name',
    content: 'Content',
    ttl: 'TTL',
    state: 'State',
    domainName: 'Domain Name',
    searchPlaceholder: 'Search records...',
    allTypes: 'All Types',
    noRecords: 'No records found',
    deleteConfirm: 'Are you sure you want to delete this record?',
    addDnsRecord: 'Add DNS Record',
    editDnsRecord: 'Edit DNS Record',
    nodeNamePlaceholder: 'e.g. www (or leave empty for root)',
    ttlSeconds: 'TTL (seconds)',
    priority: 'Priority',
    contentLabels: {
      A: 'IPv4 Address',
      AAAA: 'IPv6 Address',
      CNAME: 'Target Host (Alias)',
      MX: 'Mail Server Host',
      TXT: 'Text Content',
      SRV: 'Target Host',
      SPF: 'SPF Content',
      default: 'Content',
    },
    invalidNodeName: 'Invalid node name. Use only letters, numbers, and hyphens. Cannot start or end with a hyphen.',
    backToAllDomains: 'Back to All Domains',
    rootRecordHint: 'Leave empty or enter @ for root domain record',
  },
};

export default en;
