import { useState, useEffect, useRef, useCallback } from 'react';
import { api } from '../api';
import Modal from './Modal';
import { RefreshCw, ExternalLink, Check, CheckCheck, Trash2 } from 'lucide-react';
import { useLanguage } from '../LanguageContext';

const PAGE_SIZE = 10;

const MessagesModal = ({ isOpen, onClose, onUnreadCountChange }) => {
    const { t, language } = useLanguage();
    const [messages, setMessages] = useState([]);
    const [pagination, setPagination] = useState({ total: 0, page: 1, totalPages: 1 });
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [filterUnread, setFilterUnread] = useState(false);
    const [markingAllRead, setMarkingAllRead] = useState(false);
    const contentRef = useRef(null);

    const refreshUnreadCount = useCallback(async () => {
        try {
            const data = await api.getUnreadCount();
            if (onUnreadCountChange) {
                onUnreadCountChange(data.count || 0);
            }
        } catch {
            // ignore
        }
    }, [onUnreadCountChange]);

    const fetchMessages = useCallback(async (page = 1, unreadOnly = false) => {
        setLoading(true);
        setError('');
        try {
            const data = await api.getMessages(page, PAGE_SIZE, unreadOnly);
            setMessages(data.messages || []);
            setPagination({
                total: data.total || 0,
                page: data.page || page,
                totalPages: Math.ceil((data.total || 0) / PAGE_SIZE) || 1,
            });
            refreshUnreadCount();
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    }, [refreshUnreadCount]);

    useEffect(() => {
        if (isOpen) {
            fetchMessages(1, filterUnread);
        }
    }, [isOpen, filterUnread, fetchMessages]);

    const handlePageChange = (page) => {
        fetchMessages(page, filterUnread);
        if (contentRef.current) {
            contentRef.current.scrollTop = 0;
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
            refreshUnreadCount();
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
            refreshUnreadCount();
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
            setPagination(prev => ({
                ...prev,
                total: Math.max(0, prev.total - 1),
                totalPages: Math.ceil(Math.max(0, prev.total - 1) / PAGE_SIZE) || 1,
            }));
            refreshUnreadCount();
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
        const isEn = language === 'en';

        if (diff < 60) return isEn ? `${diff}s ago` : `${diff}秒前`;
        if (diff < 3600) return isEn ? `${Math.floor(diff / 60)}m ago` : `${Math.floor(diff / 60)}分钟前`;
        if (diff < 86400) return isEn ? `${Math.floor(diff / 3600)}h ago` : `${Math.floor(diff / 3600)}小时前`;
        return date.toLocaleDateString();
    };

    const hasUnread = messages.some(msg => !msg.is_read);

    return (
        <Modal
            isOpen={isOpen}
            onClose={onClose}
            title={t.messages.title}
            size="large"
            closeOnBackdrop={true}
        >
            <div ref={contentRef} style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                {/* Header Action Bar */}
                <div style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    flexWrap: 'wrap',
                    gap: '8px',
                    paddingBottom: '8px',
                    borderBottom: '1px solid var(--border-color)',
                }}>
                    <div style={{ display: 'flex', gap: '6px' }}>
                        <button
                            type="button"
                            onClick={() => setFilterUnread(false)}
                            className="btn"
                            style={{
                                padding: '4px 10px',
                                fontSize: '12px',
                                height: '28px',
                                background: !filterUnread ? 'var(--primary)' : 'var(--bg-secondary)',
                                color: !filterUnread ? '#fff' : 'var(--text-primary)',
                                border: `1px solid ${!filterUnread ? 'var(--primary)' : 'var(--border-color)'}`,
                            }}
                        >
                            {t.messages.all}
                        </button>
                        <button
                            type="button"
                            onClick={() => setFilterUnread(true)}
                            className="btn"
                            style={{
                                padding: '4px 10px',
                                fontSize: '12px',
                                height: '28px',
                                background: filterUnread ? 'var(--primary)' : 'var(--bg-secondary)',
                                color: filterUnread ? '#fff' : 'var(--text-primary)',
                                border: `1px solid ${filterUnread ? 'var(--primary)' : 'var(--border-color)'}`,
                            }}
                        >
                            {t.messages.unread}
                        </button>
                    </div>

                    <div style={{ display: 'flex', gap: '6px' }}>
                        <button
                            type="button"
                            onClick={handleMarkAllRead}
                            disabled={markingAllRead || !hasUnread}
                            className="btn"
                            style={{
                                padding: '4px 10px',
                                fontSize: '12px',
                                height: '28px',
                                background: 'var(--bg-secondary)',
                                color: hasUnread ? 'var(--text-primary)' : 'var(--text-tertiary)',
                                border: '1px solid var(--border-color)',
                                cursor: hasUnread ? 'pointer' : 'not-allowed',
                                opacity: hasUnread ? 1 : 0.6,
                                display: 'inline-flex',
                                alignItems: 'center',
                                gap: '4px',
                            }}
                        >
                            <CheckCheck size={13} />
                            {t.messages.markAllRead}
                        </button>
                        <button
                            type="button"
                            onClick={handleRefresh}
                            className="btn"
                            style={{
                                padding: '4px 8px',
                                fontSize: '12px',
                                height: '28px',
                                background: 'var(--bg-secondary)',
                                color: 'var(--text-primary)',
                                border: '1px solid var(--border-color)',
                                display: 'inline-flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                            }}
                            title={t.common.refresh}
                        >
                            <RefreshCw size={13} />
                        </button>
                    </div>
                </div>

                {error && (
                    <div style={{
                        padding: '8px 10px',
                        background: 'rgba(255, 0, 0, 0.05)',
                        border: '1px solid rgba(255, 0, 0, 0.15)',
                        borderRadius: 'var(--radius-sm)',
                        color: 'var(--danger)',
                        fontSize: '12px',
                    }}>
                        {error}
                    </div>
                )}

                {/* Content */}
                <div style={{ minHeight: '180px', maxHeight: '55vh', overflowY: 'auto' }}>
                    {loading && (
                        <div style={{ padding: '36px 0', textAlign: 'center' }}>
                            <div className="spinner" />
                        </div>
                    )}

                    {!loading && messages.length === 0 && (
                        <div style={{
                            padding: '40px 16px',
                            textAlign: 'center',
                            color: 'var(--text-tertiary)',
                            fontSize: '13px',
                        }}>
                            {t.messages.noMessages}
                        </div>
                    )}

                    {!loading && messages.map((msg) => (
                        <div
                            key={msg.id}
                            style={{
                                padding: '12px 14px',
                                borderBottom: '1px solid var(--border-color)',
                                display: 'flex',
                                alignItems: 'flex-start',
                                gap: '10px',
                                background: msg.is_read ? 'transparent' : 'rgba(33, 150, 243, 0.04)',
                                borderRadius: 'var(--radius-sm)',
                                marginBottom: '6px',
                                transition: 'var(--transition)',
                            }}
                        >
                            {/* Unread indicator */}
                            <div style={{ paddingTop: '5px', width: '8px', flexShrink: 0 }}>
                                {!msg.is_read && (
                                    <span style={{
                                        display: 'block',
                                        width: '7px',
                                        height: '7px',
                                        borderRadius: '50%',
                                        background: 'var(--primary)',
                                    }} />
                                )}
                            </div>

                            {/* Main message text */}
                            <div style={{ flex: 1, minWidth: 0 }}>
                                <div style={{
                                    fontSize: '13px',
                                    fontWeight: msg.is_read ? '500' : '600',
                                    color: 'var(--text-primary)',
                                    marginBottom: '3px',
                                }}>
                                    {msg.title}
                                </div>
                                <div style={{
                                    fontSize: '12px',
                                    color: 'var(--text-secondary)',
                                    lineHeight: '1.5',
                                    whiteSpace: 'pre-wrap',
                                }}>
                                    {msg.content}
                                </div>
                                <div style={{
                                    marginTop: '6px',
                                    display: 'flex',
                                    gap: '8px',
                                    alignItems: 'center',
                                    flexWrap: 'wrap',
                                }}>
                                    <span style={{
                                        fontSize: '11px',
                                        color: getStatusColor(msg.days_remaining),
                                        background: 'var(--bg-secondary)',
                                        border: '1px solid var(--border-color)',
                                        borderRadius: '999px',
                                        padding: '1px 7px',
                                    }}>
                                        {msg.days_remaining >= 0
                                            ? t.messages.daysRemaining.replace('{n}', msg.days_remaining)
                                            : (language === 'en' ? `${Math.abs(msg.days_remaining)} days ago` : `${Math.abs(msg.days_remaining)} 天前过期`)}
                                    </span>
                                    {msg.renewal_url && (
                                        <a
                                            href={msg.renewal_url}
                                            target="_blank"
                                            rel="noreferrer"
                                            style={{
                                                fontSize: '11px',
                                                color: 'var(--primary)',
                                                textDecoration: 'none',
                                                display: 'inline-flex',
                                                alignItems: 'center',
                                                gap: '2px',
                                            }}
                                        >
                                            {t.messages.renewNow}
                                            <ExternalLink size={10} />
                                        </a>
                                    )}
                                </div>
                            </div>

                            {/* Actions & Timestamp */}
                            <div style={{
                                display: 'flex',
                                flexDirection: 'column',
                                alignItems: 'flex-end',
                                gap: '6px',
                                flexShrink: 0,
                            }}>
                                <span style={{ fontSize: '11px', color: 'var(--text-tertiary)', whiteSpace: 'nowrap' }}>
                                    {formatTime(msg.created_at)}
                                </span>
                                <div style={{ display: 'flex', gap: '4px' }}>
                                    {!msg.is_read && (
                                        <button
                                            type="button"
                                            onClick={() => handleMarkRead(msg.id)}
                                            style={{
                                                background: 'var(--bg-secondary)',
                                                border: '1px solid var(--border-color)',
                                                borderRadius: 'var(--radius-sm)',
                                                color: 'var(--text-secondary)',
                                                cursor: 'pointer',
                                                padding: '3px 5px',
                                                display: 'inline-flex',
                                                alignItems: 'center',
                                                justifyContent: 'center',
                                            }}
                                            title={t.messages.markedRead}
                                        >
                                            <Check size={12} />
                                        </button>
                                    )}
                                    <button
                                        type="button"
                                        onClick={() => handleDelete(msg.id)}
                                        style={{
                                            background: 'var(--bg-secondary)',
                                            border: '1px solid var(--border-color)',
                                            borderRadius: 'var(--radius-sm)',
                                            color: 'var(--text-secondary)',
                                            cursor: 'pointer',
                                            padding: '3px 5px',
                                            display: 'inline-flex',
                                            alignItems: 'center',
                                            justifyContent: 'center',
                                        }}
                                        title={t.common.delete}
                                    >
                                        <Trash2 size={12} />
                                    </button>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>

                {/* Pagination */}
                {pagination.totalPages > 1 && (
                    <div style={{
                        display: 'flex',
                        justifyContent: 'center',
                        alignItems: 'center',
                        gap: '8px',
                        paddingTop: '8px',
                        borderTop: '1px solid var(--border-color)',
                        flexWrap: 'wrap',
                    }}>
                        <button
                            type="button"
                            onClick={() => handlePageChange(pagination.page - 1)}
                            disabled={pagination.page <= 1}
                            className="btn"
                            style={{
                                padding: '4px 10px',
                                fontSize: '12px',
                                height: '28px',
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
                                padding: '4px 10px',
                                fontSize: '12px',
                                height: '28px',
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
        </Modal>
    );
};

export default MessagesModal;
