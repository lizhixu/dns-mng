import { useState, useEffect, useRef } from 'react';
import { api } from '../api';
import { Clock, User, Activity, RefreshCw, Play } from 'lucide-react';
import { useLanguage } from '../LanguageContext';

const LogsManagement = () => {
    const { t, language } = useLanguage();
    const [operationLogs, setOperationLogs] = useState([]);
    const [schedulerLogs, setSchedulerLogs] = useState([]);
    const [opPagination, setOpPagination] = useState({ total: 0, page: 1, totalPages: 1 });
    const [schedPagination, setSchedPagination] = useState({ total: 0, page: 1, totalPages: 1 });
    const [opLoading, setOpLoading] = useState(false);
    const [schedLoading, setSchedLoading] = useState(false);
    const [opLoaded, setOpLoaded] = useState(false);
    const [schedLoaded, setSchedLoaded] = useState(false);
    const [error, setError] = useState('');
    const [activeTab, setActiveTab] = useState('operation'); // 'operation' or 'scheduler'
    const [triggering, setTriggering] = useState(false);
    const pageSize = 20;
    const topRef = useRef(null);
    const initialLoadRef = useRef(false);

    // Initial load for both logs to get counts
    useEffect(() => {
        if (!initialLoadRef.current) {
            initialLoadRef.current = true;
            loadOperationLogs();
            loadSchedulerLogs();
        }
    }, []);

    // Load data when tab becomes active is no longer needed since we load both initially

    // Load when page changes (skip initial page 1)
    useEffect(() => {
        if (opLoaded && opPagination.page !== 1) {
            loadOperationLogs();
        }
    }, [opPagination.page]);

    useEffect(() => {
        if (schedLoaded && schedPagination.page !== 1) {
            loadSchedulerLogs();
        }
    }, [schedPagination.page]);

    // Scroll to top when page changes
    useEffect(() => {
        if (topRef.current) {
            topRef.current.scrollIntoView({ behavior: 'smooth', block: 'start' });
        }
    }, [opPagination.page, schedPagination.page]);

    const loadOperationLogs = async () => {
        setOpLoading(true);
        try {
            const data = await api.getLogs(opPagination.page, pageSize);
            setOperationLogs(data.logs || []);
            setOpPagination(prev => ({
                ...prev,
                total: data.total || 0,
                totalPages: data.total_pages || 1
            }));
            setOpLoaded(true);
            setError('');
        } catch (err) {
            console.error('Failed to load operation logs:', err);
            setError(err.message);
        } finally {
            setOpLoading(false);
        }
    };

    const loadSchedulerLogs = async () => {
        setSchedLoading(true);
        try {
            const data = await api.getSchedulerLogs(schedPagination.page, pageSize);
            setSchedulerLogs(data.logs || []);
            setSchedPagination(prev => ({
                ...prev,
                total: data.total || 0,
                totalPages: data.total_pages || 1
            }));
            setSchedLoaded(true);
            setError('');
        } catch (err) {
            console.error('Failed to load scheduler logs:', err);
            setError(err.message);
        } finally {
            setSchedLoading(false);
        }
    };

    const handleTriggerCheck = async () => {
        setTriggering(true);
        try {
            await api.triggerSchedulerCheck();
            setTimeout(async () => {
                await loadSchedulerLogs();
                setTriggering(false);
            }, 2000);
        } catch (err) {
            setError(err.message);
            setTriggering(false);
        }
    };

    const handleRefresh = async () => {
        if (activeTab === 'operation') {
            if (opPagination.page === 1) {
                await loadOperationLogs();
            } else {
                setOpPagination(prev => ({ ...prev, page: 1 }));
            }
        } else {
            if (schedPagination.page === 1) {
                await loadSchedulerLogs();
            } else {
                setSchedPagination(prev => ({ ...prev, page: 1 }));
            }
        }
    };

    const getActionColor = (action) => {
        const colors = {
            create: '#10b981',
            update: '#3b82f6',
            delete: '#ef4444',
            soft_delete: '#f97316',
            restore: '#8b5cf6',
            login: '#06b6d4',
            logout: '#64748b',
            test: '#f59e0b',
            refresh: '#14b8a6',
            batch_update: '#3b82f6',
            batch_delete: '#ef4444'
        };
        return colors[action] || '#6b7280';
    };

    const renderLogDetails = (log) => {
        if (!log.details || typeof log.details !== 'object' || Object.keys(log.details).length === 0) {
            return null;
        }

        const details = log.details;
        const action = log.action;
        const resource = log.resource;

        // 特殊处理不同类型的日志
        const renderSpecialDetails = () => {
            // 登录日志
            if (action === 'login' && resource === 'user') {
                return (
                    <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', flexWrap: 'wrap' }}>
                        {details.username && (
                            <span>
                                <User size={14} style={{ display: 'inline', marginRight: '0.25rem', verticalAlign: 'middle' }} />
                                {details.username}
                            </span>
                        )}
                        {details.success !== undefined && (
                            <span style={{ 
                                color: details.success ? '#10b981' : '#ef4444',
                                fontWeight: '500'
                            }}>
                                {details.success ? '✓ 成功' : '✗ 失败'}
                            </span>
                        )}
                        {details.reason && (
                            <span style={{ color: '#ef4444' }}>
                                原因: {details.reason}
                            </span>
                        )}
                    </div>
                );
            }

            // 批量操作
            if (action.startsWith('batch_')) {
                return (
                    <div>
                        {details.count !== undefined && (
                            <span style={{ fontWeight: '500' }}>
                                📦 数量: {details.count}
                            </span>
                        )}
                        {details.domains && Array.isArray(details.domains) && (
                            <div style={{ marginTop: '0.5rem' }}>
                                {details.domains.map((domain, idx) => (
                                    <span key={idx} style={{ 
                                        display: 'inline-block',
                                        marginRight: '0.5rem',
                                        marginBottom: '0.25rem',
                                        padding: '0.125rem 0.5rem',
                                        backgroundColor: 'var(--bg-tertiary)',
                                        borderRadius: 'var(--radius-sm)',
                                        fontSize: '0.75rem'
                                    }}>
                                        {domain}
                                    </span>
                                ))}
                            </div>
                        )}
                    </div>
                );
            }

            // DNS 记录操作
            if (resource === 'record') {
                return (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.375rem' }}>
                        {details.type && (
                            <div>
                                <span style={{ color: 'var(--text-tertiary)', marginRight: '0.5rem' }}>类型:</span>
                                <span style={{ 
                                    fontWeight: '500',
                                    padding: '0.125rem 0.5rem',
                                    backgroundColor: 'var(--bg-tertiary)',
                                    borderRadius: 'var(--radius-sm)',
                                    fontSize: '0.75rem'
                                }}>{details.type}</span>
                            </div>
                        )}
                        {details.node_name !== undefined && (
                            <div>
                                <span style={{ color: 'var(--text-tertiary)', marginRight: '0.5rem' }}>主机名:</span>
                                <span style={{ fontFamily: 'monospace' }}>{details.node_name || '@'}</span>
                            </div>
                        )}
                        {details.name && (
                            <div>
                                <span style={{ color: 'var(--text-tertiary)', marginRight: '0.5rem' }}>记录名:</span>
                                <span style={{ fontFamily: 'monospace' }}>{details.name}</span>
                            </div>
                        )}
                        {details.value && (
                            <div>
                                <span style={{ color: 'var(--text-tertiary)', marginRight: '0.5rem' }}>记录值:</span>
                                <span style={{ fontFamily: 'monospace', wordBreak: 'break-all' }}>{details.value}</span>
                            </div>
                        )}
                        {details.content && (
                            <div>
                                <span style={{ color: 'var(--text-tertiary)', marginRight: '0.5rem' }}>内容:</span>
                                <span style={{ fontFamily: 'monospace', wordBreak: 'break-all' }}>{details.content}</span>
                            </div>
                        )}
                    </div>
                );
            }

            // 账户操作
            if (resource === 'account') {
                return (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.375rem' }}>
                        {details.name && (
                            <div>
                                <span style={{ color: 'var(--text-tertiary)', marginRight: '0.5rem' }}>账户名:</span>
                                <span style={{ fontWeight: '500' }}>{details.name}</span>
                            </div>
                        )}
                        {details.provider && (
                            <div>
                                <span style={{ color: 'var(--text-tertiary)', marginRight: '0.5rem' }}>服务商:</span>
                                <span style={{ 
                                    padding: '0.125rem 0.5rem',
                                    backgroundColor: 'var(--bg-tertiary)',
                                    borderRadius: 'var(--radius-sm)',
                                    fontSize: '0.75rem'
                                }}>{details.provider}</span>
                            </div>
                        )}
                    </div>
                );
            }

            // 域名缓存操作
            if (resource === 'domain_cache' || resource === 'domain') {
                // 只有在有其他信息时才显示详情区域
                const hasOtherInfo = details.renewal_date || details.renewal_url || details.count !== undefined;
                
                if (!hasOtherInfo) {
                    return null; // 域名已在标题显示，无其他信息则不显示详情
                }
                
                return (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.375rem' }}>
                        {details.renewal_date && (
                            <div>
                                <span style={{ color: 'var(--text-tertiary)', marginRight: '0.5rem' }}>续费日期:</span>
                                <span>{details.renewal_date}</span>
                            </div>
                        )}
                        {details.renewal_url && (
                            <div>
                                <span style={{ color: 'var(--text-tertiary)', marginRight: '0.5rem' }}>续费地址:</span>
                                <span style={{ wordBreak: 'break-all' }}>{details.renewal_url}</span>
                            </div>
                        )}
                        {details.count !== undefined && (
                            <div>
                                <span style={{ color: 'var(--text-tertiary)', marginRight: '0.5rem' }}>数量:</span>
                                <span style={{ fontWeight: '500' }}>{details.count}</span>
                            </div>
                        )}
                    </div>
                );
            }

            // 邮件配置
            if (resource === 'email_config') {
                return (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.375rem' }}>
                        {details.smtp_host && (
                            <div>
                                <span style={{ color: 'var(--text-tertiary)', marginRight: '0.5rem' }}>SMTP服务器:</span>
                                <span style={{ fontFamily: 'monospace' }}>{details.smtp_host}:{details.smtp_port || 587}</span>
                            </div>
                        )}
                        {details.enabled !== undefined && (
                            <div>
                                <span style={{ color: 'var(--text-tertiary)', marginRight: '0.5rem' }}>状态:</span>
                                <span style={{ color: details.enabled ? '#10b981' : '#ef4444' }}>
                                    {details.enabled ? '✓ 已启用' : '✗ 已禁用'}
                                </span>
                            </div>
                        )}
                    </div>
                );
            }

            // 通知设置
            if (resource === 'notification_setting') {
                return (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.375rem' }}>
                        {details.days_before !== undefined && (
                            <div>
                                <span style={{ color: 'var(--text-tertiary)', marginRight: '0.5rem' }}>提前天数:</span>
                                <span style={{ fontWeight: '500' }}>{details.days_before} 天</span>
                            </div>
                        )}
                        {details.enabled !== undefined && (
                            <div>
                                <span style={{ color: 'var(--text-tertiary)', marginRight: '0.5rem' }}>状态:</span>
                                <span style={{ color: details.enabled ? '#10b981' : '#ef4444' }}>
                                    {details.enabled ? '✓ 已启用' : '✗ 已禁用'}
                                </span>
                            </div>
                        )}
                    </div>
                );
            }

            return null;
        };

        const specialContent = renderSpecialDetails();
        if (specialContent) {
            return (
                <div style={{ 
                    fontSize: '0.8125rem', 
                    color: 'var(--text-secondary)',
                    marginTop: '0.5rem',
                    padding: '0.75rem',
                    backgroundColor: 'var(--bg-secondary)',
                    borderRadius: 'var(--radius-sm)',
                    borderLeft: '3px solid ' + getActionColor(action)
                }}>
                    {specialContent}
                </div>
            );
        }

        // 默认显示所有字段
        return (
            <div style={{ 
                fontSize: '0.8125rem', 
                color: 'var(--text-secondary)',
                marginTop: '0.5rem',
                padding: '0.75rem',
                backgroundColor: 'var(--bg-secondary)',
                borderRadius: 'var(--radius-sm)',
                borderLeft: '3px solid var(--border-color)'
            }}>
                {Object.entries(details).map(([key, value]) => (
                    <div key={key} style={{ 
                        display: 'flex', 
                        gap: '0.5rem',
                        marginBottom: '0.375rem',
                        alignItems: 'flex-start'
                    }}>
                        <span style={{ 
                            color: 'var(--text-tertiary)', 
                            fontWeight: '500',
                            minWidth: '100px',
                            flexShrink: 0
                        }}>
                            {key}:
                        </span>
                        <span style={{ 
                            color: 'var(--text-primary)',
                            wordBreak: 'break-word',
                            fontFamily: typeof value === 'string' && value.length > 50 ? 'monospace' : 'inherit',
                            fontSize: typeof value === 'string' && value.length > 50 ? '0.75rem' : 'inherit'
                        }}>
                            {typeof value === 'object' ? JSON.stringify(value, null, 2) : String(value)}
                        </span>
                    </div>
                ))}
            </div>
        );
    };

    const getStatusColor = (status) => {
        const colors = {
            success: '#10b981',
            partial_success: '#f59e0b',
            error: '#ef4444',
            running: '#3b82f6'
        };
        return colors[status] || '#6b7280';
    };

    const getStatusIcon = (status) => {
        const icons = {
            success: '✓',
            partial_success: '⚠',
            error: '✗',
            running: '⟳'
        };
        return icons[status] || '•';
    };

    const formatDate = (dateString) => {
        const locale = language === 'en' ? 'en-US' : 'zh-CN';
        return new Date(dateString).toLocaleString(locale, {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit'
        });
    };

    // Pagination controls
    const handleOpPageChange = (newPage) => {
        if (newPage >= 1 && newPage <= opPagination.totalPages) {
            setOpPagination(prev => ({ ...prev, page: newPage }));
        }
    };

    const handleSchedPageChange = (newPage) => {
        if (newPage >= 1 && newPage <= schedPagination.totalPages) {
            setSchedPagination(prev => ({ ...prev, page: newPage }));
        }
    };

    return (
        <div>
            <div ref={topRef} style={{ marginBottom: '2rem' }}>
                <div className="page-title-row" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
                    <div style={{ minWidth: 0, flex: 1 }}>
                        <h2 style={{ fontSize: '1.5rem', fontWeight: 'bold' }}>{t.logsManagement.title}</h2>
                        <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', marginTop: '0.25rem' }}>
                            {t.logsManagement.subtitle}
                        </p>
                    </div>
                    <div className="page-actions-bar logs-toolbar" style={{ display: 'flex', gap: '0.75rem' }}>
                        <button
                            onClick={handleRefresh}
                            disabled={opLoading || schedLoading}
                            className="btn btn-secondary"
                            title={t.logsManagement.refresh}
                        >
                            <RefreshCw size={18} className={opLoading || schedLoading ? "spin" : ""} style={{ animation: opLoading || schedLoading ? 'spin 1s linear infinite' : 'none' }} />
                        </button>
                        {activeTab === 'scheduler' && (
                            <button
                                onClick={handleTriggerCheck}
                                disabled={triggering}
                                className="btn btn-primary"
                                title={t.logsManagement.triggerCheck}
                            >
                                <Play size={18} />
                            </button>
                        )}
                    </div>
                </div>
            </div>

            {/* Tab Navigation */}
            <div className="tab-nav" style={{
                display: 'flex',
                gap: '0.5rem',
                marginBottom: '1.5rem',
                borderBottom: '1px solid var(--border-color)'
            }}>
                <button
                    className={`tab-nav-btn${activeTab === 'operation' ? ' active' : ''}`}
                    onClick={() => setActiveTab('operation')}
                    style={{
                        padding: '0.75rem 1.5rem',
                        background: 'none',
                        border: 'none',
                        borderBottom: activeTab === 'operation' ? '2px solid var(--accent-primary)' : '2px solid transparent',
                        color: activeTab === 'operation' ? 'var(--accent-primary)' : 'var(--text-secondary)',
                        fontWeight: activeTab === 'operation' ? '600' : '400',
                        cursor: 'pointer',
                        transition: 'all 0.2s',
                        display: 'flex',
                        alignItems: 'center',
                        gap: '0.5rem'
                    }}
                >
                    <User size={18} />
                    {t.logsManagement.operationTab} ({opPagination.total})
                </button>
                <button
                    className={`tab-nav-btn${activeTab === 'scheduler' ? ' active' : ''}`}
                    onClick={() => setActiveTab('scheduler')}
                    style={{
                        padding: '0.75rem 1.5rem',
                        background: 'none',
                        border: 'none',
                        borderBottom: activeTab === 'scheduler' ? '2px solid var(--accent-primary)' : '2px solid transparent',
                        color: activeTab === 'scheduler' ? 'var(--accent-primary)' : 'var(--text-secondary)',
                        fontWeight: activeTab === 'scheduler' ? '600' : '400',
                        cursor: 'pointer',
                        transition: 'all 0.2s',
                        display: 'flex',
                        alignItems: 'center',
                        gap: '0.5rem'
                    }}
                >
                    <Clock size={18} />
                    {t.logsManagement.schedulerTab} ({schedPagination.total})
                </button>
            </div>

            {error && (
                <div style={{ 
                    color: 'var(--danger)', 
                    marginBottom: '1rem', 
                    padding: '1rem', 
                    backgroundColor: 'rgba(239, 68, 68, 0.1)', 
                    borderRadius: 'var(--radius-md)' 
                }}>
                    {error}
                </div>
            )}

            {(activeTab === 'operation' ? opLoading : schedLoading) ? (
                <div style={{ textAlign: 'center', padding: '4rem' }}>
                    <div className="spinner" style={{ margin: '0 auto' }}></div>
                </div>
            ) : (
                <>
                    {/* Operation Logs */}
                    {activeTab === 'operation' && (
                        <>
                            <div className="logs-list" style={{ display: 'grid', gap: '0.75rem' }}>
                                {operationLogs.map(log => {
                                    const hasDetails = log.details && typeof log.details === 'object' && Object.keys(log.details).length > 0;
                                    const domainName = log.details?.domain || log.details?.name || null;
                                    const showDomainInTitle = domainName && (log.resource === 'domain' || log.resource === 'domain_cache' || log.resource === 'record');
                                    
                                    return (
                                    <div key={log.id} className="glass-panel log-card" style={{ padding: '1rem' }}>
                                        <div className="log-card-layout" style={{ display: 'flex', alignItems: hasDetails ? 'flex-start' : 'center', gap: '1rem' }}>
                                            <Activity size={20} style={{ color: getActionColor(log.action), flexShrink: 0, marginTop: hasDetails ? '0.125rem' : '0' }} />
                                            <div style={{ flex: 1, minWidth: 0 }}>
                                                <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', flexWrap: 'wrap', marginBottom: hasDetails ? '0.5rem' : '0' }}>
                                                    <span style={{ 
                                                        fontSize: '0.875rem', 
                                                        fontWeight: '600',
                                                        color: getActionColor(log.action),
                                                        padding: '0.25rem 0.625rem',
                                                        backgroundColor: `${getActionColor(log.action)}15`,
                                                        borderRadius: 'var(--radius-sm)'
                                                    }}>
                                                        {t.logs?.actions?.[log.action] || log.action}
                                                    </span>
                                                    <span style={{ fontSize: '0.9375rem', color: 'var(--text-primary)', fontWeight: '500' }}>
                                                        {t.logs?.resources?.[log.resource] || log.resource}
                                                    </span>
                                                    {showDomainInTitle && (
                                                        <span style={{ 
                                                            fontSize: '0.875rem', 
                                                            color: 'var(--text-secondary)',
                                                            fontFamily: 'monospace',
                                                            fontWeight: '500'
                                                        }}>
                                                            {domainName}
                                                        </span>
                                                    )}
                                                    {log.resource_id && !showDomainInTitle && (
                                                        <span style={{ 
                                                            fontSize: '0.75rem', 
                                                            color: 'var(--text-tertiary)',
                                                            fontFamily: 'monospace',
                                                            backgroundColor: 'var(--bg-secondary)',
                                                            padding: '0.125rem 0.5rem',
                                                            borderRadius: 'var(--radius-sm)'
                                                        }}>
                                                            #{log.resource_id}
                                                        </span>
                                                    )}
                                                </div>
                                                {renderLogDetails(log)}
                                            </div>
                                            <div style={{ textAlign: 'right', flexShrink: 0, display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: '0.375rem' }}>
                                                <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)', whiteSpace: 'nowrap' }}>
                                                    <Clock size={12} style={{ display: 'inline', marginRight: '0.25rem', verticalAlign: 'middle' }} />
                                                    {formatDate(log.created_at)}
                                                </div>
                                                {log.ip_address && (
                                                    <div style={{ 
                                                        fontSize: '0.6875rem', 
                                                        color: 'var(--text-tertiary)',
                                                        fontFamily: 'monospace',
                                                        backgroundColor: 'var(--bg-secondary)',
                                                        padding: '0.125rem 0.375rem',
                                                        borderRadius: 'var(--radius-sm)'
                                                    }}>
                                                        {log.ip_address}
                                                    </div>
                                                )}
                                            </div>
                                        </div>
                                    </div>
                                    );
                                })}
                                {operationLogs.length === 0 && (
                                    <div style={{ textAlign: 'center', padding: '4rem', color: 'var(--text-secondary)' }}>
                                        {t.logsManagement.noOperationLogs}
                                    </div>
                                )}
                            </div>
                            {opPagination.totalPages > 1 && (
                                <div className="pagination-controls" style={{
                                    display: 'flex',
                                    justifyContent: 'center',
                                    alignItems: 'center',
                                    gap: '0.75rem',
                                    marginTop: '2rem',
                                    padding: '1rem',
                                    borderTop: '1px solid var(--border-color)'
                                }}>
                                    <button
                                        onClick={() => handleOpPageChange(opPagination.page - 1)}
                                        disabled={opPagination.page === 1}
                                        className="btn btn-secondary"
                                        style={{ 
                                            minWidth: '80px',
                                            opacity: opPagination.page === 1 ? 0.5 : 1,
                                            cursor: opPagination.page === 1 ? 'not-allowed' : 'pointer'
                                        }}
                                    >
                                        {t.common.previousPage}
                                    </button>
                                    <div style={{ 
                                        display: 'flex', 
                                        alignItems: 'center', 
                                        gap: '0.5rem',
                                        padding: '0.5rem 1rem',
                                        backgroundColor: 'var(--bg-secondary)',
                                        borderRadius: 'var(--radius-sm)',
                                        fontSize: '0.875rem'
                                    }}>
                                        <span style={{ color: 'var(--text-secondary)' }}>{t.logsManagement.page}</span>
                                        <span style={{ 
                                            color: 'var(--accent-primary)', 
                                            fontWeight: '600',
                                            minWidth: '2ch',
                                            textAlign: 'center'
                                        }}>
                                            {opPagination.page}
                                        </span>
                                        <span style={{ color: 'var(--text-secondary)' }}>/</span>
                                        <span style={{ color: 'var(--text-primary)', fontWeight: '500' }}>
                                            {opPagination.totalPages}
                                        </span>
                                        <span style={{ color: 'var(--text-secondary)' }}>{t.logsManagement.pages}</span>
                                        <span style={{ 
                                            marginLeft: '0.5rem', 
                                            paddingLeft: '0.5rem', 
                                            borderLeft: '1px solid var(--border-color)',
                                            color: 'var(--text-tertiary)',
                                            fontSize: '0.8125rem'
                                        }}>
                                            {t.common.totalItems.replace('{count}', opPagination.total)}
                                        </span>
                                    </div>
                                    <button
                                        onClick={() => handleOpPageChange(opPagination.page + 1)}
                                        disabled={opPagination.page === opPagination.totalPages}
                                        className="btn btn-secondary"
                                        style={{ 
                                            minWidth: '80px',
                                            opacity: opPagination.page === opPagination.totalPages ? 0.5 : 1,
                                            cursor: opPagination.page === opPagination.totalPages ? 'not-allowed' : 'pointer'
                                        }}
                                    >
                                        {t.common.nextPage}
                                    </button>
                                </div>
                            )}
                        </>
                    )}

                    {/* Scheduler Logs */}
                    {activeTab === 'scheduler' && (
                        <>
                            <div className="logs-list" style={{ display: 'grid', gap: '0.75rem' }}>
                                {schedulerLogs.map(log => {
                                    let details = {};
                                    try {
                                        details = log.details ? JSON.parse(log.details) : {};
                                    } catch (e) {
                                        details = {};
                                    }

                                    return (
                                        <div key={log.id} className="glass-panel log-card" style={{ padding: '1rem' }}>
                                            <div className="log-card-layout" style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                                                <span style={{ 
                                                    fontSize: '1.25rem',
                                                    color: getStatusColor(log.status),
                                                    flexShrink: 0
                                                }}>
                                                    {getStatusIcon(log.status)}
                                                </span>
                                                <div style={{ flex: 1, minWidth: 0 }}>
                                                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', flexWrap: 'wrap', marginBottom: '0.25rem' }}>
                                                        <span style={{ 
                                                            fontSize: '0.875rem', 
                                                            fontWeight: '600',
                                                            color: 'var(--text-primary)'
                                                        }}>
                                                            {t.logsManagement.taskNames?.[log.task_name] || log.task_name}
                                                        </span>
                                                        <span style={{ 
                                                            fontSize: '0.75rem', 
                                                            fontWeight: '600',
                                                            color: getStatusColor(log.status),
                                                            textTransform: 'uppercase',
                                                            letterSpacing: '0.05em',
                                                            padding: '0.125rem 0.5rem',
                                                            backgroundColor: `${getStatusColor(log.status)}20`,
                                                            borderRadius: 'var(--radius-sm)'
                                                        }}>
                                                            {log.status}
                                                        </span>
                                                    </div>
                                                    
                                                    {log.message && (
                                                        <div style={{ 
                                                            fontSize: '0.875rem', 
                                                            color: 'var(--text-secondary)',
                                                            marginBottom: '0.25rem'
                                                        }}>
                                                            {log.message}
                                                        </div>
                                                    )}

                                                    {details && Object.keys(details).length > 0 && (
                                                        <div style={{ 
                                                            fontSize: '0.75rem', 
                                                            color: 'var(--text-tertiary)',
                                                            marginTop: '0.5rem'
                                                        }}>
                                                            {details.total_domains !== undefined && (
                                                                <div style={{ marginBottom: '0.25rem' }}>
                                                                    📊 {t.logsManagement.schedulerDetails.totalDomains}: {details.total_domains} {t.logsManagement.schedulerDetails.domainsUnit}
                                                                    {details.success_count !== undefined && (
                                                                        <span style={{ marginLeft: '1rem', color: '#10b981' }}>
                                                                            ✓ {t.logsManagement.schedulerDetails.success}: {details.success_count}
                                                                        </span>
                                                                    )}
                                                                    {details.error_count !== undefined && details.error_count > 0 && (
                                                                        <span style={{ marginLeft: '1rem', color: '#ef4444' }}>
                                                                            ✗ {t.logsManagement.schedulerDetails.failed}: {details.error_count}
                                                                        </span>
                                                                    )}
                                                                </div>
                                                            )}
                                                            {details.user_domains && Object.keys(details.user_domains).length > 0 && (
                                                                <div style={{ 
                                                                    marginTop: '0.5rem',
                                                                    padding: '0.5rem',
                                                                    backgroundColor: 'var(--bg-secondary)',
                                                                    borderRadius: 'var(--radius-sm)'
                                                                }}>
                                                                    {Object.entries(details.user_domains).map(([userId, domains]) => (
                                                                        <div key={userId} style={{ marginBottom: '0.25rem' }}>
                                                                            👤 {t.logsManagement.schedulerDetails.user} {userId}: {domains.join(', ')}
                                                                        </div>
                                                                    ))}
                                                                </div>
                                                            )}
                                                            {details.errors && details.errors.length > 0 && (
                                                                <div style={{ 
                                                                    marginTop: '0.5rem',
                                                                    padding: '0.5rem',
                                                                    backgroundColor: 'rgba(239, 68, 68, 0.1)',
                                                                    borderRadius: 'var(--radius-sm)',
                                                                    color: '#ef4444'
                                                                }}>
                                                                    {details.errors.map((err, idx) => (
                                                                        <div key={idx} style={{ marginBottom: '0.25rem' }}>
                                                                            ⚠ {err.domain}: {err.error}
                                                                        </div>
                                                                    ))}
                                                                </div>
                                                            )}
                                                        </div>
                                                    )}
                                                </div>
                                                <div style={{ textAlign: 'right', flexShrink: 0, display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: '0.25rem' }}>
                                                    <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)' }}>
                                                        {formatDate(log.started_at)}
                                                    </div>
                                                    {log.duration_ms && (
                                                        <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)' }}>
                                                            ⏱ {log.duration_ms}ms
                                                        </div>
                                                    )}
                                                </div>
                                            </div>
                                        </div>
                                    );
                                })}
                                {schedulerLogs.length === 0 && (
                                    <div style={{ textAlign: 'center', padding: '4rem', color: 'var(--text-secondary)' }}>
                                        {t.logsManagement.noSchedulerLogs}
                                    </div>
                                )}
                            </div>
                            {schedPagination.totalPages > 1 && (
                                <div className="pagination-controls" style={{
                                    display: 'flex',
                                    justifyContent: 'center',
                                    alignItems: 'center',
                                    gap: '0.75rem',
                                    marginTop: '2rem',
                                    padding: '1rem',
                                    borderTop: '1px solid var(--border-color)'
                                }}>
                                    <button
                                        onClick={() => handleSchedPageChange(schedPagination.page - 1)}
                                        disabled={schedPagination.page === 1}
                                        className="btn btn-secondary"
                                        style={{ 
                                            minWidth: '80px',
                                            opacity: schedPagination.page === 1 ? 0.5 : 1,
                                            cursor: schedPagination.page === 1 ? 'not-allowed' : 'pointer'
                                        }}
                                    >
                                        {t.common.previousPage}
                                    </button>
                                    <div style={{ 
                                        display: 'flex', 
                                        alignItems: 'center', 
                                        gap: '0.5rem',
                                        padding: '0.5rem 1rem',
                                        backgroundColor: 'var(--bg-secondary)',
                                        borderRadius: 'var(--radius-sm)',
                                        fontSize: '0.875rem'
                                    }}>
                                        <span style={{ color: 'var(--text-secondary)' }}>{t.logsManagement.page}</span>
                                        <span style={{ 
                                            color: 'var(--accent-primary)', 
                                            fontWeight: '600',
                                            minWidth: '2ch',
                                            textAlign: 'center'
                                        }}>
                                            {schedPagination.page}
                                        </span>
                                        <span style={{ color: 'var(--text-secondary)' }}>/</span>
                                        <span style={{ color: 'var(--text-primary)', fontWeight: '500' }}>
                                            {schedPagination.totalPages}
                                        </span>
                                        <span style={{ color: 'var(--text-secondary)' }}>{t.logsManagement.pages}</span>
                                        <span style={{ 
                                            marginLeft: '0.5rem', 
                                            paddingLeft: '0.5rem', 
                                            borderLeft: '1px solid var(--border-color)',
                                            color: 'var(--text-tertiary)',
                                            fontSize: '0.8125rem'
                                        }}>
                                            {t.common.totalItems.replace('{count}', schedPagination.total)}
                                        </span>
                                    </div>
                                    <button
                                        onClick={() => handleSchedPageChange(schedPagination.page + 1)}
                                        disabled={schedPagination.page === schedPagination.totalPages}
                                        className="btn btn-secondary"
                                        style={{ 
                                            minWidth: '80px',
                                            opacity: schedPagination.page === schedPagination.totalPages ? 0.5 : 1,
                                            cursor: schedPagination.page === schedPagination.totalPages ? 'not-allowed' : 'pointer'
                                        }}
                                    >
                                        {t.common.nextPage}
                                    </button>
                                </div>
                            )}
                        </>
                    )}
                </>
            )}
        </div>
    );
};

export default LogsManagement;
