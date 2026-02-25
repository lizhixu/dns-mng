import { useState, useEffect } from 'react';
import { api } from '../api';
import { Clock, Activity, ChevronDown, ChevronUp, RefreshCw } from 'lucide-react';

export default function Logs() {
    const [logs, setLogs] = useState([]);
    const [loading, setLoading] = useState(true);
    const [expandedLogs, setExpandedLogs] = useState(new Set());

    useEffect(() => {
        loadLogs();
    }, []);

    const loadLogs = async () => {
        setLoading(true);
        try {
            const data = await api.getLogs();
            setLogs(data || []);
        } catch (err) {
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    const toggleExpand = (logId) => {
        const newExpanded = new Set(expandedLogs);
        if (newExpanded.has(logId)) {
            newExpanded.delete(logId);
        } else {
            newExpanded.add(logId);
        }
        setExpandedLogs(newExpanded);
    };

    const getActionText = (action) => {
        const actions = {
            create: '创建',
            update: '更新',
            delete: '删除',
            login: '登录',
            login_failed: '登录失败',
            register: '注册',
            update_password: '修改密码'
        };
        return actions[action] || action;
    };

    const getActionColor = (action) => {
        const colors = {
            create: '#10b981',
            update: '#3b82f6',
            delete: '#ef4444',
            login: '#10b981',
            login_failed: '#ef4444',
            register: '#8b5cf6',
            update_password: '#f59e0b'
        };
        return colors[action] || 'var(--accent-primary)';
    };

    const getResourceText = (resource) => {
        const resources = {
            account: '账户',
            record: 'DNS记录',
            auth: '认证'
        };
        return resources[resource] || resource;
    };

    const renderDetails = (log) => {
        try {
            const details = JSON.parse(log.details);
            const isExpanded = expandedLogs.has(log.id);
            
            // Basic info - handle account, record, and auth
            let summary = details.name || details.node_name || details.username || log.resource_id;
            if (details.record_type) summary += ` (${details.record_type})`;
            else if (details.type) summary += ` (${details.type})`;
            
            // Check if there are changes or additional details
            const hasChanges = details.changes && Object.keys(details.changes).length > 0;
            const hasMoreInfo = details.content || details.ttl || details.domain || details.provider || details.reason;
            
            return (
                <div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                        <span style={{ color: 'var(--text-tertiary)' }}>{summary}</span>
                        {(hasChanges || hasMoreInfo) && (
                            <button
                                onClick={(e) => {
                                    e.stopPropagation();
                                    toggleExpand(log.id);
                                }}
                                style={{
                                    background: 'none',
                                    border: 'none',
                                    padding: '0.25rem',
                                    cursor: 'pointer',
                                    color: 'var(--accent-primary)',
                                    display: 'flex',
                                    alignItems: 'center'
                                }}
                            >
                                {isExpanded ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
                            </button>
                        )}
                    </div>
                    
                    {isExpanded && (
                        <div style={{ 
                            marginTop: '0.75rem', 
                            padding: '0.75rem', 
                            background: 'rgba(255,255,255,0.03)',
                            borderRadius: '0.375rem',
                            fontSize: '0.75rem'
                        }}>
                            {details.reason && (
                                <div style={{ marginBottom: '0.5rem' }}>
                                    <span style={{ color: 'var(--text-secondary)' }}>原因：</span>
                                    <span style={{ color: '#ef4444', marginLeft: '0.5rem' }}>{details.reason}</span>
                                </div>
                            )}
                            {details.provider && (
                                <div style={{ marginBottom: '0.5rem' }}>
                                    <span style={{ color: 'var(--text-secondary)' }}>服务商：</span>
                                    <span style={{ color: 'var(--text-primary)', marginLeft: '0.5rem' }}>{details.provider}</span>
                                </div>
                            )}
                            {details.domain && (
                                <div style={{ marginBottom: '0.5rem' }}>
                                    <span style={{ color: 'var(--text-secondary)' }}>域名：</span>
                                    <span style={{ color: 'var(--text-primary)', marginLeft: '0.5rem' }}>{details.domain}</span>
                                </div>
                            )}
                            {details.content && (
                                <div style={{ marginBottom: '0.5rem' }}>
                                    <span style={{ color: 'var(--text-secondary)' }}>值：</span>
                                    <span style={{ color: 'var(--text-primary)', marginLeft: '0.5rem', fontFamily: 'monospace' }}>{details.content}</span>
                                </div>
                            )}
                            {details.ttl && (
                                <div style={{ marginBottom: '0.5rem' }}>
                                    <span style={{ color: 'var(--text-secondary)' }}>TTL：</span>
                                    <span style={{ color: 'var(--text-primary)', marginLeft: '0.5rem' }}>{details.ttl}</span>
                                </div>
                            )}
                            {hasChanges && (
                                <div style={{ marginTop: '0.75rem', paddingTop: '0.75rem', borderTop: '1px solid rgba(255,255,255,0.05)' }}>
                                    <div style={{ color: 'var(--text-secondary)', marginBottom: '0.5rem', fontWeight: '500' }}>变更内容：</div>
                                    {Object.entries(details.changes).map(([key, change]) => (
                                        <div key={key} style={{ marginBottom: '0.5rem', paddingLeft: '0.5rem' }}>
                                            <div style={{ color: 'var(--text-tertiary)', marginBottom: '0.25rem' }}>{key}:</div>
                                            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', paddingLeft: '0.5rem' }}>
                                                <span style={{ color: '#ef4444', fontFamily: 'monospace' }}>
                                                    {typeof change.old === 'boolean' ? (change.old ? '启用' : '禁用') : change.old}
                                                </span>
                                                <span style={{ color: 'var(--text-tertiary)' }}>→</span>
                                                <span style={{ color: '#10b981', fontFamily: 'monospace' }}>
                                                    {typeof change.new === 'boolean' ? (change.new ? '启用' : '禁用') : change.new}
                                                </span>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </div>
                    )}
                </div>
            );
        } catch (e) {
            return <span style={{ color: 'var(--text-tertiary)' }}>{log.resource_id}</span>;
        }
    };

    return (
        <div>
            <div style={{ marginBottom: '2rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
                    <div>
                        <h2 style={{ fontSize: '1.5rem', fontWeight: 'bold' }}>操作日志</h2>
                        <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', marginTop: '0.25rem' }}>
                            查看最近的操作记录，点击展开查看详细信息
                        </p>
                    </div>
                    <button onClick={loadLogs} className="btn btn-secondary" title="刷新">
                        <RefreshCw size={18} className={loading ? "spin" : ""} style={{ animation: loading ? 'spin 1s linear infinite' : 'none' }} />
                    </button>
                </div>
            </div>

            {loading && !logs.length ? (
                <div className="spinner" style={{ margin: '4rem auto' }}></div>
            ) : (
                <div style={{ display: 'grid', gap: '1rem' }}>
                    {logs.map(log => (
                        <div 
                            key={log.id} 
                            className="glass-panel" 
                            style={{ 
                                padding: '1rem 1.25rem',
                                transition: 'all 0.2s',
                                cursor: 'pointer'
                            }}
                            onMouseEnter={(e) => {
                                e.currentTarget.style.transform = 'translateY(-2px)';
                                e.currentTarget.style.boxShadow = '0 4px 12px rgba(0, 0, 0, 0.1)';
                            }}
                            onMouseLeave={(e) => {
                                e.currentTarget.style.transform = 'translateY(0)';
                                e.currentTarget.style.boxShadow = 'none';
                            }}
                        >
                            <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'flex-start' }}>
                                <Activity 
                                    size={20} 
                                    style={{ 
                                        color: getActionColor(log.action), 
                                        flexShrink: 0,
                                        marginTop: '0.125rem'
                                    }} 
                                />
                                <div style={{ flex: 1, minWidth: 0 }}>
                                    <h3 style={{ fontSize: '1rem', fontWeight: '500', margin: 0, marginBottom: '0.25rem' }}>
                                        <span style={{ color: getActionColor(log.action) }}>
                                            {getActionText(log.action)}
                                        </span>
                                        {' '}
                                        <span style={{ color: 'var(--text-secondary)' }}>{getResourceText(log.resource)}</span>
                                    </h3>
                                    {log.details && renderDetails(log)}
                                    <div style={{ 
                                        display: 'flex', 
                                        alignItems: 'center', 
                                        gap: '0.75rem', 
                                        flexWrap: 'wrap',
                                        fontSize: '0.75rem', 
                                        color: 'var(--text-tertiary)',
                                        marginTop: '0.75rem'
                                    }}>
                                        <span style={{ display: 'inline-flex', alignItems: 'center', gap: '0.25rem' }}>
                                            <Clock size={12} />
                                            {new Date(log.created_at).toLocaleString('zh-CN', {
                                                year: 'numeric',
                                                month: '2-digit',
                                                day: '2-digit',
                                                hour: '2-digit',
                                                minute: '2-digit',
                                                second: '2-digit'
                                            })}
                                        </span>
                                        {log.ip_address && (
                                            <>
                                                <span>•</span>
                                                <span>IP: {log.ip_address}</span>
                                            </>
                                        )}
                                        {log.username && (
                                            <>
                                                <span>•</span>
                                                <span>用户: {log.username}</span>
                                            </>
                                        )}
                                    </div>
                                </div>
                            </div>
                        </div>
                    ))}

                    {logs.length === 0 && (
                        <div style={{ textAlign: 'center', padding: '4rem', color: 'var(--text-secondary)' }}>
                            暂无操作记录
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}
