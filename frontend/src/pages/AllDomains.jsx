import { useState, useEffect, useCallback, useRef } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api';
import { Search, RefreshCw, Globe, ExternalLink, Server, Calendar, Link as LinkIcon, Edit2 } from 'lucide-react';
import Modal from '../components/Modal';
import { useLanguage } from '../LanguageContext';

const AllDomains = () => {
    const { t, language } = useLanguage();
    const [domains, setDomains] = useState([]);
    const [filteredDomains, setFilteredDomains] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [searchTerm, setSearchTerm] = useState('');
    const fetchedRef = useRef(false);

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
            // 根据是否强制刷新选择不同的 API
            const data = forceRefresh 
                ? await api.refreshAllDomains() 
                : await api.getAllDomains();
            setDomains(data || []);
            setFilteredDomains(data || []);
            setError('');
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        // Prevent duplicate requests in React StrictMode
        if (fetchedRef.current) return;
        fetchedRef.current = true;
        loadDomains();
    }, [loadDomains]);

    useEffect(() => {
        if (!searchTerm) {
            setFilteredDomains(domains);
        } else {
            const term = searchTerm.toLowerCase();
            setFilteredDomains(
                domains.filter(d =>
                    d.name.toLowerCase().includes(term) ||
                    (d.account_name && d.account_name.toLowerCase().includes(term)) ||
                    (d.ipv4_address && d.ipv4_address.toLowerCase().includes(term))
                )
            );
        }
    }, [searchTerm, domains]);

    const openRenewalModal = async (domain, e) => {
        if (e) e.preventDefault();
        e?.stopPropagation?.();
        
        // Load notification settings
        let notifySetting = { days_before: 30, enabled: false };
        try {
            const setting = await api.getNotificationSetting(domain.account_id, domain.id);
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
                renewalModal.domain.account_id,
                renewalModal.domain.id,
                payload
            );
            
            // 更新本地状态
            const updateDomain = (d) => {
                if (d.account_id === renewalModal.domain.account_id && d.id === renewalModal.domain.id) {
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
            <div style={{ marginBottom: '2rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
                    <div>
                        <h2 style={{ fontSize: '1.5rem', fontWeight: 'bold' }}>{t.allDomains.title}</h2>
                        <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', marginTop: '0.25rem' }}>
                            {t.allDomains.subtitle}
                            {!loading && domains.length > 0 && (
                                <span style={{ marginLeft: '0.5rem' }}>
                                    • {t.allDomains.totalDomains || 'Total'}: {domains.length}
                                </span>
                            )}
                        </p>
                    </div>
                    <div style={{ display: 'flex', gap: '0.75rem' }}>
                        <div style={{ position: 'relative' }}>
                            <Search size={16} style={{ position: 'absolute', left: '0.75rem', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-tertiary)' }} />
                            <input
                                type="text"
                                className="form-input"
                                placeholder={t.allDomains.searchPlaceholder}
                                style={{ paddingLeft: '2.5rem', width: '250px' }}
                                value={searchTerm}
                                onChange={e => setSearchTerm(e.target.value)}
                            />
                        </div>
                        <button onClick={() => loadDomains(true)} className="btn btn-secondary" title={t.common.refresh}>
                            <RefreshCw size={18} className={loading ? "spin" : ""} style={{ animation: loading ? 'spin 1s linear infinite' : 'none' }} />
                        </button>
                    </div>
                </div>
            </div>

            {error && <div style={{ color: 'var(--danger)', marginBottom: '1rem', padding: '1rem', backgroundColor: 'rgba(239, 68, 68, 0.1)', borderRadius: 'var(--radius-md)' }}>{t.common.error}: {error}</div>}

            {loading && !domains.length ? (
                <div style={{ textAlign: 'center', padding: '4rem' }}>
                    <div className="spinner" style={{ margin: '0 auto 1rem' }}></div>
                    <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
                        {t.allDomains.loadingFromAccounts}
                    </p>
                </div>
            ) : (
                <div style={{ display: 'grid', gap: '1rem' }}>
                    {filteredDomains.map(domain => (
                        <div key={`${domain.account_id}-${domain.id}`} className="glass-panel" style={{ 
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
                        }}>
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: '1rem' }}>
                                <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center', flex: 1, minWidth: 0 }}>
                                    <Globe size={20} style={{ color: 'var(--accent-primary)', flexShrink: 0 }} />
                                    <div style={{ flex: 1, minWidth: 0 }}>
                                        <h3 style={{ fontSize: '1rem', fontWeight: '500', margin: 0, marginBottom: '0.25rem' }}>{domain.name}</h3>
                                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', flexWrap: 'wrap' }}>
                                            <span style={{ display: 'inline-flex', alignItems: 'center', gap: '0.25rem', color: 'var(--text-tertiary)', fontSize: '0.75rem' }}>
                                                <Server size={12} />
                                                {domain.account_name}
                                            </span>
                                            {domain.updated_on && (
                                                <>
                                                    <span style={{ color: 'var(--text-tertiary)', fontSize: '0.75rem' }}>•</span>
                                                    <span style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)' }}>
                                                        {new Date(domain.updated_on).toLocaleString(language === 'en' ? 'en-US' : 'zh-CN', { 
                                                            year: 'numeric', 
                                                            month: '2-digit', 
                                                            day: '2-digit',
                                                            hour: '2-digit',
                                                            minute: '2-digit'
                                                        })}
                                                    </span>
                                                </>
                                            )}
                                            {domain.renewal_date === 'permanent' ? (
                                                <>
                                                    <span style={{ color: 'var(--text-tertiary)', fontSize: '0.75rem' }}>•</span>
                                                    <span style={{ 
                                                        display: 'inline-flex', 
                                                        alignItems: 'center', 
                                                        gap: '0.25rem', 
                                                        fontSize: '0.75rem', 
                                                        color: 'var(--text-tertiary)',
                                                        fontWeight: 'normal'
                                                    }}>
                                                        <Calendar size={12} />
                                                        {t.allDomains.permanentFree || '永久免费'}
                                                    </span>
                                                </>
                                            ) : domain.renewal_date && (() => {
                                                const expiryInfo = getExpiryInfo(domain.renewal_date);
                                                return (
                                                    <>
                                                        <span style={{ color: 'var(--text-tertiary)', fontSize: '0.75rem' }}>•</span>
                                                        <span style={{ 
                                                            display: 'inline-flex', 
                                                            alignItems: 'center', 
                                                            gap: '0.25rem', 
                                                            fontSize: '0.75rem',
                                                            color: expiryInfo?.color || 'var(--text-tertiary)',
                                                            backgroundColor: expiryInfo?.bgColor || 'transparent',
                                                            padding: expiryInfo?.bgColor !== 'transparent' ? '0.15rem 0.5rem' : '0',
                                                            borderRadius: 'var(--radius-sm)',
                                                            fontWeight: expiryInfo?.fontWeight || 'normal'
                                                        }}>
                                                            <Calendar size={12} />
                                                            {expiryInfo?.text || domain.renewal_date}
                                                        </span>
                                                    </>
                                                );
                                            })()}
                                            {domain.renewal_url && (
                                                <>
                                                    <span style={{ color: 'var(--text-tertiary)', fontSize: '0.75rem' }}>•</span>
                                                    <span style={{ display: 'inline-flex', alignItems: 'center', gap: '0.25rem', fontSize: '0.75rem', maxWidth: '200px' }}>
                                                        <LinkIcon size={12} style={{ color: 'var(--accent-primary)', flexShrink: 0 }} />
                                                        <a
                                                            href={domain.renewal_url}
                                                            target="_blank"
                                                            rel="noopener noreferrer"
                                                            onClick={(e) => e.stopPropagation()}
                                                            style={{ color: 'var(--accent-primary)', textDecoration: 'none', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}
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
                                <div style={{ display: 'flex', gap: '0.5rem', flexShrink: 0 }}>
                                    <button
                                        onClick={(e) => openRenewalModal(domain, e)}
                                        className="btn btn-secondary"
                                        title={t.allDomains.editRenewal || 'Edit Renewal Info'}
                                        style={{ fontSize: '0.875rem', padding: '0.5rem 1rem' }}
                                    >
                                        <Edit2 size={14} />
                                        {t.common.edit}
                                    </button>
                                    <Link
                                        to={`/accounts/${domain.account_id}/domains/${domain.id}/records`}
                                        state={{ from: '/domains' }}
                                        className="btn btn-secondary"
                                        style={{ fontSize: '0.875rem', padding: '0.5rem 1rem', flexShrink: 0 }}
                                    >
                                        {t.domains.manageRecords}
                                        <ExternalLink size={14} />
                                    </Link>
                                </div>
                            </div>
                        </div>
                    ))}

                    {filteredDomains.length === 0 && (
                        <div style={{ textAlign: 'center', padding: '4rem', color: 'var(--text-secondary)' }}>
                            {searchTerm ? t.allDomains.noSearchResults : t.allDomains.noDomains}
                        </div>
                    )}
                </div>
            )}

            {/* Renewal Info Modal */}
            <Modal
                isOpen={renewalModal.open}
                onClose={closeRenewalModal}
                title={t.allDomains.renewalModalTitle || 'Edit Renewal Info'}
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
                            {t.allDomains.permanentFree || '永久免费'}
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
                        <label className="form-label">{t.allDomains.renewalUrl || 'Renewal URL'}</label>
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
        </div>
    );
};

export default AllDomains;
