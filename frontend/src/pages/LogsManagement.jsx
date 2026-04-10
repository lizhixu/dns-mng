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
            login: '#8b5cf6',
            test: '#f59e0b'
        };
        return colors[action] || '#6b7280';
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
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
                    <div>
                        <h2 style={{ fontSize: '1.5rem', fontWeight: 'bold' }}>{t.logsManagement.title}</h2>
                        <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', marginTop: '0.25rem' }}>
                            {t.logsManagement.subtitle}
                        </p>
                    </div>
                    <div style={{ display: 'flex', gap: '0.75rem' }}>
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
            <div style={{ 
                display: 'flex', 
                gap: '0.5rem', 
                marginBottom: '1.5rem',
                borderBottom: '1px solid var(--border-color)'
            }}>
                <button
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
                            <div style={{ display: 'grid', gap: '0.75rem' }}>
                                {operationLogs.map(log => (
                                    <div key={log.id} className="glass-panel" style={{ padding: '1rem' }}>
                                        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                                            <Activity size={18} style={{ color: getActionColor(log.action), flexShrink: 0 }} />
                                            <div style={{ flex: 1, minWidth: 0 }}>
                                                <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', flexWrap: 'wrap' }}>
                                                    <span style={{ 
                                                        fontSize: '0.75rem', 
                                                        fontWeight: '600',
                                                        color: getActionColor(log.action),
                                                        textTransform: 'uppercase',
                                                        letterSpacing: '0.05em'
                                                    }}>
                                                        {log.action}
                                                    </span>
                                                    <span style={{ fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
                                                        {log.resource}
                                                    </span>
                                                    {log.resource_id && (
                                                        <span style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)' }}>
                                                            #{log.resource_id}
                                                        </span>
                                                    )}
                                                </div>
                                                {log.details && typeof log.details === 'object' && Object.keys(log.details).length > 0 && (
                                                    <div style={{ 
                                                        fontSize: '0.75rem', 
                                                        color: 'var(--text-tertiary)',
                                                        marginTop: '0.5rem',
                                                        padding: '0.5rem',
                                                        backgroundColor: 'var(--bg-secondary)',
                                                        borderRadius: 'var(--radius-sm)',
                                                        fontFamily: 'monospace'
                                                    }}>
                                                        {JSON.stringify(log.details, null, 2)}
                                                    </div>
                                                )}
                                            </div>
                                            <div style={{ textAlign: 'right', flexShrink: 0, display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: '0.25rem' }}>
                                                <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)' }}>
                                                    {formatDate(log.created_at)}
                                                </div>
                                                {log.ip_address && (
                                                    <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)' }}>
                                                        {log.ip_address}
                                                    </div>
                                                )}
                                            </div>
                                        </div>
                                    </div>
                                ))}
                                {operationLogs.length === 0 && (
                                    <div style={{ textAlign: 'center', padding: '4rem', color: 'var(--text-secondary)' }}>
                                        {t.logsManagement.noOperationLogs}
                                    </div>
                                )}
                            </div>
                            {opPagination.totalPages > 1 && (
                                <div style={{ 
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
                            <div style={{ display: 'grid', gap: '0.75rem' }}>
                                {schedulerLogs.map(log => {
                                    let details = {};
                                    try {
                                        details = log.details ? JSON.parse(log.details) : {};
                                    } catch (e) {
                                        details = {};
                                    }

                                    return (
                                        <div key={log.id} className="glass-panel" style={{ padding: '1rem' }}>
                                            <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
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
                                <div style={{ 
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
