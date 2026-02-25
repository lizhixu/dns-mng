export const zh = {
  // 通用
  common: {
    loading: '加载中...',
    error: '错误',
    cancel: '取消',
    save: '保存',
    delete: '删除',
    edit: '编辑',
    add: '添加',
    search: '搜索',
    refresh: '刷新',
    actions: '操作',
    confirm: '确认',
    active: '启用',
    inactive: '禁用',
    all: '全部',
    seconds: '秒',
  },

  // 登录页面
  login: {
    title: '登录',
    username: '用户名',
    password: '密码',
    loginBtn: '登录',
    noAccount: '还没有账号？',
    registerLink: '立即注册',
  },

  // 注册页面
  register: {
    title: '注册',
    username: '用户名',
    password: '密码',
    registerBtn: '注册',
    hasAccount: '已有账号？',
    loginLink: '立即登录',
  },

  // 布局
  layout: {
    title: 'DNS 管理器',
    accounts: '账户管理',
    domains: '域名列表',
    profile: '个人设置',
    logout: '退出登录',
    themeLight: '亮色模式',
    themeDark: '暗色模式',
    themeSystem: '跟随系统',
  },

  // 账户管理
  accounts: {
    title: '账户管理',
    subtitle: '管理您的 DNS 服务商账户',
    addAccount: '添加账户',
    accountName: '账户名称',
    accountNamePlaceholder: '例如：我的个人 DNS',
    provider: '服务商',
    selectProvider: '选择服务商',
    providerCannotChange: '服务商创建后不可更改',
    apiKey: 'API 密钥',
    apiKeyPlaceholder: '输入您的服务商 API 密钥',
    apiKeyKeepBlank: '留空则保持当前密钥不变',
    addedOn: '添加于',
    viewDomains: '查看域名',
    editAccount: '编辑账户',
    deleteAccount: '删除账户',
    deleteConfirm: '确定要删除此账户吗？',
    linkNewAccount: '关联新账户',
    linkAccount: '关联账户',
    saveChanges: '保存更改',
    noAccounts: '暂无关联账户',
    linkFirst: '关联您的第一个账户',
  },

  // 所有域名
  allDomains: {
    title: '所有域名',
    subtitle: '查看所有账户下的域名',
    searchPlaceholder: '搜索域名或账户...',
    noDomains: '暂无域名',
    noSearchResults: '未找到匹配的域名',
  },

  // 域名管理
  domains: {
    title: '域名管理',
    backToAccounts: '返回账户列表',
    searchPlaceholder: '搜索域名...',
    ttl: 'TTL',
    manageRecords: '管理记录',
    noDomains: '此账户暂无域名',
    noSearchResults: '未找到匹配的域名',
  },

  // DNS 记录管理
  records: {
    title: 'DNS 记录',
    backToDomains: '返回域名列表',
    addRecord: '添加记录',
    type: '类型',
    nodeName: '主机名',
    content: '记录值',
    ttl: 'TTL',
    state: '状态',
    domainName: '域名',
    searchPlaceholder: '搜索记录...',
    allTypes: '全部类型',
    noRecords: '暂无记录',
    deleteConfirm: '确定要删除此记录吗？',
    addDnsRecord: '添加 DNS 记录',
    editDnsRecord: '编辑 DNS 记录',
    nodeNamePlaceholder: '例如：www（留空则为主机记录）',
    ttlSeconds: 'TTL（秒）',
    priority: '优先级',
    // 记录类型说明
    contentLabels: {
      A: 'IPv4 地址',
      AAAA: 'IPv6 地址',
      CNAME: '目标主机（别名）',
      MX: '邮件服务器',
      TXT: '文本内容',
      SRV: '目标主机',
      SPF: 'SPF 内容',
      default: '记录值',
    },
    invalidNodeName: '主机名无效。只能使用字母、数字和连字符，不能以连字符开头或结尾。',
    backToAllDomains: '返回所有域名',
    rootRecordHint: '留空或输入 @ 表示根域名记录',
  },

  // 个人设置
  profile: {
    user_profile: '个人设置',
    account_info: '账户信息',
    username: '用户名',
    created_at: '注册时间',
    change_password: '修改密码',
    old_password: '当前密码',
    new_password: '新密码',
    confirm_password: '确认密码',
    update_password: '更新密码',
    updating: '更新中...',
    password_updated: '密码修改成功',
    passwords_not_match: '两次输入的密码不一致',
    password_min_length: '密码至少需要6位',
  },
};

export default zh;
