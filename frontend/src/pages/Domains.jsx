import { useState, useEffect, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api } from '../api';
import { ArrowLeft, Search, RefreshCw, Globe, ExternalLink, Calendar, Link as LinkIcon, Edit2, Clock, Server } from 'lucide-react';
import Modal from '../components/Modal';
import { useLanguage } from '../LanguageContext';

const Domains = () => {
    const { t, language } = useLanguage();
    const { accountId } = useParams();
    const [domains, setDomains] = useState([]);
    const [filteredDomains, setFilteredDomains] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [searchTerm, setSearchTerm] = useState('');
    const [cacheTimestamp, setCacheTimestamp] = useState(null);
    const [domainsToDelete, setDomainsToDelete] = useState([]);
    const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
    const [restoredDomains, setRestoredDomains] = useState([]);

    // Renewal modal state
    const [renewalModal, setRenewalModal] = useState({ open: false, domain: null });
    const [renewalForm, setRenewalForm] = useState({
        renewal_date: '',
        renewal_url: '',
        is_permanent_free: false,
        notify_days_before: 30,
        notify_enabled: false
    });
    const [renewalSubmitting, setRenewalSubmitting] = useState(false);
    const [renewalError, setRenewalError] = useState('');

    const loadDomains = useCallback(async (forceRefresh = false) => {
        setLoading(true);
        try {
            let data;
            let deletedItems = [];
            let restored = [];

            if (forceRefresh) {
                // 强制刷新时调用refresh API
                const refreshData = await api.refreshDomains(accountId);
                data = refreshData.domains || [];
                deletedItems = refreshData.domains_to_delete || [];
                restored = refreshData.restored_domains || [];
                setCacheTimestamp(refreshData.cache_timestamp || null);

                // 如果有需要删除的域名，显示确认对话框
                if (deletedItems.length > 0) {
                    setDomainsToDelete(deletedItems);
                    setShowDeleteConfirm(true);
                }

                // 如果有恢复的域名，显示提示
                if (restored.length > 0) {
                    setRestoredDomains(restored);
                    setTimeout(() => setRestoredDomains([]), 5000);
                }
            } else {
                // 正常加载从缓存
                const response = await api.getDomains(accountId);
                data = response.domains || response || [];
                setCacheTimestamp(response.cache_timestamp || null);
            }

            setDomains(data);
            setFilteredDomains(data);
            setError('');
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    }, [accountId]);

    useEffect(() => {
        loadDomains();
    }, [loadDomains]);

    const openRenewalModal = async (domain, e) => {
        if (e) e.preventDefault();
        e?.stopPropagation?.();
        
        // Load notification settings
        let notifySetting = { days_before: 30, enabled: false };
        try {
            const setting = await api.getNotificationSetting(accountId, domain.id);
            if (setting && setting.days_before) {
                notifySetting = {
                    days_before: setting.days_before,
                    enabled: setting.enabled
                };
            }
        } catch {
            // Use default if not found
        }
        
        setRenewalForm({
            renewal_date: domain.renewal_date || '',
            renewal_url: domain.renewal_url || '',
            is_permanent_free: domain.renewal_date === 'permanent',
            notify_days_before: notifySetting.days_before,
            notify_enabled: notifySetting.enabled
        });
        setRenewalError('');
        setRenewalModal({ open: true, domain });
    };

    const handleRenewalSubmit = async (e) => {
        e.preventDefault();
        if (!renewalModal.domain) return;
        setRenewalSubmitting(true);
        setRenewalError('');
        try {
            const payload = {
                renewal_date: renewalForm.is_permanent_free ? 'permanent' : renewalForm.renewal_date,
                renewal_url: renewalForm.renewal_url,
                notify_days_before: renewalForm.notify_enabled ? renewalForm.notify_days_before : 0,
                notify_enabled: renewalForm.notify_enabled
            };
            
            await api.updateDomainCache(
                accountId,
                renewalModal.domain.id,
                payload
            );
            
            // 更新本地状态
            const updateDomain = (d) => {
                if (d.id === renewalModal.domain.id) {
                    return {
                        ...d,
                        renewal_date: payload.renewal_date,
                        renewal_url: payload.renewal_url
                    };
                }
                return d;
            };
            
            setDomains(domains.map(updateDomain));
            setFilteredDomains(filteredDomains.map(updateDomain));
            
            setRenewalModal({ open: false, domain: null });
        } catch (err) {
            setRenewalError(err.message);
        } finally {
            setRenewalSubmitting(false);
        }
    };

    const closeRenewalModal = () => {
        setRenewalModal({ open: false, domain: null });
        setRenewalError('');
    };

    const handleConfirmDelete = async () => {
        try {
            await api.batchSoftDeleteDomains(domainsToDelete);
            setShowDeleteConfirm(false);
            setDomainsToDelete([]);
            // 重新加载域名列表
            await loadDomains(false);
        } catch (err) {
            setError(err.message);
        }
    };

    const handleCancelDelete = () => {
        setShowDeleteConfirm(false);
        setDomainsToDelete([]);
    };

    // Calculate days until expiry
    const getDaysUntilExpiry = (renewalDate) => {
        if (!renewalDate || renewalDate === 'permanent') return null;
        const today = new Date();
        today.setHours(0, 0, 0, 0);
        const expiry = new Date(renewalDate);
        expiry.setHours(0, 0, 0, 0);
        const diffTime = expiry - today;
        const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
        return diffDays;
    };

    // Get expiry display info
    const getExpiryInfo = (renewalDate) => {
        const days = getDaysUntilExpiry(renewalDate);
        if (days === null) return null;
        
        let text = '';
        let color = 'var(--text-tertiary)';
        let bgColor = 'transparent';
        let fontWeight = 'normal';
        
        if (days < 0) {
            text = t.expiry.expiredDaysAgo.replace('{days}', Math.abs(days));
            color = '#fff';
            bgColor = 'var(--danger)';
            fontWeight = '600';
        } else if (days === 0) {
            text = t.expiry.expiresToday;
            color = '#fff';
            bgColor = 'var(--danger)';
            fontWeight = '600';
        } else if (days === 1) {
            text = t.expiry.expiresTomorrow;
            color = '#fff';
            bgColor = '#ff6b35';
            fontWeight = '600';
        } else if (days <= 7) {
            text = t.expiry.expiresInDays.replace('{days}', days);
            color = '#fff';
            bgColor = '#ff8c42';
            fontWeight = '600';
        } else if (days <= 30) {
            text = t.expiry.expiresInDays.replace('{days}', days);
            color = '#fff';
            bgColor = '#ffa726';
            fontWeight = '500';
        } else {
            text = t.expiry.expiresInDays.replace('{days}', days);
            color = 'var(--text-tertiary)';
        }
        
        return { text, color, bgColor, fontWeight, days };
    };

    return (
        <div>
            <div style={{ marginBottom: '1.5rem' }}>
                <Link to="/accounts" style={{ display: 'inline-flex', alignItems: 'center', gap: '0.5rem', color: 'var(--text-secondary)', fontSize: '0.875rem', marginBottom: '0.75rem', transition: 'color 0.15s' }} className="hover-text-primary">
                    <ArrowLeft size={14} />
                    {t.domains.backToAccounts}
                </Link>
                <div className="page-title-row" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: '1rem' }}>
                    <div style={{ minWidth: 0, flex: '1 1 0' }}>
                        <h2 style={{ fontSize: '1.5rem', fontWeight: 'bold', marginBottom: '0.5rem', letterSpacing: '-0.02em' }}>{t.domains.title}</h2>
                        {cacheTimestamp && (
                            <div className="badge badge-neutral" style={{ display: 'inline-flex', alignItems: 'center', gap: '0.25rem', fontFamily: 'monospace' }}>
                                <Clock size={11} />
                                <span>{t.common.cacheTime}: {new Date(cacheTimestamp).toLocaleString(language === 'en' ? 'en-US' : 'zh-CN')}</span>
                            </div>
                        )}
                    </div>
                    <div className="page-actions-bar domains-toolbar" style={{ display: 'flex', gap: '0.75rem', flexShrink: 0 }}>
                        <div style={{ position: 'relative' }} className="search-box-fixed">
                            <Search size={14} style={{ position: 'absolute', left: '0.75rem', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-tertiary)', pointerEvents: 'none' }} />
                            <input
                                type="text"
                                className="form-input domains-search-input"
                                placeholder={t.domains.searchPlaceholder}
                                style={{ paddingLeft: '2.25rem', width: '220px', height: '34px' }}
                                value={searchTerm}
                                onChange={e => setSearchTerm(e.target.value)}
                            />
                        </div>
                        <button onClick={() => loadDomains(true)} className="btn btn-secondary" style={{ height: '34px', padding: '0 10px' }} title={t.common.refresh}>
                            <RefreshCw size={15} className={loading ? "spin" : ""} style={{ animation: loading ? 'spin 1s linear infinite' : 'none' }} />
                        </button>
                    </div>
                </div>
            </div>

            {error && <div style={{ color: 'var(--danger)', marginBottom: '1rem', padding: '0.75rem 1rem', backgroundColor: 'rgba(255, 0, 0, 0.05)', border: '1px solid rgba(255, 0, 0, 0.15)', borderRadius: 'var(--radius-sm)', fontSize: '14px' }}>{t.common.error}: {error}</div>}

            {restoredDomains.length > 0 && (
                <div style={{
                    color: 'var(--success)',
                    marginBottom: '1rem',
                    padding: '0.75rem 1rem',
                    backgroundColor: 'rgba(0, 224, 84, 0.05)',
                    borderRadius: 'var(--radius-sm)',
                    border: '1px solid rgba(0, 224, 84, 0.15)',
                    fontSize: '14px'
                }}>
                    <div style={{ fontWeight: '600', marginBottom: '0.25rem' }}>✓ {t.domains.restoredDomainsTitle}</div>
                    <div className="font-mono">
                        {restoredDomains.join(', ')}
                    </div>
                </div>
            )}

            {loading && !domains.length ? (
                <div style={{ display: 'flex', justifyContent: 'center', padding: '4rem 0' }}>
                    <div className="spinner"></div>
                </div>
            ) : (
                <div className="domains-list" style={{ display: 'grid', gap: '0.75rem' }}>
                    {filteredDomains.map(domain => (
                        <div key={domain.id} className="domain-list-card">
                            <div className="domain-card-layout" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: '0.75rem', flexWrap: 'wrap' }}>
                                <div className="domain-card-main" style={{ display: 'flex', gap: '0.75rem', alignItems: 'center', flex: 1, minWidth: 0 }}>
                                    <Globe size={18} style={{ color: 'var(--text-secondary)', flexShrink: 0 }} />
                                    <div style={{ minWidth: 0 }}>
                                        <h3 className="font-mono" style={{ fontSize: '15px', fontWeight: '600', margin: 0, marginBottom: '0.25rem', color: 'var(--text-primary)', wordBreak: 'break-all' }}>{domain.name}</h3>
                                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', flexWrap: 'wrap' }}>
                                            {domain.renewal_date === 'permanent' ? (
                                                <span className="badge badge-neutral" style={{ gap: '0.25rem' }}>
                                                    <Calendar size={11} />
                                                    <span>{t.allDomains.permanentFree}</span>
                                                </span>
                                            ) : domain.renewal_date && (() => {
                                                const expiryInfo = getExpiryInfo(domain.renewal_date);
                                                return (
                                                    <span className={`badge ${expiryInfo?.bgColor !== 'transparent' ? 'badge-warning' : 'badge-neutral'}`} style={{ gap: '0.25rem' }}>
                                                        <Calendar size={11} />
                                                        <span>{expiryInfo?.text || domain.renewal_date}</span>
                                                    </span>
                                                );
                                            })()}
                                            {domain.renewal_url && (
                                                <>
                                                    {(domain.renewal_date || domain.renewal_date === 'permanent') && (
                                                        <span style={{ color: 'var(--text-tertiary)', fontSize: '0.75rem' }}>•</span>
                                                    )}
                                                    <span style={{ display: 'inline-flex', alignItems: 'center', gap: '0.25rem', fontSize: '0.75rem', maxWidth: '200px', minWidth: 0 }}>
                                                        <LinkIcon size={11} style={{ color: 'var(--accent-primary)', flexShrink: 0 }} />
                                                        <a
                                                            href={domain.renewal_url}
                                                            target="_blank"
                                                            rel="noopener noreferrer"
                                                            onClick={(e) => e.stopPropagation()}
                                                            style={{ color: 'var(--accent-primary)', textDecoration: 'none', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', display: 'block' }}
                                                            title={domain.renewal_url}
                                                        >
                                                            {domain.renewal_url}
                                                        </a>
                                                    </span>
                                                </>
                                            )}
                                        </div>
                                    </div>
                                </div>
                                <div className="domain-card-actions" style={{ display: 'flex', gap: '0.5rem', flexShrink: 0 }} onClick={(e) => e.stopPropagation()}>
                                    <button
                                        onClick={(e) => openRenewalModal(domain, e)}
                                        className="btn btn-secondary"
                                        title={t.allDomains.renewalModalTitle}
                                        style={{ fontSize: '13px', padding: '0 10px', height: '32px' }}
                                    >
                                        <Edit2 size={13} />
                                        {t.common.edit}
                                    </button>
                                    <Link to={`/accounts/${accountId}/domains/${domain.id}/records`} className="btn btn-secondary" style={{ fontSize: '13px', padding: '0 10px', height: '32px' }}>
                                        {t.domains.manageRecords}
                                        <ExternalLink size={13} />
                                    </Link>
                                </div>
                            </div>
                        </div>
                    ))}

                    {filteredDomains.length === 0 && (
                        <div style={{ textAlign: 'center', padding: '4rem', color: 'var(--text-secondary)', border: '1px dashed var(--border-color)', borderRadius: 'var(--radius-md)' }}>
                            {searchTerm ? t.common.noSearchResults : t.domains.noDomains}
                        </div>
                    )}
                </div>
            )}

            {/* Renewal Info Modal */}
            <Modal
                isOpen={renewalModal.open}
                onClose={closeRenewalModal}
                title={t.allDomains.renewalModalTitle}
            >
                <form onSubmit={handleRenewalSubmit}>
                    {renewalError && (
                        <div style={{ backgroundColor: 'rgba(239, 68, 68, 0.1)', color: 'var(--danger)', padding: '0.75rem', borderRadius: 'var(--radius-md)', marginBottom: '1rem', fontSize: '0.875rem' }}>
                            {renewalError}
                        </div>
                    )}
                    <div style={{ marginBottom: '1rem' }}>
                        <div style={{ fontSize: '0.875rem', fontWeight: '500', marginBottom: '0.5rem' }}>
                            {renewalModal.domain?.name}
                        </div>
                    </div>
                    <div className="form-group">
                        <label className="form-label" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                            <input
                                type="checkbox"
                                checked={renewalForm.is_permanent_free}
                                onChange={e => setRenewalForm({
                                    ...renewalForm,
                                    is_permanent_free: e.target.checked,
                                    renewal_date: e.target.checked ? 'permanent' : ''
                                })}
                                style={{ width: 'auto' }}
                            />
                            {t.allDomains.permanentFree}
                        </label>
                        {!renewalForm.is_permanent_free && (
                            <input
                                type="date"
                                className="form-input"
                                value={renewalForm.renewal_date}
                                onChange={e => setRenewalForm({ ...renewalForm, renewal_date: e.target.value })}
                            />
                        )}
                    </div>
                    <div className="form-group">
                        <label className="form-label">{t.allDomains.renewalUrl}</label>
                        <input
                            type="url"
                            className="form-input"
                            placeholder="https://..."
                            value={renewalForm.renewal_url}
                            onChange={e => setRenewalForm({ ...renewalForm, renewal_url: e.target.value })}
                        />
                    </div>

                    <div style={{ borderTop: '1px solid var(--border-color)', paddingTop: '1rem', marginTop: '1rem' }}>
                        <div className="form-group">
                            <label className="form-label" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                <input
                                    type="checkbox"
                                    checked={renewalForm.notify_enabled}
                                    onChange={e => setRenewalForm({ ...renewalForm, notify_enabled: e.target.checked })}
                                    style={{ width: 'auto' }}
                                />
                                {t.expiry.notifyEnabled}
                            </label>
                        </div>
                        {renewalForm.notify_enabled && (
                            <div className="form-group">
                                <label className="form-label">{t.expiry.notifyDaysBefore}</label>
                                <input
                                    type="number"
                                    className="form-input"
                                    min="1"
                                    max="365"
                                    value={renewalForm.notify_days_before}
                                    onChange={e => setRenewalForm({ ...renewalForm, notify_days_before: parseInt(e.target.value) || 30 })}
                                    placeholder="30"
                                />
                                <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)', marginTop: '0.25rem' }}>
                                    {t.expiry.notifyHint.replace('{days}', renewalForm.notify_days_before)}
                                </div>
                            </div>
                        )}
                    </div>

                    <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '2rem' }}>
                        <button type="button" onClick={closeRenewalModal} className="btn btn-ghost">{t.common.cancel}</button>
                        <button type="submit" className="btn btn-primary" disabled={renewalSubmitting}>
                            {renewalSubmitting ? <div className="spinner" style={{ width: '1rem', height: '1rem', borderWidth: '2px' }}></div> : t.common.save}
                        </button>
                    </div>
                </form>
            </Modal>

            {/* Delete Confirmation Modal */}
            <Modal
                isOpen={showDeleteConfirm}
                onClose={handleCancelDelete}
                title={t.domains.confirmDeleteTitle}
            >
                <div style={{ marginBottom: '1.5rem' }}>
                    <p style={{ marginBottom: '1rem', color: 'var(--text-secondary)' }}>
                        {t.domains.confirmDeleteMessage}
                    </p>
                    <div style={{
                        maxHeight: '200px',
                        overflowY: 'auto',
                        padding: '1rem',
                        backgroundColor: 'var(--bg-secondary)',
                        borderRadius: 'var(--radius-md)',
                        border: '1px solid var(--border-color)'
                    }}>
                        {domainsToDelete.map((item, index) => (
                            <div key={index} style={{
                                padding: '0.75rem',
                                marginBottom: '0.5rem',
                                backgroundColor: 'var(--bg-primary)',
                                borderRadius: 'var(--radius-sm)',
                                fontSize: '0.875rem',
                                display: 'flex',
                                justifyContent: 'space-between',
                                alignItems: 'center',
                                gap: '1rem'
                            }}>
                                <div style={{ flex: 1, minWidth: 0 }}>
                                    <div style={{
                                        fontWeight: '500',
                                        color: 'var(--text-primary)',
                                        marginBottom: '0.25rem',
                                        overflow: 'hidden',
                                        textOverflow: 'ellipsis',
                                        whiteSpace: 'nowrap'
                                    }}>
                                        {item.domain_name || item.domain_id}
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>
                    <p style={{ marginTop: '1rem', fontSize: '0.875rem', color: 'var(--text-tertiary)' }}>
                        {t.domains.softDeleteNote}
                    </p>
                </div>
                <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem' }}>
                    <button onClick={handleCancelDelete} className="btn btn-ghost">{t.common.cancel}</button>
                    <button onClick={handleConfirmDelete} className="btn btn-primary">{t.common.confirmDelete}</button>
                </div>
            </Modal>
        </div>
    );
};

export default Domains;
