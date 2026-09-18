import { useState, useEffect, useRef } from 'react';
import { api } from '../api';
import { Bell, RefreshCw, ChevronDown, ChevronUp, ExternalLink, Check, CheckCheck, Trash2 } from 'lucide-react';
import { useLanguage } from '../LanguageContext';

const pageSize = 20;

const getMessagesData = (page, useInitialRequest, unreadOnly) => {
    const requestKey = `messages-${unreadOnly ? 'unread-' : ''}${page}`;
    if (!getMessagesData.initialRequests) {
        getMessagesData.initialRequests = {};
    }
    if (useInitialRequest && page === 1 && !getMessagesData.initialRequests[requestKey]) {
        getMessagesData.initialRequests[requestKey] = api.getMessages(page, pageSize, unreadOnly).catch(error => {
            getMessagesData.initialRequests[requestKey] = null;
            throw error;
        }).finally(() => {
            setTimeout(() => {
                if (getMessagesData.initialRequests && getMessagesData.initialRequests[requestKey] === requestKey) {
                    getMessagesData.initialRequests[requestKey] = null;
                }
            }, 1000);
        });
        return getMessagesData.initialRequests[requestKey];
    }
    return api.getMessages(page, pageSize, unreadOnly);
};
getMessagesData.initialRequests = {};

const Messages = () => {
    const { t, language } = useLanguage();
    const [messages, setMessages] = useState([]);
    const [pagination, setPagination] = useState({ total: 0, page: 1, totalPages: 1 });
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [filterUnread, setFilterUnread] = useState(false);
    const [markingAllRead, setMarkingAllRead] = useState(false);
    const topRef = useRef(null);
    const initialLoadRef = useRef(false);
    const fetchedRef = useRef(false);

    const fetchMessages = async (page = 1, unreadOnly = false) => {
        setLoading(true);
        setError('');
        try {
            const data = await getMessagesData(page, initialLoadRef.current && page === 1, unreadOnly);
            setMessages(data.messages || []);
            setPagination({
                total: data.total || 0,
                page: data.page || page,
                totalPages: Math.ceil((data.total || 0) / pageSize),
            });
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (fetchedRef.current) return;
        fetchedRef.current = true;
        initialLoadRef.current = true;
        fetchMessages(1, filterUnread);
        return () => {
            fetchedRef.current = false;
            initialLoadRef.current = false;
        };
    }, []);

    useEffect(() => {
        if (!initialLoadRef.current) return;
        fetchMessages(1, filterUnread);
    }, [filterUnread]);

    const handlePageChange = (page) => {
        fetchMessages(page, filterUnread);
        if (topRef.current) {
            topRef.current.scrollIntoView({ behavior: 'smooth' });
        }
    };

    const handleRefresh = () => {
        fetchMessages(pagination.page, filterUnread);
    };

    const handleMarkRead = async (id) => {
        try {
            await api.markMessageRead(id);
            setMessages(prev => prev.map(msg =>
                msg.id === id ? { ...msg, is_read: true } : msg
            ));
        } catch (err) {
            setError(err.message);
        }
    };

    const handleMarkAllRead = async () => {
        setMarkingAllRead(true);
        setError('');
        try {
            await api.markAllMessagesRead();
            setMessages(prev => prev.map(msg => ({ ...msg, is_read: true })));
        } catch (err) {
            setError(err.message);
        } finally {
            setMarkingAllRead(false);
        }
    };

    const handleDelete = async (id) => {
        try {
            await api.deleteMessage(id);
            setMessages(prev => prev.filter(msg => msg.id !== id));
            setPagination(prev => ({ ...prev, total: prev.total - 1 }));
        } catch (err) {
            setError(err.message);
        }
    };

    const getStatusColor = (daysRemaining) => {
        if (daysRemaining < 0) return 'var(--danger)';
        if (daysRemaining === 0) return 'var(--danger)';
        if (daysRemaining === 1) return '#ff6c42';
        if (daysRemaining <= 7) return '#ff8c42';
        if (daysRemaining <= 30) return '#ffa726';
        return 'var(--text-tertiary)';
    };

    const formatTime = (dateString) => {
        if (!dateString) return '';
        const date = new Date(dateString);
        const now = new Date();
        const diff = Math.floor((now - date) / 1000);
        if (diff < 60) return `${diff}秒前`;
        if (diff < 3600) return `${Math.floor(diff / 60)}分钟前`;
        if (diff < 86400) return `${Math.floor(diff / 3600)}小时前`;
        return date.toLocaleDateString();
    };

    return (
        <div ref={topRef} style={{ padding: '20px 0' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '20px', flexWrap: 'wrap', gap: '12px' }}>
                <div>
                    <h1 style={{ fontSize: '20px', fontWeight: '600', margin: 0 }}>{t.messages.title}</h1>
                    <p style={{ fontSize: '13px', color: 'var(--text-secondary)', margin: '4px 0 0' }}>{t.messages.subtitle}</p>
                </div>
                <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                    <div className="tab-nav" style={{ margin: 0, border: 'none', gap: '1.25rem' }}>
                        <button
                            type="button"
                            onClick={() => setFilterUnread(false)}
                            className={`tab-nav-btn${!filterUnread ? ' active' : ''}`}
                            style={{ padding: '6px 2px', fontSize: '13px' }}
                        >
                            {t.messages.all}
                            {pagination.total > 0 && !filterUnread && (
                                <span style={{
                                    fontSize: '11px',
                                    padding: '1px 6px',
                                    borderRadius: '999px',
                                    background: 'var(--bg-tertiary)',
                                    color: 'var(--text-secondary)',
                                }}>
                                    {pagination.total}
                                </span>
                            )}
                        </button>
                        <button
                            type="button"
                            onClick={() => setFilterUnread(true)}
                            className={`tab-nav-btn${filterUnread ? ' active' : ''}`}
                            style={{ padding: '6px 2px', fontSize: '13px' }}
                        >
                            {t.messages.unread}
                        </button>
                    </div>
                    <button
                        type="button"
                        onClick={handleMarkAllRead}
                        disabled={markingAllRead || messages.every(msg => msg.is_read)}
                        className="btn btn-secondary"
                        style={{
                            padding: '6px 12px',
                            fontSize: '12px',
                            height: '32px',
                            cursor: messages.some(msg => !msg.is_read) ? 'pointer' : 'not-allowed',
                            opacity: messages.some(msg => !msg.is_read) ? 1 : 0.6,
                        }}
                    >
                        <CheckCheck size={13} style={{ marginRight: '4px' }} />
                        {t.messages.markAllRead}
                    </button>
                    <button
                        type="button"
                        onClick={handleRefresh}
                        className="login-toolbar-btn"
                        style={{
                            width: '32px',
                            height: '32px',
                            cursor: 'pointer',
                        }}
                    >
                        <RefreshCw size={13} />
                    </button>
                </div>
            </div>

            {error && (
                <div style={{
                    padding: '10px 12px',
                    marginBottom: '16px',
                    background: 'rgba(255, 0, 0, 0.05)',
                    border: '1px solid rgba(255, 0, 0, 0.15)',
                    borderRadius: 'var(--radius-sm)',
                    color: 'var(--danger)',
                    fontSize: '13px',
                }}>
                    {error}
                </div>
            )}

            <div className="domain-list-card" style={{ padding: '0' }}>
                {loading && (
                    <div style={{ padding: '40px 0', textAlign: 'center' }}>
                        <div className="spinner" />
                    </div>
                )}

                {!loading && messages.length === 0 && (
                    <div className="domain-list-card" style={{ padding: '40px 20px', textAlign: 'center', color: 'var(--text-tertiary)' }}>
                        {t.messages.noMessages}
                    </div>
                )}

                {!loading && messages.map((msg) => (
                    <div
                        key={msg.id}
                        style={{
                            padding: '14px 16px',
                            borderBottom: '1px solid var(--border-color)',
                            display: 'flex',
                            alignItems: 'flex-start',
                            gap: '12px',
                            background: msg.is_read ? 'transparent' : 'rgba(33, 150, 243, 0.03)',
                        }}
                    >
                        <div style={{ paddingTop: '2px' }}>
                            {!msg.is_read && (
                                <span style={{
                                    display: 'inline-block',
                                    width: '8px',
                                    height: '8px',
                                    borderRadius: '50%',
                                    background: 'var(--accent-primary)',
                                    flexShrink: 0,
                                }} />
                            )}
                        </div>

                        <div style={{ flex: 1, minWidth: 0 }}>
                            <div style={{ fontSize: '14px', fontWeight: msg.is_read ? '400' : '600', color: 'var(--text-primary)', marginBottom: '4px' }}>
                                {msg.title}
                            </div>
                            <div style={{ fontSize: '12px', color: 'var(--text-secondary)', lineHeight: '1.5', whiteSpace: 'pre-wrap' }}>
                                {msg.content}
                            </div>
                            <div style={{ marginTop: '6px', display: 'flex', gap: '8px', alignItems: 'center', flexWrap: 'wrap' }}>
                                <span style={{
                                    fontSize: '11px',
                                    color: getStatusColor(msg.days_remaining),
                                    background: 'var(--bg-secondary)',
                                    border: '1px solid var(--border-color)',
                                    borderRadius: '999px',
                                    padding: '2px 8px',
                                }}>
                                    {msg.days_remaining >= 0 ? t.messages.daysRemaining.replace('{n}', msg.days_remaining) : `${Math.abs(msg.days_remaining)} 天前过期`}
                                </span>
                                {msg.renewal_url && (
                                    <a
                                        href={msg.renewal_url}
                                        target="_blank"
                                        rel="noreferrer"
                                        style={{ fontSize: '11px', color: 'var(--accent-primary)', textDecoration: 'none', display: 'inline-flex', alignItems: 'center', gap: '2px' }}
                                    >
                                        {t.messages.renewNow}
                                        <ExternalLink size={10} />
                                    </a>
                                )}
                            </div>
                        </div>

                        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: '6px', flexShrink: 0 }}>
                            <span style={{ fontSize: '11px', color: 'var(--text-tertiary)', whiteSpace: 'nowrap' }}>
                                {formatTime(msg.created_at)}
                            </span>
                            <div style={{ display: 'flex', gap: '6px' }}>
                                {!msg.is_read && (
                                    <button
                                        type="button"
                                        onClick={() => handleMarkRead(msg.id)}
                                        style={{
                                            background: 'transparent',
                                            border: '1px solid var(--border-color)',
                                            borderRadius: 'var(--radius-sm)',
                                            color: 'var(--text-secondary)',
                                            cursor: 'pointer',
                                            padding: '4px',
                                            display: 'inline-flex',
                                            alignItems: 'center',
                                            justifyContent: 'center',
                                        }}
                                        title={t.messages.markedRead}
                                    >
                                        <Check size={13} />
                                    </button>
                                )}
                                <button
                                    type="button"
                                    onClick={() => handleDelete(msg.id)}
                                    style={{
                                        background: 'transparent',
                                        border: '1px solid var(--border-color)',
                                        borderRadius: 'var(--radius-sm)',
                                        color: 'var(--text-secondary)',
                                        cursor: 'pointer',
                                        padding: '4px',
                                        display: 'inline-flex',
                                        alignItems: 'center',
                                        justifyContent: 'center',
                                    }}
                                    title={t.common.delete}
                                >
                                    <Trash2 size={13} />
                                </button>
                            </div>
                        </div>
                    </div>
                ))}
            </div>

            {pagination.totalPages > 1 && (
                <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', gap: '8px', marginTop: '20px', flexWrap: 'wrap' }}>
                    <button
                        type="button"
                        onClick={() => handlePageChange(pagination.page - 1)}
                        disabled={pagination.page <= 1}
                        className="btn"
                        style={{
                            padding: '6px 12px',
                            fontSize: '12px',
                            height: '32px',
                            background: 'var(--bg-secondary)',
                            color: 'var(--text-primary)',
                            border: '1px solid var(--border-color)',
                            opacity: pagination.page <= 1 ? 0.5 : 1,
                            cursor: pagination.page <= 1 ? 'not-allowed' : 'pointer',
                        }}
                    >
                        {t.common.previousPage}
                    </button>
                    <span style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
                        {t.common.totalItems.replace('{count}', pagination.total)}
                    </span>
                    <button
                        type="button"
                        onClick={() => handlePageChange(pagination.page + 1)}
                        disabled={pagination.page >= pagination.totalPages}
                        className="btn"
                        style={{
                            padding: '6px 12px',
                            fontSize: '12px',
                            height: '32px',
                            background: 'var(--bg-secondary)',
                            color: 'var(--text-primary)',
                            border: '1px solid var(--border-color)',
                            opacity: pagination.page >= pagination.totalPages ? 0.5 : 1,
                            cursor: pagination.page >= pagination.totalPages ? 'not-allowed' : 'pointer',
                        }}
                    >
                        {t.common.nextPage}
                    </button>
                </div>
            )}
        </div>
    );
};

export default Messages;
