import { useState, useEffect, useRef } from 'react';
import { api } from '../api';
import { Clock, User, Activity, RefreshCw, ChevronDown, ChevronUp, Globe, Play, LogIn, MapPin } from 'lucide-react';
import { useLanguage } from '../LanguageContext';

let initialAPILogsRequest = null;
let initialSchedulerLogsRequest = null;
let initialLoginLogsRequest = null;
const pageSize = 20;

const clearInitialRequestSoon = (clearRequest) => {
    setTimeout(clearRequest, 1000);
};

const getAPILogsData = (page, useInitialRequest) => {
    if (useInitialRequest && page === 1) {
        initialAPILogsRequest ||= api.getAPICallLogs(page, 20).catch(error => {
            initialAPILogsRequest = null;
            throw error;
        }).finally(() => {
            clearInitialRequestSoon(() => {
                initialAPILogsRequest = null;
            });
        });
        return initialAPILogsRequest;
    }

    return api.getAPICallLogs(page, 20);
};

const getSchedulerLogsData = (page, useInitialRequest) => {
    if (useInitialRequest && page === 1) {
        initialSchedulerLogsRequest ||= api.getSchedulerLogs(page, 20).catch(error => {
            initialSchedulerLogsRequest = null;
            throw error;
        }).finally(() => {
            clearInitialRequestSoon(() => {
                initialSchedulerLogsRequest = null;
            });
        });
        return initialSchedulerLogsRequest;
    }

    return api.getSchedulerLogs(page, 20);
};

const getLoginLogsData = (page, useInitialRequest) => {
    if (useInitialRequest && page === 1) {
        initialLoginLogsRequest ||= api.getLoginLogs(page, 20).catch(error => {
            initialLoginLogsRequest = null;
            throw error;
        }).finally(() => {
            clearInitialRequestSoon(() => {
                initialLoginLogsRequest = null;
            });
        });
        return initialLoginLogsRequest;
    }

    return api.getLoginLogs(page, 20);
};

const APILogsManagement = () => {
    const { t, language } = useLanguage();
    const [apiLogs, setApiLogs] = useState([]);
    const [schedulerLogs, setSchedulerLogs] = useState([]);
    const [loginLogs, setLoginLogs] = useState([]);
    const [apiPagination, setApiPagination] = useState({ total: 0, page: 1, totalPages: 1 });
    const [schedPagination, setSchedPagination] = useState({ total: 0, page: 1, totalPages: 1 });
    const [loginPagination, setLoginPagination] = useState({ total: 0, page: 1, totalPages: 1 });
    const [apiLoading, setApiLoading] = useState(false);
    const [schedLoading, setSchedLoading] = useState(false);
    const [loginLoading, setLoginLoading] = useState(false);
    const [error, setError] = useState('');
    const [activeTab, setActiveTab] = useState('api'); // 'api', 'scheduler', or 'login'
    const [expandedLogs, setExpandedLogs] = useState(new Set());
    const [triggering, setTriggering] = useState(false);
    const topRef = useRef(null);
    const initialLoadRef = useRef(false);
    const apiPageEffectReadyRef = useRef(false);
    const schedPageEffectReadyRef = useRef(false);
    const loginPageEffectReadyRef = useRef(false);

    const pageTitle = activeTab === 'api'
        ? t.logsManagement.title
        : activeTab === 'scheduler'
        ? t.logsManagement.schedulerTitle
        : t.logsManagement.loginTitle;
    const pageSubtitle = activeTab === 'api'
        ? t.logsManagement.subtitle
        : activeTab === 'scheduler'
        ? t.logsManagement.schedulerSubtitle
        : t.logsManagement.loginSubtitle;

    useEffect(() => {
        if (!initialLoadRef.current) {
            initialLoadRef.current = true;
            loadAPILogs(1, { useInitialRequest: true });
            loadSchedulerLogs(1, { useInitialRequest: true });
            loadLoginLogs(1, { useInitialRequest: true });
        }
    }, []);

    useEffect(() => {
        if (!apiPageEffectReadyRef.current) {
            return;
        }
        loadAPILogs(apiPagination.page);
    }, [apiPagination.page]);

    useEffect(() => {
        if (!schedPageEffectReadyRef.current) {
            return;
        }
        loadSchedulerLogs(schedPagination.page);
    }, [schedPagination.page]);

    useEffect(() => {
        if (!loginPageEffectReadyRef.current) {
            return;
        }
        loadLoginLogs(loginPagination.page);
    }, [loginPagination.page]);

    useEffect(() => {
        if (topRef.current) {
            topRef.current.scrollIntoView({ behavior: 'smooth', block: 'start' });
        }
    }, [apiPagination.page, schedPagination.page, loginPagination.page]);

    const loadAPILogs = async (page = apiPagination.page, options = {}) => {
        setApiLoading(true);
        try {
            const data = await getAPILogsData(page, options.useInitialRequest);
            setApiLogs(data.logs || []);
            setApiPagination(prev => ({
                ...prev,
                page: data.page || page,
                total: data.total || 0,
                totalPages: data.total_pages || 1
            }));
            setError('');
            
            // 初始加载后，启用分页 effect
            if (options.useInitialRequest && !apiPageEffectReadyRef.current) {
                apiPageEffectReadyRef.current = true;
            }
        } catch (err) {
            console.error('Failed to load API call logs:', err);
            setError(err.message);
        } finally {
            setApiLoading(false);
        }
    };

    const loadSchedulerLogs = async (page = schedPagination.page, options = {}) => {
        setSchedLoading(true);
        try {
            const data = await getSchedulerLogsData(page, options.useInitialRequest);
            setSchedulerLogs(data.logs || []);
            setSchedPagination(prev => ({
                ...prev,
                page: data.page || page,
                total: data.total || 0,
                totalPages: data.total_pages || 1
            }));
            setError('');
            
            // 初始加载后，启用分页 effect
            if (options.useInitialRequest && !schedPageEffectReadyRef.current) {
                schedPageEffectReadyRef.current = true;
            }
        } catch (err) {
            console.error('Failed to load scheduler logs:', err);
            setError(err.message);
        } finally {
            setSchedLoading(false);
        }
    };

    const loadLoginLogs = async (page = loginPagination.page, options = {}) => {
        setLoginLoading(true);
        try {
            const data = await getLoginLogsData(page, options.useInitialRequest);
            setLoginLogs(data.logs || []);
            setLoginPagination(prev => ({
                ...prev,
                page: data.page || page,
                total: data.total || 0,
                totalPages: data.total_pages || 1
            }));    
            setError('');
            
            if (options.useInitialRequest && !loginPageEffectReadyRef.current) {
                loginPageEffectReadyRef.current = true;
            }
        } catch (err) {
            console.error('Failed to load login logs:', err);
            setError(err.message);
        } finally {
            setLoginLoading(false);
        }
    };

    const handleTriggerCheck = async () => {
        setTriggering(true);
        try {
            await api.triggerSchedulerCheck();
            setTimeout(async () => {
                await loadSchedulerLogs(schedPagination.page);
                setTriggering(false);
            }, 2000);
        } catch (err) {
            setError(err.message);
            setTriggering(false);
        }
    };

    const handleRefresh = async () => {
        if (activeTab === 'api') {
            if (apiPagination.page === 1) {
                await loadAPILogs(1);
            } else {
                setApiPagination(prev => ({ ...prev, page: 1 }));
            }
        } else if (activeTab === 'scheduler') {
            if (schedPagination.page === 1) {
                await loadSchedulerLogs(1);
            } else {
                setSchedPagination(prev => ({ ...prev, page: 1 }));
            }
        } else {
            if (loginPagination.page === 1) {
                await loadLoginLogs(1);
            } else {
                setLoginPagination(prev => ({ ...prev, page: 1 }));
            }
        }
    };

    const toggleExpand = (logId) => {
        setExpandedLogs(prev => {
            const newSet = new Set(prev);
            if (newSet.has(logId)) {
                newSet.delete(logId);
            } else {
                newSet.add(logId);
            }
            return newSet;
        });
    };

    const getMethodColor = (method) => {
        const colors = {
            GET: '#10b981',
            POST: '#3b82f6',
            PUT: '#f59e0b',
            DELETE: '#ef4444',
            PATCH: '#8b5cf6',
            LEGACY: '#6b7280'
        };
        return colors[method] || '#6b7280';
    };

    const getStatusCodeColor = (statusCode) => {
        if (statusCode >= 200 && statusCode < 300) return '#10b981';
        if (statusCode >= 300 && statusCode < 400) return '#3b82f6';
        if (statusCode >= 400 && statusCode < 500) return '#f59e0b';
        if (statusCode >= 500) return '#ef4444';
        return '#6b7280';
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

    const formatJSON = (jsonString) => {
        try {
            const obj = JSON.parse(jsonString);
            return JSON.stringify(obj, null, 2);
        } catch {
            return jsonString;
        }
    };

    const handlePageChange = (newPage) => {
        if (activeTab === 'api') {
            if (newPage >= 1 && newPage <= apiPagination.totalPages) {
                setApiPagination(prev => ({ ...prev, page: newPage }));
            }
        } else if (activeTab === 'scheduler') {
            if (newPage >= 1 && newPage <= schedPagination.totalPages) {
                setSchedPagination(prev => ({ ...prev, page: newPage }));
            }
        } else {
            if (newPage >= 1 && newPage <= loginPagination.totalPages) {
                setLoginPagination(prev => ({ ...prev, page: newPage }));
            }
        }
    };

    return (
        <div>
            <div ref={topRef} style={{ marginBottom: '1.5rem' }}>
                <div className="page-title-row" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
                    <div style={{ minWidth: 0, flex: 1 }}>
                        <h2 style={{ fontSize: '1.5rem', fontWeight: 'bold', letterSpacing: '-0.02em', margin: 0 }}>
                            {pageTitle}
                        </h2>
                        <p style={{ color: 'var(--text-secondary)', fontSize: '13px', marginTop: '0.25rem', marginBottom: 0 }}>
                            {pageSubtitle}
                        </p>
                    </div>
                    <div className="page-actions-bar" style={{ display: 'flex', gap: '0.5rem' }}>
                        <button
                            onClick={handleRefresh}
                            disabled={apiLoading || schedLoading}
                            className="btn btn-secondary"
                            style={{ height: '34px', padding: '0 10px' }}
                            title={t.common.refresh}
                        >
                            <RefreshCw size={14} className={apiLoading || schedLoading ? "spin" : ""} />
                        </button>
                        {activeTab === 'scheduler' && (
                            <button
                                onClick={handleTriggerCheck}
                                disabled={triggering}
                                className="btn btn-primary"
                                style={{ height: '34px', fontSize: '13px' }}
                                title={t.logsManagement.triggerCheck}
                            >
                                <Play size={14} />
                            </button>
                        )}
                    </div>
                </div>
            </div>

            {/* Tab Navigation */}
            <div className="tab-nav" style={{
                display: 'flex',
                gap: '1.5rem',
                marginBottom: '1.5rem',
                borderBottom: '1px solid var(--border-color)'
            }}>
                <button
                    className={`tab-nav-btn${activeTab === 'api' ? ' active' : ''}`}
                    onClick={() => setActiveTab('api')}
                    style={{
                        padding: '0.625rem 0',
                        background: 'none',
                        border: 'none',
                        borderBottom: activeTab === 'api' ? '2px solid var(--text-primary)' : '2px solid transparent',
                        color: activeTab === 'api' ? 'var(--text-primary)' : 'var(--text-secondary)',
                        fontWeight: activeTab === 'api' ? '500' : '400',
                        cursor: 'pointer',
                        transition: 'all 0.15s',
                        display: 'flex',
                        alignItems: 'center',
                        gap: '0.375rem',
                        fontSize: '14px'
                    }}
                >
                    <Activity size={14} />
                    {t.logsManagement.operationTab} <span style={{ color: 'var(--text-tertiary)', fontSize: '12px' }}>{apiPagination.total}</span>
                </button>
                <button
                    className={`tab-nav-btn${activeTab === 'scheduler' ? ' active' : ''}`}
                    onClick={() => setActiveTab('scheduler')}
                    style={{
                        padding: '0.625rem 0',
                        background: 'none',
                        border: 'none',
                        borderBottom: activeTab === 'scheduler' ? '2px solid var(--text-primary)' : '2px solid transparent',
                        color: activeTab === 'scheduler' ? 'var(--text-primary)' : 'var(--text-secondary)',
                        fontWeight: activeTab === 'scheduler' ? '500' : '400',
                        cursor: 'pointer',
                        transition: 'all 0.15s',
                        display: 'flex',
                        alignItems: 'center',
                        gap: '0.375rem',
                        fontSize: '14px'
                    }}
                >
                    <Clock size={14} />
                    {t.logsManagement.schedulerTitle} <span style={{ color: 'var(--text-tertiary)', fontSize: '12px' }}>{schedPagination.total}</span>
                </button>
                <button
                    className={`tab-nav-btn${activeTab === 'login' ? ' active' : ''}`}
                    onClick={() => setActiveTab('login')}
                    style={{
                        padding: '0.625rem 0',
                        background: 'none',
                        border: 'none',
                        borderBottom: activeTab === 'login' ? '2px solid var(--text-primary)' : '2px solid transparent',
                        color: activeTab === 'login' ? 'var(--text-primary)' : 'var(--text-secondary)',
                        fontWeight: activeTab === 'login' ? '500' : '400',
                        cursor: 'pointer',
                        transition: 'all 0.15s',
                        display: 'flex',
                        alignItems: 'center',
                        gap: '0.375rem',
                        fontSize: '14px'
                    }}
                >
                    <LogIn size={14} />
                    {t.logsManagement.loginTitle} <span style={{ color: 'var(--text-tertiary)', fontSize: '12px' }}>{loginPagination.total}</span>
                </button>
            </div>

            {error && (
                <div style={{ 
                    color: 'var(--danger)', 
                    marginBottom: '1rem', 
                    padding: '0.75rem 1rem', 
                    backgroundColor: 'rgba(255, 0, 0, 0.05)', 
                    border: '1px solid rgba(255, 0, 0, 0.15)',
                    borderRadius: 'var(--radius-sm)',
                    fontSize: '14px'
                }}>
                    {error}
                </div>
            )}

            {((activeTab === 'api' && apiLoading) || (activeTab === 'scheduler' && schedLoading) || (activeTab === 'login' && loginLoading)) ? (
                <div style={{ display: 'flex', justifyContent: 'center', padding: '4rem 0' }}>
                    <div className="spinner"></div>
                </div>
            ) : (
                <>
                    {/* API Logs */}
                    {activeTab === 'api' && (
                        <>
                            <div className="logs-list" style={{ display: 'grid', gap: '0.5rem' }}>
                                {apiLogs.map(log => {
                            const isExpanded = expandedLogs.has(log.id);
                            const hasRequestBody = log.request_body && log.request_body.trim();
                            const hasResponseBody = log.response_body && log.response_body.trim();
                            
                            return (
                                <div key={log.id} className="domain-list-card" style={{
                                    padding: '0.875rem 1rem',
                                    cursor: 'default'
                                }}>
                                    <div style={{ display: 'flex', alignItems: 'flex-start', gap: '0.5rem', flexDirection: 'column' }}>
                                        {/* Top Row: Status + Method + Path */}
                                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', width: '100%', flexWrap: 'wrap' }}>
                                            {/* Status Code */}
                                            <span style={{ 
                                                fontSize: '11px', 
                                                fontWeight: '700',
                                                color: getStatusCodeColor(log.status_code),
                                                padding: '2px 6px',
                                                backgroundColor: `${getStatusCodeColor(log.status_code)}15`,
                                                borderRadius: 'var(--radius-sm)',
                                                fontFamily: 'monospace',
                                                minWidth: '40px',
                                                textAlign: 'center',
                                                flexShrink: 0
                                            }}>
                                                {log.status_code}
                                            </span>

                                            {/* Method */}
                                            <span style={{ 
                                                fontSize: '11px', 
                                                fontWeight: '700',
                                                color: getMethodColor(log.method),
                                                padding: '2px 6px',
                                                backgroundColor: `${getMethodColor(log.method)}15`,
                                                borderRadius: 'var(--radius-sm)',
                                                fontFamily: 'monospace',
                                                textTransform: 'uppercase',
                                                letterSpacing: '0.05em',
                                                flexShrink: 0
                                            }}>
                                                {log.method}
                                            </span>

                                            {/* Path - takes remaining space */}
                                            <span className="font-mono" style={{ 
                                                fontSize: '13px', 
                                                fontWeight: '600',
                                                color: 'var(--text-primary)',
                                                flex: 1,
                                                minWidth: 0,
                                                wordBreak: 'break-all'
                                            }}
                                            title={log.path}
                                            >
                                                {log.path}
                                            </span>

                                            {/* Expand Button */}
                                            <button
                                                onClick={() => toggleExpand(log.id)}
                                                style={{
                                                    background: 'none',
                                                    border: 'none',
                                                    cursor: 'pointer',
                                                    color: 'var(--text-secondary)',
                                                    padding: '0.25rem',
                                                    display: 'flex',
                                                    alignItems: 'center',
                                                    flexShrink: 0,
                                                    marginLeft: 'auto'
                                                }}
                                            >
                                                {isExpanded ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
                                            </button>
                                        </div>

                                        {/* Bottom Row: Meta Info */}
                                        <div style={{ 
                                            fontSize: '12px', 
                                            color: 'var(--text-tertiary)',
                                            display: 'flex',
                                            alignItems: 'center',
                                            gap: '0.75rem',
                                            flexWrap: 'wrap',
                                            width: '100%',
                                            paddingLeft: '0.25rem'
                                        }}>
                                            {/* Time */}
                                            <div style={{ display: 'flex', alignItems: 'center', gap: '0.25rem', whiteSpace: 'nowrap' }}>
                                                <Clock size={12} />
                                                <span>{formatDate(log.created_at)}</span>
                                            </div>

                                            {/* Duration */}
                                            <div style={{ display: 'flex', alignItems: 'center', gap: '0.25rem', whiteSpace: 'nowrap' }}>
                                                ⏱ {log.duration_ms}ms
                                            </div>

                                            {/* Username */}
                                            {log.username && (
                                                <div style={{ display: 'flex', alignItems: 'center', gap: '0.25rem', whiteSpace: 'nowrap' }}>
                                                    <User size={12} />
                                                    {log.username}
                                                </div>
                                            )}

                                            {/* IP Address */}
                                            {log.ip_address && (
                                                <div className="font-mono" style={{ 
                                                    display: 'flex', 
                                                    alignItems: 'center', 
                                                    gap: '0.25rem',
                                                    whiteSpace: 'nowrap'
                                                }}>
                                                    <Globe size={12} />
                                                    {log.ip_address}
                                                </div>
                                            )}

                                            {/* Query - can wrap to new line if needed */}
                                            {log.query && (
                                                <div className="font-mono" style={{ 
                                                    color: 'var(--text-secondary)',
                                                    wordBreak: 'break-all',
                                                    flex: '1 1 100%'
                                                }}
                                                title={`?${log.query}`}
                                                >
                                                    ?{log.query}
                                                </div>
                                            )}

                                            {/* Error Message - full width */}
                                            {log.error_message && (
                                                <div style={{ 
                                                    color: '#ef4444', 
                                                    fontWeight: '500',
                                                    flex: '1 1 100%',
                                                    wordBreak: 'break-word'
                                                }}>
                                                    ⚠ {log.error_message}
                                                </div>
                                            )}
                                        </div>
                                    </div>

                                    {/* Expanded Details */}
                                    {isExpanded && (
                                        <div style={{ marginTop: '0.75rem', display: 'grid', gap: '0.75rem', borderTop: '1px solid var(--border-color)', paddingTop: '0.75rem' }}>
                                            {/* Request Headers */}
                                            {log.request_headers && (
                                                <div>
                                                    <div style={{ 
                                                        fontSize: '11px', 
                                                        fontWeight: '600', 
                                                        color: 'var(--text-secondary)',
                                                        marginBottom: '0.25rem',
                                                        textTransform: 'uppercase',
                                                        letterSpacing: '0.05em'
                                                    }}>
                                                        {t.logsManagement.requestHeaders}
                                                    </div>
                                                    <pre style={{ 
                                                        fontSize: '12px',
                                                        backgroundColor: 'var(--bg-secondary)',
                                                        padding: '0.5rem 0.75rem',
                                                        borderRadius: 'var(--radius-sm)',
                                                        overflow: 'auto',
                                                        overflowWrap: 'break-word',
                                                        wordBreak: 'break-word',
                                                        whiteSpace: 'pre-wrap',
                                                        margin: 0,
                                                        fontFamily: 'monospace',
                                                        color: 'var(--text-primary)',
                                                        maxHeight: '200px',
                                                        lineHeight: '1.5',
                                                        border: '1px solid var(--border-color)'
                                                    }}>
                                                        {formatJSON(log.request_headers)}
                                                    </pre>
                                                </div>
                                            )}

                                            {/* Request Body */}
                                            {hasRequestBody && (
                                                <div>
                                                    <div style={{ 
                                                        fontSize: '11px', 
                                                        fontWeight: '600', 
                                                        color: 'var(--text-secondary)',
                                                        marginBottom: '0.25rem',
                                                        textTransform: 'uppercase',
                                                        letterSpacing: '0.05em'
                                                    }}>
                                                        {t.logsManagement.requestBody}
                                                    </div>
                                                    <pre style={{ 
                                                        fontSize: '12px',
                                                        backgroundColor: 'var(--bg-secondary)',
                                                        padding: '0.5rem 0.75rem',
                                                        borderRadius: 'var(--radius-sm)',
                                                        overflow: 'auto',
                                                        overflowWrap: 'break-word',
                                                        wordBreak: 'break-word',
                                                        whiteSpace: 'pre-wrap',
                                                        margin: 0,
                                                        fontFamily: 'monospace',
                                                        color: 'var(--text-primary)',
                                                        maxHeight: '300px',
                                                        lineHeight: '1.5',
                                                        border: '1px solid var(--border-color)'
                                                    }}>
                                                        {formatJSON(log.request_body)}
                                                    </pre>
                                                </div>
                                            )}

                                            {/* Response Body */}
                                            {hasResponseBody && (
                                                <div>
                                                    <div style={{ 
                                                        fontSize: '11px', 
                                                        fontWeight: '600', 
                                                        color: 'var(--text-secondary)',
                                                        marginBottom: '0.25rem',
                                                        textTransform: 'uppercase',
                                                        letterSpacing: '0.05em'
                                                    }}>
                                                        {t.logsManagement.responseBody}
                                                    </div>
                                                    <pre style={{ 
                                                        fontSize: '12px',
                                                        backgroundColor: 'var(--bg-secondary)',
                                                        padding: '0.5rem 0.75rem',
                                                        borderRadius: 'var(--radius-sm)',
                                                        overflow: 'auto',
                                                        overflowWrap: 'break-word',
                                                        wordBreak: 'break-word',
                                                        whiteSpace: 'pre-wrap',
                                                        margin: 0,
                                                        fontFamily: 'monospace',
                                                        color: 'var(--text-primary)',
                                                        maxHeight: '300px',
                                                        lineHeight: '1.5',
                                                        border: '1px solid var(--border-color)'
                                                    }}>
                                                        {formatJSON(log.response_body)}
                                                    </pre>
                                                </div>
                                            )}

                                            {/* User Agent */}
                                            {log.user_agent && (
                                                <div>
                                                    <div style={{ 
                                                        fontSize: '11px', 
                                                        fontWeight: '600', 
                                                        color: 'var(--text-secondary)',
                                                        marginBottom: '0.25rem',
                                                        textTransform: 'uppercase',
                                                        letterSpacing: '0.05em'
                                                    }}>
                                                        User Agent
                                                    </div>
                                                    <div className="font-mono" style={{ 
                                                        fontSize: '12px',
                                                        backgroundColor: 'var(--bg-secondary)',
                                                        padding: '0.5rem 0.75rem',
                                                        borderRadius: 'var(--radius-sm)',
                                                        color: 'var(--text-tertiary)',
                                                        wordBreak: 'break-all',
                                                        border: '1px solid var(--border-color)'
                                                    }}>
                                                        {log.user_agent}
                                                    </div>
                                                </div>
                                            )}
                                        </div>
                                    )}
                                </div>
                            );
                        })}
                        {apiLogs.length === 0 && (
                            <div className="domain-list-card" style={{ textAlign: 'center', padding: '4rem', color: 'var(--text-secondary)', cursor: 'default' }}>
                                {t.logsManagement.noOperationLogs}
                            </div>
                        )}
                    </div>

                    {/* API Logs Pagination */}
                    {apiPagination.totalPages > 1 && (
                        <div className="pagination-controls" style={{
                            display: 'flex',
                            justifyContent: 'center',
                            alignItems: 'center',
                            gap: '0.75rem',
                            marginTop: '1.5rem',
                            padding: '1rem 0 0 0',
                            borderTop: '1px solid var(--border-color)'
                        }}>
                            <button
                                onClick={() => handlePageChange(apiPagination.page - 1)}
                                disabled={apiPagination.page === 1}
                                className="btn btn-secondary"
                                style={{ 
                                    minWidth: '80px',
                                    height: '34px',
                                    fontSize: '13px',
                                    opacity: apiPagination.page === 1 ? 0.5 : 1,
                                    cursor: apiPagination.page === 1 ? 'not-allowed' : 'pointer'
                                }}
                            >
                                {t.common.previousPage}
                            </button>
                            <div style={{ 
                                display: 'flex', 
                                alignItems: 'center', 
                                gap: '0.5rem',
                                padding: '0 0.75rem',
                                height: '34px',
                                backgroundColor: 'var(--bg-secondary)',
                                border: '1px solid var(--border-color)',
                                borderRadius: 'var(--radius-sm)',
                                fontSize: '13px'
                            }}>
                                <span style={{ color: 'var(--text-secondary)' }}>
                                    {t.logsManagement.page}
                                </span>
                                <span style={{ 
                                    color: 'var(--text-primary)', 
                                    fontWeight: '600',
                                    minWidth: '2ch',
                                    textAlign: 'center'
                                }}>
                                    {apiPagination.page}
                                </span>
                                <span style={{ color: 'var(--text-secondary)' }}>/</span>
                                <span style={{ color: 'var(--text-primary)', fontWeight: '500' }}>
                                    {apiPagination.totalPages}
                                </span>
                                <span style={{ color: 'var(--text-secondary)' }}>
                                    {t.logsManagement.pages}
                                </span>
                                <span style={{ 
                                    marginLeft: '0.5rem', 
                                    paddingLeft: '0.5rem', 
                                    borderLeft: '1px solid var(--border-color)',
                                    color: 'var(--text-tertiary)',
                                    fontSize: '11px'
                                }}>
                                    {t.common.totalItems.replace('{count}', apiPagination.total)}
                                </span>
                            </div>
                            <button
                                onClick={() => handlePageChange(apiPagination.page + 1)}
                                disabled={apiPagination.page === apiPagination.totalPages}
                                className="btn btn-secondary"
                                style={{ 
                                    minWidth: '80px',
                                    height: '34px',
                                    fontSize: '13px',
                                    opacity: apiPagination.page === apiPagination.totalPages ? 0.5 : 1,
                                    cursor: apiPagination.page === apiPagination.totalPages ? 'not-allowed' : 'pointer'
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
                            <div className="logs-list" style={{ display: 'grid', gap: '0.5rem' }}>
                                {schedulerLogs.map(log => {
                                    let details = {};
                                    try {
                                        details = log.details ? JSON.parse(log.details) : {};
                                    } catch (e) {
                                        details = {};
                                    }

                                    return (
                                        <div key={log.id} className="domain-list-card" style={{
                                            padding: '0.875rem 1rem',
                                            cursor: 'default'
                                        }}>
                                            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                                                <span style={{ 
                                                    width: '8px',
                                                    height: '8px',
                                                    borderRadius: '50%',
                                                    backgroundColor: getStatusColor(log.status),
                                                    flexShrink: 0
                                                }} />
                                                <div style={{ flex: 1, minWidth: 0 }}>
                                                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', flexWrap: 'wrap', marginBottom: '0.25rem' }}>
                                                        <span style={{ 
                                                            fontSize: '14px', 
                                                            fontWeight: '600',
                                                            color: 'var(--text-primary)'
                                                        }}>
                                                            {t.logsManagement?.taskNames?.[log.task_name] || log.task_name}
                                                        </span>
                                                        <span className="badge" style={{ 
                                                            fontSize: '11px', 
                                                            fontWeight: '600',
                                                            color: getStatusColor(log.status),
                                                            backgroundColor: `${getStatusColor(log.status)}15`,
                                                            height: '20px',
                                                            padding: '1px 6px'
                                                        }}>
                                                            {log.status}
                                                        </span>
                                                    </div>
                                                    
                                                    {log.message && (
                                                        <div style={{ 
                                                            fontSize: '13px', 
                                                            color: 'var(--text-secondary)',
                                                            marginBottom: '0.25rem'
                                                        }}>
                                                            {log.message}
                                                        </div>
                                                    )}

                                                    {details && Object.keys(details).length > 0 && (
                                                        <div style={{ 
                                                            fontSize: '12px', 
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
                                                                    backgroundColor: 'rgba(255, 0, 0, 0.05)',
                                                                    borderRadius: 'var(--radius-sm)',
                                                                    color: '#ef4444',
                                                                    border: '1px solid rgba(255, 0, 0, 0.1)'
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
                                                    <div style={{ fontSize: '11px', color: 'var(--text-tertiary)' }}>
                                                        {formatDate(log.started_at)}
                                                    </div>
                                                    {log.duration_ms && (
                                                        <div style={{ fontSize: '11px', color: 'var(--text-tertiary)' }}>
                                                            ⏱ {log.duration_ms}ms
                                                        </div>
                                                    )}
                                                </div>
                                            </div>
                                        </div>
                                    );
                                })}
                                {schedulerLogs.length === 0 && (
                                    <div className="domain-list-card" style={{ textAlign: 'center', padding: '4rem', color: 'var(--text-secondary)', cursor: 'default' }}>
                                        {t.logsManagement.noSchedulerLogs}
                                    </div>
                                )}
                            </div>

                            {/* Scheduler Logs Pagination */}
                            {schedPagination.totalPages > 1 && (
                                <div className="pagination-controls" style={{
                                    display: 'flex',
                                    justifyContent: 'center',
                                    alignItems: 'center',
                                    gap: '0.75rem',
                                    marginTop: '1.5rem',
                                    padding: '1rem 0 0 0',
                                    borderTop: '1px solid var(--border-color)'
                                }}>
                                    <button
                                        onClick={() => handlePageChange(schedPagination.page - 1)}
                                        disabled={schedPagination.page === 1}
                                        className="btn btn-secondary"
                                        style={{ 
                                            minWidth: '80px',
                                            height: '34px',
                                            fontSize: '13px',
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
                                        padding: '0 0.75rem',
                                        height: '34px',
                                        backgroundColor: 'var(--bg-secondary)',
                                        border: '1px solid var(--border-color)',
                                        borderRadius: 'var(--radius-sm)',
                                        fontSize: '13px'
                                    }}>
                                        <span style={{ color: 'var(--text-secondary)' }}>
                                            {t.logsManagement.page}
                                        </span>
                                        <span style={{ 
                                            color: 'var(--text-primary)', 
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
                                        <span style={{ color: 'var(--text-secondary)' }}>
                                            {t.logsManagement.pages}
                                        </span>
                                        <span style={{ 
                                            marginLeft: '0.5rem', 
                                            paddingLeft: '0.5rem', 
                                            borderLeft: '1px solid var(--border-color)',
                                            color: 'var(--text-tertiary)',
                                            fontSize: '11px'
                                        }}>
                                            {t.common.totalItems.replace('{count}', schedPagination.total)}
                                        </span>
                                    </div>
                                    <button
                                        onClick={() => handlePageChange(schedPagination.page + 1)}
                                        disabled={schedPagination.page === schedPagination.totalPages}
                                        className="btn btn-secondary"
                                        style={{ 
                                            minWidth: '80px',
                                            height: '34px',
                                            fontSize: '13px',
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

                    {/* Login Logs */}
                    {activeTab === 'login' && (
                        <>
                            <div className="logs-list" style={{ display: 'grid', gap: '0.5rem' }}>
                                {loginLogs.map(log => (
                                    <div key={log.id} className="domain-list-card" style={{
                                        padding: '0.875rem 1rem',
                                        cursor: 'default'
                                    }}>
                                        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                                            <LogIn size={16} style={{ 
                                                color: log.status === 'success' ? '#10b981' : '#ef4444',
                                                flexShrink: 0
                                            }} />
                                            <div style={{ flex: 1, minWidth: 0 }}>
                                                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', flexWrap: 'wrap', marginBottom: '0.25rem' }}>
                                                    <span className="badge" style={{ 
                                                        fontSize: '11px', 
                                                        fontWeight: '600',
                                                        color: log.status === 'success' ? '#10b981' : '#ef4444',
                                                        backgroundColor: log.status === 'success' ? 'rgba(16, 185, 129, 0.1)' : 'rgba(239, 68, 68, 0.1)',
                                                        height: '20px',
                                                        padding: '1px 6px'
                                                    }}>
                                                        {log.status === 'success' 
                                                            ? t.logsManagement.loginSuccess
                                                            : t.logsManagement.loginFailed}
                                                    </span>
                                                    <span style={{ fontSize: '14px', color: 'var(--text-primary)', fontWeight: '600' }}>
                                                        {log.username}
                                                    </span>
                                                </div>
                                                <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', flexWrap: 'wrap', fontSize: '12px', color: 'var(--text-secondary)' }}>
                                                    {log.ip_address && (
                                                        <span className="font-mono" style={{ display: 'inline-flex', alignItems: 'center', gap: '0.25rem' }}>
                                                            <Globe size={12} style={{ flexShrink: 0 }} />
                                                            {log.ip_address}
                                                        </span>
                                                    )}
                                                    {log.ip_location && (
                                                        <span style={{ display: 'inline-flex', alignItems: 'center', gap: '0.25rem', color: 'var(--text-tertiary)' }}>
                                                            <MapPin size={12} style={{ flexShrink: 0 }} />
                                                            {log.ip_location}
                                                        </span>
                                                    )}
                                                    {log.device && (
                                                        <span style={{ color: 'var(--text-tertiary)' }}>
                                                            {log.device}
                                                        </span>
                                                    )}
                                                </div>
                                                {log.message && log.status === 'failed' && (
                                                    <div style={{ 
                                                        fontSize: '11px', 
                                                        color: '#ef4444',
                                                        marginTop: '0.25rem'
                                                    }}>
                                                        {log.message}
                                                    </div>
                                                )}
                                            </div>
                                            <div style={{ textAlign: 'right', flexShrink: 0, display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: '0.25rem' }}>
                                                <div style={{ fontSize: '11px', color: 'var(--text-tertiary)' }}>
                                                    {formatDate(log.created_at)}
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                ))}
                                {loginLogs.length === 0 && (
                                    <div className="domain-list-card" style={{ textAlign: 'center', padding: '4rem', color: 'var(--text-secondary)', cursor: 'default' }}>
                                        {t.logsManagement.noLoginLogs}
                                    </div>
                                )}
                            </div>

                            {/* Login Logs Pagination */}
                            {loginPagination.totalPages > 1 && (
                                <div className="pagination-controls" style={{
                                    display: 'flex',
                                    justifyContent: 'center',
                                    alignItems: 'center',
                                    gap: '0.75rem',
                                    marginTop: '1.5rem',
                                    padding: '1rem 0 0 0',
                                    borderTop: '1px solid var(--border-color)'
                                }}>
                                    <button
                                        onClick={() => handlePageChange(loginPagination.page - 1)}
                                        disabled={loginPagination.page === 1}
                                        className="btn btn-secondary"
                                        style={{ 
                                            minWidth: '80px',
                                            height: '34px',
                                            fontSize: '13px',
                                            opacity: loginPagination.page === 1 ? 0.5 : 1,
                                            cursor: loginPagination.page === 1 ? 'not-allowed' : 'pointer'
                                        }}
                                    >
                                        {t.common.previousPage}
                                    </button>
                                    <div style={{ 
                                        display: 'flex', 
                                        alignItems: 'center', 
                                        gap: '0.5rem',
                                        padding: '0 0.75rem',
                                        height: '34px',
                                        backgroundColor: 'var(--bg-secondary)',
                                        border: '1px solid var(--border-color)',
                                        borderRadius: 'var(--radius-sm)',
                                        fontSize: '13px'
                                    }}>
                                        <span style={{ color: 'var(--text-secondary)' }}>
                                            {t.logsManagement.page}
                                        </span>
                                        <span style={{ 
                                            color: 'var(--text-primary)', 
                                            fontWeight: '600',
                                            minWidth: '2ch',
                                            textAlign: 'center'
                                        }}>
                                            {loginPagination.page}
                                        </span>
                                        <span style={{ color: 'var(--text-secondary)' }}>/</span>
                                        <span style={{ color: 'var(--text-primary)', fontWeight: '500' }}>
                                            {loginPagination.totalPages}
                                        </span>
                                        <span style={{ color: 'var(--text-secondary)' }}>
                                            {t.logsManagement.pages}
                                        </span>
                                        <span style={{ 
                                            marginLeft: '0.5rem', 
                                            paddingLeft: '0.5rem', 
                                            borderLeft: '1px solid var(--border-color)',
                                            color: 'var(--text-tertiary)',
                                            fontSize: '11px'
                                        }}>
                                            {t.common.totalItems.replace('{count}', loginPagination.total)}
                                        </span>
                                    </div>
                                    <button
                                        onClick={() => handlePageChange(loginPagination.page + 1)}
                                        disabled={loginPagination.page === loginPagination.totalPages}
                                        className="btn btn-secondary"
                                        style={{ 
                                            minWidth: '80px',
                                            height: '34px',
                                            fontSize: '13px',
                                            opacity: loginPagination.page === loginPagination.totalPages ? 0.5 : 1,
                                            cursor: loginPagination.page === loginPagination.totalPages ? 'not-allowed' : 'pointer'
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

export default APILogsManagement;
