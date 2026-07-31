import { useState, useEffect, useCallback, useRef } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api';
import {
    Globe2, Search, RefreshCw, Plus, Trash2, RotateCw, ExternalLink,
    Server, Calendar, AlertCircle, CheckCircle, ToggleLeft, ToggleRight, Loader, Cloud
} from 'lucide-react';
import Modal from '../components/Modal';

const PRESET_ROOT_DOMAINS = ['cc.cd', 'de5.net', 'bot.cd', 'ccwu.cc', 'bbroot.com', 'ddns.ge'];
import ConfirmDialog from '../components/ConfirmDialog';
import { useLanguage } from '../LanguageContext';

const DNSHE = () => {
    const { t, language } = useLanguage();
    const fetchedRef = useRef(false);

    const [accounts, setAccounts] = useState([]);
    const [selectedAccountId, setSelectedAccountId] = useState('');
    const [domains, setDomains] = useState([]);
    const [filteredDomains, setFilteredDomains] = useState([]);
    const [searchTerm, setSearchTerm] = useState('');
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');

    // Quota
    const [quota, setQuota] = useState(null);
    const [quotaLoading, setQuotaLoading] = useState(false);

    // Auto-renew config
    const [autoRenew, setAutoRenew] = useState({ enabled: false, days_before: 7, last_run_at: null });
    const [autoRenewLoading, setAutoRenewLoading] = useState(false);
    const [autoRenewTriggering, setAutoRenewTriggering] = useState(false);

    // Register modal
    const [registerOpen, setRegisterOpen] = useState(false);
    const [registerForm, setRegisterForm] = useState({ subdomain: '', rootdomain: '' });
    const [registerSubmitting, setRegisterSubmitting] = useState(false);
    const [registerError, setRegisterError] = useState('');
    const [useCustomRoot, setUseCustomRoot] = useState(false);

    // Delete confirm
    const [deleteTarget, setDeleteTarget] = useState(null);
    const [deleting, setDeleting] = useState(false);

    // Resolution toggle in-flight per domain
    const [togglingResolutionId, setTogglingResolutionId] = useState(null);

    // Resolve to Cloudflare modal
    const [cfModalOpen, setCfModalOpen] = useState(false);
    const [cfTargetDomain, setCfTargetDomain] = useState(null);
    const [cfAccounts, setCfAccounts] = useState([]);
    const [cfSelectedAccountId, setCfSelectedAccountId] = useState('');
    const [cfSubmitting, setCfSubmitting] = useState(false);
    const [cfError, setCfError] = useState('');

    const selectedAccount = accounts.find(a => String(a.id) === String(selectedAccountId));

    const showSuccess = (msg) => {
        setSuccess(msg);
        setTimeout(() => setSuccess(''), 4000);
    };

    const loadAccounts = useCallback(async () => {
        setLoading(true);
        try {
            const data = await api.dnsheGetAccounts();
            const list = data.accounts || [];
            setAccounts(list);
            if (list.length > 0 && !selectedAccountId) {
                setSelectedAccountId(String(list[0].id));
            }
            setError('');
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    }, [selectedAccountId]);

    const loadAutoRenew = useCallback(async () => {
        try {
            const cfg = await api.dnsheGetAutoRenew();
            setAutoRenew({
                enabled: !!cfg.enabled,
                days_before: cfg.days_before || 7,
                last_run_at: cfg.last_run_at || null,
            });
        } catch {
            // ignore — defaults remain
        }
    }, []);

    useEffect(() => {
        if (fetchedRef.current) return;
        fetchedRef.current = true;
        loadAccounts();
        loadAutoRenew();
    }, [loadAccounts, loadAutoRenew]);

    const loadDomains = useCallback(async (forceRefresh = false) => {
        if (!selectedAccountId) {
            setDomains([]);
            setFilteredDomains([]);
            return;
        }
        setLoading(true);
        try {
            let data;
            if (forceRefresh) {
                const refreshData = await api.refreshDomains(selectedAccountId);
                data = refreshData.domains || [];
            } else {
                const response = await api.getDomains(selectedAccountId);
                data = response.domains || [];
            }
            setDomains(data);
            setFilteredDomains(data);
            setError('');
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    }, [selectedAccountId]);

    useEffect(() => {
        loadDomains();
    }, [loadDomains]);

    useEffect(() => {
        if (!searchTerm) {
            setFilteredDomains(domains);
        } else {
            const term = searchTerm.toLowerCase();
            setFilteredDomains(domains.filter(d =>
                (d.name && d.name.toLowerCase().includes(term)) ||
                (d.account_name && d.account_name.toLowerCase().includes(term))
            ));
        }
    }, [searchTerm, domains]);

    const loadQuota = useCallback(async () => {
        if (!selectedAccountId) {
            setQuota(null);
            return;
        }
        setQuotaLoading(true);
        try {
            const q = await api.dnsheGetQuota(selectedAccountId);
            setQuota(q.quota || q);
            setError('');
        } catch (err) {
            setQuota(null);
        } finally {
            setQuotaLoading(false);
        }
    }, [selectedAccountId]);

    useEffect(() => {
        loadQuota();
    }, [loadQuota]);

    // Register submit
    const handleRegister = async (e) => {
        e.preventDefault();
        if (!selectedAccountId) return;
        setRegisterSubmitting(true);
        setRegisterError('');
        try {
            const resp = await api.dnsheRegister(selectedAccountId, {
                subdomain: registerForm.subdomain.trim(),
                rootdomain: registerForm.rootdomain.trim(),
            });
            showSuccess(t.dnshe.register.success.replace('{domain}', resp.full_domain || `${registerForm.subdomain}.${registerForm.rootdomain}`));
            setRegisterOpen(false);
            setRegisterForm({ subdomain: '', rootdomain: '' });
            setUseCustomRoot(false);
            await loadDomains(true);
            loadQuota();
        } catch (err) {
            setRegisterError(err.message);
        } finally {
            setRegisterSubmitting(false);
        }
    };

    // Delete
    const handleDelete = async () => {
        if (!deleteTarget || !selectedAccountId) return;
        setDeleting(true);
        try {
            await api.dnsheDelete(selectedAccountId, { subdomain_id: parseInt(deleteTarget.id) });
            showSuccess(t.dnshe.delete.success);
            setDeleteTarget(null);
            await loadDomains(true);
            loadQuota();
        } catch (err) {
            setError(err.message);
        } finally {
            setDeleting(false);
        }
    };

    // Toggle resolution
    const handleToggleResolution = async (domain) => {
        if (!selectedAccountId) return;
        const currentUsesDNSHE = domain.uses_dnshe_dns !== false; // default true when undefined
        const target = !currentUsesDNSHE; // flip
        setTogglingResolutionId(domain.id);
        try {
            await api.dnsheSetResolution(selectedAccountId, domain.id, { uses_dnshe_dns: target });
            showSuccess(t.dnshe.resolution.updated);
            // update local state
            setDomains(prev => prev.map(d => d.id === domain.id ? { ...d, uses_dnshe_dns: target } : d));
            setFilteredDomains(prev => prev.map(d => d.id === domain.id ? { ...d, uses_dnshe_dns: target } : d));
        } catch (err) {
            setError(err.message);
        } finally {
            setTogglingResolutionId(null);
        }
    };

    // Resolve to Cloudflare
    const openCfModal = async (domain) => {
        setCfTargetDomain(domain);
        setCfError('');
        setCfSelectedAccountId('');
        try {
            const data = await api.getAccounts();
            const cfList = (Array.isArray(data) ? data : (data.accounts || [])).filter(a => a.provider_type === 'cloudflare');
            setCfAccounts(cfList);
            if (cfList.length > 0) {
                setCfSelectedAccountId(String(cfList[0].id));
            }
        } catch (err) {
            setCfError(err.message);
        }
        setCfModalOpen(true);
    };

    const handleResolveToCloudflare = async (e) => {
        e.preventDefault();
        if (!cfTargetDomain || !cfSelectedAccountId || !selectedAccountId) return;
        setCfSubmitting(true);
        setCfError('');
        try {
            const resp = await api.dnsheResolveToCloudflare(selectedAccountId, cfTargetDomain.id, {
                cloudflare_account_id: Number(cfSelectedAccountId),
            });
            const nsList = (resp.name_servers || []).join(', ');
            showSuccess(t.dnshe.resolveToCF.success
                .replace('{domain}', resp.domain_name || cfTargetDomain.name)
                .replace('{zone}', resp.zone_name || '')
                .replace('{ns}', nsList));
            setCfModalOpen(false);
            // 刷新域名列表，同步 uses_dnshe_dns 状态
            await loadDomains(true);
        } catch (err) {
            setCfError(err.message);
        } finally {
            setCfSubmitting(false);
        }
    };

    // Auto-renew config save
    const handleAutoRenewChange = async (patch) => {
        const next = { ...autoRenew, ...patch };
        setAutoRenew(next);
        setAutoRenewLoading(true);
        try {
            const payload = {};
            if (patch.enabled !== undefined) payload.enabled = patch.enabled;
            if (patch.days_before !== undefined) payload.days_before = patch.days_before;
            const cfg = await api.dnsheUpdateAutoRenew(payload);
            setAutoRenew({
                enabled: !!cfg.enabled,
                days_before: cfg.days_before || 7,
                last_run_at: cfg.last_run_at || null,
            });
            showSuccess(t.dnshe.autoRenew.saved);
        } catch (err) {
            setError(err.message);
        } finally {
            setAutoRenewLoading(false);
        }
    };

    const handleTriggerAutoRenew = async () => {
        setAutoRenewTriggering(true);
        try {
            const res = await api.dnsheTriggerAutoRenew();
            let details = '';
            const renewedDomains = (res.renewed_domains || []).join(', ');
            const failedDomains = (res.failed_domains || []).join(', ');
            if (renewedDomains || failedDomains) {
                details = t.dnshe.autoRenew.triggerSuccessDetails
                    .replace('{renewedDomains}', renewedDomains)
                    .replace('{failedDomains}', failedDomains);
            }
            showSuccess(t.dnshe.autoRenew.triggerSuccess
                .replace('{renewed}', res.renewed || 0)
                .replace('{failed}', res.failed || 0)
                .replace('{details}', details));
            loadAutoRenew();
            await loadDomains(true);
            loadQuota();
        } catch (err) {
            setError(err.message);
        } finally {
            setAutoRenewTriggering(false);
        }
    };

    const fmtDateTime = (ts) => {
        if (!ts) return t.dnshe.autoRenew.never;
        try {
            return new Date(ts).toLocaleString(language === 'en' ? 'en-US' : 'zh-CN', {
                year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit'
            });
        } catch {
            return ts;
        }
    };

    const quotaPct = quota && quota.total > 0 ? Math.min(100, Math.round((quota.used / quota.total) * 100)) : 0;

    return (
        <div>
            <div style={{ marginBottom: '1.5rem' }}>
                <div className="page-title-row" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
                    <div style={{ minWidth: 0, flex: 1 }}>
                        <h2 style={{ fontSize: '1.5rem', fontWeight: 'bold', letterSpacing: '-0.02em', margin: 0, display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                            <Globe2 size={22} />
                            {t.dnshe.title}
                        </h2>
                        <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', marginTop: '0.25rem' }}>
                            {t.dnshe.subtitle}
                        </p>
                    </div>
                    <div className="page-actions-bar" style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
                        <select
                            className="form-input"
                            style={{ height: '34px', minWidth: '180px' }}
                            value={selectedAccountId}
                            onChange={e => setSelectedAccountId(e.target.value)}
                        >
                            {accounts.length === 0 && <option value="">{t.dnshe.noAccounts}</option>}
                            {accounts.map(a => (
                                <option key={a.id} value={a.id}>{a.name}</option>
                            ))}
                        </select>
                        <button
                            onClick={() => { loadDomains(true); loadQuota(); }}
                            className="btn btn-secondary"
                            style={{ height: '34px', padding: '0 10px' }}
                            title={t.dnshe.refresh}
                            disabled={!selectedAccountId}
                        >
                            <RefreshCw size={15} className={loading ? "spin" : ""} style={{ animation: loading ? 'spin 1s linear infinite' : 'none' }} />
                        </button>
                    </div>
                </div>
            </div>

            {error && (
                <div style={{ color: 'var(--danger)', marginBottom: '1rem', padding: '0.75rem 1rem', backgroundColor: 'rgba(255, 0, 0, 0.05)', border: '1px solid rgba(255, 0, 0, 0.15)', borderRadius: 'var(--radius-sm)', fontSize: '14px' }}>
                    {t.common.error}: {error}
                </div>
            )}
            {success && (
                <div style={{ color: 'var(--success)', marginBottom: '1rem', padding: '0.75rem 1rem', backgroundColor: 'rgba(0, 224, 84, 0.05)', border: '1px solid rgba(0, 224, 84, 0.15)', borderRadius: 'var(--radius-sm)', fontSize: '14px', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                    <CheckCircle size={14} />
                    {success}
                </div>
            )}

            {accounts.length === 0 && !loading ? (
                <div style={{ textAlign: 'center', padding: '4rem', color: 'var(--text-secondary)', border: '1px dashed var(--border-color)', borderRadius: 'var(--radius-md)' }}>
                    {t.dnshe.noAccounts}
                </div>
            ) : (
                <>
                    {/* Quota + Auto-renew cards */}
                    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '1rem', marginBottom: '1.5rem' }}>
                        {/* Quota card */}
                        <div style={{ padding: '1rem 1.25rem', border: '1px solid var(--border-color)', borderRadius: 'var(--radius-md)', background: 'var(--bg-secondary)' }}>
                            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.75rem', fontWeight: 600, fontSize: '0.9rem' }}>
                                <Server size={15} />
                                {t.dnshe.quota.title}
                            </div>
                            {quotaLoading ? (
                                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--text-tertiary)', fontSize: '0.8rem' }}>
                                    <Loader size={13} className="spin" style={{ animation: 'spin 1s linear infinite' }} />
                                    {t.dnshe.domains.loading}
                                </div>
                            ) : quota ? (
                                <div>
                                    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.8rem', marginBottom: '0.4rem', color: 'var(--text-secondary)' }}>
                                        <span>{t.dnshe.quota.progress.replace('{used}', quota.used).replace('{total}', quota.total)}</span>
                                        <span style={{ color: quota.available > 0 ? 'var(--success)' : 'var(--danger)', fontWeight: 600 }}>
                                            {t.dnshe.quota.available}: {quota.available}
                                        </span>
                                    </div>
                                    <div style={{ height: '6px', background: 'var(--bg-primary)', borderRadius: '3px', overflow: 'hidden' }}>
                                        <div style={{ width: `${quotaPct}%`, height: '100%', background: quota.available > 0 ? 'var(--accent-primary)' : 'var(--danger)', transition: 'width 0.3s' }} />
                                    </div>
                                    <div style={{ display: 'flex', gap: '1rem', marginTop: '0.6rem', fontSize: '0.75rem', color: 'var(--text-tertiary)', flexWrap: 'wrap' }}>
                                        <span>{t.dnshe.quota.base}: {quota.base}</span>
                                        <span>{t.dnshe.quota.inviteBonus}: {quota.invite_bonus}</span>
                                        <span>{t.dnshe.quota.used}: {quota.used}</span>
                                    </div>
                                </div>
                            ) : (
                                <div style={{ fontSize: '0.8rem', color: 'var(--text-tertiary)' }}>—</div>
                            )}
                        </div>

                        {/* Auto-renew card */}
                        <div style={{ padding: '1rem 1.25rem', border: '1px solid var(--border-color)', borderRadius: 'var(--radius-md)', background: 'var(--bg-secondary)' }}>
                            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.75rem', fontWeight: 600, fontSize: '0.9rem' }}>
                                <RotateCw size={15} />
                                {t.dnshe.autoRenew.title}
                            </div>
                            <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.85rem', cursor: 'pointer', marginBottom: '0.6rem' }}>
                                <input
                                    type="checkbox"
                                    checked={autoRenew.enabled}
                                    onChange={e => handleAutoRenewChange({ enabled: e.target.checked })}
                                    style={{ width: 'auto' }}
                                />
                                {t.dnshe.autoRenew.enabled}
                            </label>
                            <div style={{ marginBottom: '0.4rem' }}>
                                <label className="form-label" style={{ fontSize: '0.75rem', marginBottom: '0.25rem' }}>{t.dnshe.autoRenew.daysBefore}</label>
                                <input
                                    type="number"
                                    min="1"
                                    max="365"
                                    className="form-input"
                                    style={{ height: '32px', fontSize: '0.85rem', width: '100%' }}
                                    value={autoRenew.days_before}
                                    onChange={e => {
                                        const v = parseInt(e.target.value) || 7;
                                        setAutoRenew(prev => ({ ...prev, days_before: v }));
                                    }}
                                    onBlur={e => handleAutoRenewChange({ days_before: parseInt(e.target.value) || 7 })}
                                />
                                <div style={{ fontSize: '0.7rem', color: 'var(--text-tertiary)', marginTop: '0.25rem' }}>
                                    {t.dnshe.autoRenew.daysBeforeHint}
                                </div>
                            </div>
                            <div style={{ fontSize: '0.72rem', color: 'var(--text-tertiary)', marginBottom: '0.6rem' }}>
                                {t.dnshe.autoRenew.lastRun}: {fmtDateTime(autoRenew.last_run_at)}
                            </div>
                            <button
                                onClick={handleTriggerAutoRenew}
                                className="btn btn-secondary"
                                style={{ height: '30px', fontSize: '0.8rem', padding: '0 10px', display: 'flex', alignItems: 'center', gap: '0.4rem' }}
                                disabled={autoRenewTriggering}
                            >
                                {autoRenewTriggering ? <Loader size={12} className="spin" style={{ animation: 'spin 1s linear infinite' }} /> : <RotateCw size={12} />}
                                {autoRenewTriggering ? t.dnshe.autoRenew.triggering : t.dnshe.autoRenew.trigger}
                            </button>
                        </div>
                    </div>

                    {/* Register button + search */}
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: '1rem', marginBottom: '1rem', flexWrap: 'wrap' }}>
                        <button
                            onClick={() => setRegisterOpen(true)}
                            className="btn btn-primary"
                            style={{ height: '34px', padding: '0 12px', display: 'flex', alignItems: 'center', gap: '0.4rem' }}
                            disabled={!selectedAccountId}
                        >
                            <Plus size={15} />
                            {t.dnshe.register.button}
                        </button>
                        <div className="search-box-fixed" style={{ position: 'relative' }}>
                            <Search size={14} style={{ position: 'absolute', left: '0.75rem', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-tertiary)' }} />
                            <input
                                type="text"
                                className="form-input"
                                placeholder={t.dnshe.domains.searchPlaceholder}
                                style={{ paddingLeft: '2.25rem', width: '220px', height: '34px' }}
                                value={searchTerm}
                                onChange={e => setSearchTerm(e.target.value)}
                            />
                        </div>
                    </div>

                    {/* Domain list */}
                    {loading && !domains.length ? (
                        <div style={{ textAlign: 'center', padding: '4rem' }}>
                            <div className="spinner" style={{ margin: '0 auto 1rem' }}></div>
                            <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>{t.dnshe.domains.loading}</p>
                        </div>
                    ) : (
                        <div style={{ display: 'grid', gap: '0.75rem' }}>
                            {domains.some(d => d.uses_dnshe_dns === false) && (
                                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', padding: '0.6rem 1rem', backgroundColor: 'rgba(255, 140, 0, 0.06)', border: '1px solid rgba(255, 140, 0, 0.18)', borderRadius: 'var(--radius-sm)', fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
                                    <AlertCircle size={14} style={{ color: '#ff8c42', flexShrink: 0 }} />
                                    <span>{t.dnshe.resolution.thirdPartyHint}</span>
                                </div>
                            )}
                            {filteredDomains.map(domain => {
                                const usesDNSHE = domain.uses_dnshe_dns !== false; // default true when undefined
                                return (
                                    <div key={`${domain.account_id}-${domain.id}`} className="domain-list-card">
                                        <div className="domain-card-layout" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: '1rem' }}>
                                            <div className="domain-card-main" style={{ display: 'flex', gap: '0.75rem', alignItems: 'center', flex: 1, minWidth: 0 }}>
                                                <Globe2 size={18} style={{ color: 'var(--text-secondary)', flexShrink: 0 }} />
                                                <div style={{ flex: 1, minWidth: 0 }}>
                                                    <h3 className="font-mono" style={{ fontSize: '15px', fontWeight: 600, margin: 0, marginBottom: '0.25rem', color: 'var(--text-primary)', wordBreak: 'break-all' }}>
                                                        {domain.name}
                                                    </h3>
                                                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', flexWrap: 'wrap' }}>
                                                        <span className="badge badge-neutral" style={{ gap: '0.25rem' }}>
                                                            <Server size={11} />
                                                            <span>{domain.account_name || selectedAccount?.name}</span>
                                                        </span>
                                                        {domain.renewal_date === 'permanent' ? (
                                                            <span className="badge badge-neutral" style={{ gap: '0.25rem' }}>
                                                                <Calendar size={11} />
                                                                <span>{t.dnshe.domains.permanentFree}</span>
                                                            </span>
                                                        ) : domain.renewal_date && (
                                                            <span className="badge badge-neutral" style={{ gap: '0.25rem' }}>
                                                                <Calendar size={11} />
                                                                <span>{domain.renewal_date}</span>
                                                            </span>
                                                        )}
                                                        {/* Resolution ownership badge */}
                                                        <span
                                                            className="badge"
                                                            style={{
                                                                gap: '0.25rem',
                                                                background: usesDNSHE ? 'rgba(0, 112, 243, 0.1)' : 'rgba(255, 140, 0, 0.1)',
                                                                color: usesDNSHE ? 'var(--accent-primary)' : '#ff8c42',
                                                                border: `1px solid ${usesDNSHE ? 'rgba(0, 112, 243, 0.2)' : 'rgba(255, 140, 0, 0.2)'}`,
                                                            }}
                                                            title={usesDNSHE ? t.dnshe.resolution.dnsheHint : t.dnshe.resolution.thirdPartyHint}
                                                        >
                                                            {usesDNSHE ? <ToggleRight size={12} /> : <ToggleLeft size={12} />}
                                                            {usesDNSHE ? t.dnshe.resolution.dnshe : t.dnshe.resolution.thirdParty}
                                                        </span>
                                                    </div>
                                                </div>
                                            </div>
                                            <div className="domain-card-actions" style={{ display: 'flex', gap: '0.5rem', flexShrink: 0 }} onClick={e => e.stopPropagation()}>
                                                <button
                                                    onClick={() => handleToggleResolution(domain)}
                                                    className="btn btn-secondary"
                                                    title={usesDNSHE ? t.dnshe.resolution.toggleToThirdParty : t.dnshe.resolution.toggleToDNSHE}
                                                    style={{ fontSize: '13px', padding: '0 10px', height: '32px', display: 'flex', alignItems: 'center', gap: '0.3rem' }}
                                                    disabled={togglingResolutionId === domain.id}
                                                >
                                                    {togglingResolutionId === domain.id
                                                        ? <Loader size={13} className="spin" style={{ animation: 'spin 1s linear infinite' }} />
                                                        : (usesDNSHE ? <ToggleLeft size={13} /> : <ToggleRight size={13} />)}
                                                    {usesDNSHE ? t.dnshe.resolution.thirdParty : t.dnshe.resolution.dnshe}
                                                </button>
                                                <button
                                                    onClick={() => openCfModal(domain)}
                                                    className="btn btn-secondary"
                                                    title={t.dnshe.resolveToCF.button}
                                                    style={{ fontSize: '13px', padding: '0 10px', height: '32px', display: 'flex', alignItems: 'center', gap: '0.3rem' }}
                                                >
                                                    <Cloud size={13} />
                                                    {t.dnshe.resolveToCF.button}
                                                </button>
                                                <Link
                                                    to={`/accounts/${selectedAccountId}/domains/${domain.id}/records`}
                                                    className="btn btn-secondary"
                                                    style={{ fontSize: '13px', padding: '0 10px', height: '32px', display: 'flex', alignItems: 'center', gap: '0.3rem' }}
                                                >
                                                    {t.dnshe.domains.manageRecords}
                                                    <ExternalLink size={13} />
                                                </Link>
                                                <button
                                                    onClick={() => setDeleteTarget(domain)}
                                                    className="btn btn-secondary"
                                                    title={t.dnshe.delete.button}
                                                    style={{ fontSize: '13px', padding: '0 10px', height: '32px', color: 'var(--danger)' }}
                                                >
                                                    <Trash2 size={13} />
                                                </button>
                                            </div>
                                        </div>
                                    </div>
                                );
                            })}
                            {filteredDomains.length === 0 && (
                                <div style={{ textAlign: 'center', padding: '4rem', color: 'var(--text-secondary)', border: '1px dashed var(--border-color)', borderRadius: 'var(--radius-md)' }}>
                                    {searchTerm ? t.common.noSearchResults : t.dnshe.domains.noDomains}
                                </div>
                            )}
                        </div>
                    )}
                </>
            )}

            {/* Register Modal */}
            <Modal isOpen={registerOpen} onClose={() => setRegisterOpen(false)} title={t.dnshe.register.title}>
                <form onSubmit={handleRegister}>
                    {registerError && (
                        <div style={{ backgroundColor: 'rgba(239, 68, 68, 0.1)', color: 'var(--danger)', padding: '0.75rem', borderRadius: 'var(--radius-md)', marginBottom: '1rem', fontSize: '0.875rem' }}>
                            {registerError}
                        </div>
                    )}
                    <div className="form-group" style={{ marginBottom: '1rem' }}>
                        <label className="form-label" style={{ fontSize: '0.85rem' }}>{t.dnshe.register.subdomain}</label>
                        <input
                            type="text"
                            className="form-input"
                            placeholder={t.dnshe.register.subdomainPlaceholder}
                            value={registerForm.subdomain}
                            onChange={e => setRegisterForm(prev => ({ ...prev, subdomain: e.target.value }))}
                            required
                            style={{ height: '36px' }}
                        />
                    </div>
                    <div className="form-group" style={{ marginBottom: '1rem' }}>
                        <label className="form-label" style={{ fontSize: '0.85rem' }}>{t.dnshe.register.rootdomain}</label>
                        <select
                            className="form-input"
                            value={useCustomRoot ? '__custom__' : registerForm.rootdomain}
                            onChange={e => {
                                if (e.target.value === '__custom__') {
                                    setUseCustomRoot(true);
                                    setRegisterForm(prev => ({ ...prev, rootdomain: '' }));
                                } else {
                                    setUseCustomRoot(false);
                                    setRegisterForm(prev => ({ ...prev, rootdomain: e.target.value }));
                                }
                            }}
                            style={{ height: '36px', marginBottom: useCustomRoot ? '0.5rem' : 0 }}
                        >
                            <option value="" disabled>{t.dnshe.register.rootdomainSelect}</option>
                            {PRESET_ROOT_DOMAINS.map(d => (
                                <option key={d} value={d}>{d}</option>
                            ))}
                            <option value="__custom__">{t.dnshe.register.rootdomainCustom}</option>
                        </select>
                        {useCustomRoot && (
                            <input
                                type="text"
                                className="form-input"
                                placeholder={t.dnshe.register.rootdomainPlaceholder}
                                value={registerForm.rootdomain}
                                onChange={e => setRegisterForm(prev => ({ ...prev, rootdomain: e.target.value }))}
                                required
                                style={{ height: '36px' }}
                            />
                        )}
                    </div>
                    <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '1.5rem' }}>
                        <button type="button" onClick={() => setRegisterOpen(false)} className="btn btn-ghost">{t.common.cancel}</button>
                        <button type="submit" className="btn btn-primary" disabled={registerSubmitting}>
                            {registerSubmitting ? <div className="spinner" style={{ width: '1rem', height: '1rem', borderWidth: '2px' }}></div> : t.dnshe.register.button}
                        </button>
                    </div>
                </form>
            </Modal>

            {/* Resolve to Cloudflare Modal */}
            <Modal isOpen={cfModalOpen} onClose={() => !cfSubmitting && setCfModalOpen(false)} title={t.dnshe.resolveToCF.modalTitle}>
                <form onSubmit={handleResolveToCloudflare}>
                    <div style={{ marginBottom: '1rem' }}>
                        <p style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '1rem', lineHeight: 1.5 }}>
                            {t.dnshe.resolveToCF.hint}
                        </p>
                        {cfTargetDomain && (
                            <div style={{ marginBottom: '1rem', padding: '0.6rem 1rem', backgroundColor: 'var(--bg-secondary)', border: '1px solid var(--border-color)', borderRadius: 'var(--radius-sm)', fontSize: '14px' }}>
                                <span className="font-mono" style={{ fontWeight: 600 }}>{cfTargetDomain.name}</span>
                            </div>
                        )}
                        <label style={{ display: 'block', fontSize: '13px', fontWeight: 500, marginBottom: '0.4rem', color: 'var(--text-secondary)' }}>
                            {t.dnshe.resolveToCF.selectAccount}
                        </label>
                        {cfAccounts.length === 0 ? (
                            <p style={{ fontSize: '13px', color: 'var(--danger)' }}>{t.dnshe.resolveToCF.noCFAccounts}</p>
                        ) : (
                            <select
                                className="form-input"
                                value={cfSelectedAccountId}
                                onChange={e => setCfSelectedAccountId(e.target.value)}
                                style={{ width: '100%', height: '38px' }}
                                disabled={cfSubmitting}
                            >
                                {cfAccounts.map(acc => (
                                    <option key={acc.id} value={acc.id}>{acc.name}</option>
                                ))}
                            </select>
                        )}
                    </div>
                    {cfError && (
                        <div style={{ color: 'var(--danger)', marginBottom: '1rem', fontSize: '13px' }}>{cfError}</div>
                    )}
                    <div style={{ display: 'flex', gap: '0.5rem', justifyContent: 'flex-end' }}>
                        <button type="button" onClick={() => setCfModalOpen(false)} className="btn btn-secondary" style={{ height: '38px', padding: '0 16px' }} disabled={cfSubmitting}>
                            {t.common.cancel}
                        </button>
                        <button type="submit" className="btn btn-primary" style={{ height: '38px', padding: '0 16px', display: 'flex', alignItems: 'center', gap: '0.4rem' }} disabled={cfSubmitting || cfAccounts.length === 0}>
                            {cfSubmitting ? <Loader size={14} className="spin" style={{ animation: 'spin 1s linear infinite' }} /> : <Cloud size={14} />}
                            {cfSubmitting ? t.dnshe.resolveToCF.submitting : t.dnshe.resolveToCF.submit}
                        </button>
                    </div>
                </form>
            </Modal>

            {/* Delete Confirm */}
            <ConfirmDialog
                isOpen={!!deleteTarget}
                onClose={() => setDeleteTarget(null)}
                onConfirm={handleDelete}
                title={t.dnshe.delete.confirmTitle}
                message={deleteTarget ? t.dnshe.delete.confirmMessage.replace('{name}', deleteTarget.name) : ''}
                confirmText={t.dnshe.delete.button}
                loading={deleting}
            />
        </div>
    );
};

export default DNSHE;