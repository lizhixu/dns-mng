import { useState, useEffect, useCallback, useRef } from 'react';
import { api } from '../api';
import { Zap, RefreshCw, Trash2, AlertCircle, CheckCircle, Server } from 'lucide-react';
import Modal from '../components/Modal';
import ConfirmDialog from '../components/ConfirmDialog';
import { useLanguage } from '../LanguageContext';
import useMediaQuery from '../hooks/useMediaQuery';

const CFOptimize = () => {
    const { t } = useLanguage();
    const isMobile = useMediaQuery('(max-width: 768px)');
    const fetchedRef = useRef(false);
    const [configs, setConfigs] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');

    // Accounts & domains state
    const [cfAccounts, setCfAccounts] = useState([]);
    const [allDomains, setAllDomains] = useState([]);

    // Modal state
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [formData, setFormData] = useState({
        account_id: '',
        zone_name: '',
        hostname: '',
        origin_ip: '',
        cname_target: '',
    });
    const [formError, setFormError] = useState('');
    const [submitting, setSubmitting] = useState(false);

    // Derived: filter zones by selected account
    const zones = formData.account_id
        ? allDomains.filter(d => String(d.account_id) === String(formData.account_id))
        : [];

    // CNAME target presets
    const cnamePresets = [
        { value: 'cloudflare.468123.xyz', label: 'cloudflare.468123.xyz', desc: t.cfOptimize.cnamePresets.cloudflare },
        { value: 'cf.468123.xyz', label: 'cf.468123.xyz', desc: t.cfOptimize.cnamePresets.cf },
        { value: 'cloudflare-dl.468123.xyz', label: 'cloudflare-dl.468123.xyz', desc: t.cfOptimize.cnamePresets.dl },
    ];
    const [cnameMode, setCnameMode] = useState('preset');

    // Delete confirmation
    const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
    const [deletingConfig, setDeletingConfig] = useState(null);
    const [deleting, setDeleting] = useState(false);

    // Refresh state
    const [refreshingId, setRefreshingId] = useState(null);

    // Load configs and CF accounts
    const loadData = useCallback(async () => {
        setLoading(true);
        try {
            const [configsData, accountsData, domainsRes] = await Promise.all([
                api.cfOptimizeList(),
                api.getAccounts(),
                api.getAllDomains(),
            ]);
            setConfigs(configsData);
            const cfOnly = accountsData.filter(a => a.provider_type === 'cloudflare');
            setCfAccounts(cfOnly);
            setAllDomains(Array.isArray(domainsRes) ? domainsRes : (domainsRes.domains || []));
            setError('');
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        if (fetchedRef.current) return;
        fetchedRef.current = true;
        loadData();
    }, [loadData]);

    const openModal = () => {
        setFormData({ account_id: '', zone_name: '', hostname: '', origin_ip: '', cname_target: cnamePresets[0].value });
        setFormError('');
        setCnameMode('preset');
        setIsModalOpen(true);
    };

    const closeModal = () => {
        setIsModalOpen(false);
        setFormError('');
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (!formData.account_id || !formData.zone_name || !formData.hostname || !formData.origin_ip) {
            setFormError(t.common.error);
            return;
        }
        setFormError('');
        setSubmitting(true);
        try {
            await api.cfOptimizeCreate({
                account_id: Number(formData.account_id),
                zone_name: formData.zone_name,
                hostname: formData.hostname,
                origin_ip: formData.origin_ip,
                cname_target: formData.cname_target || undefined,
            });
            closeModal();
            setSuccess(t.cfOptimize.messages.createSuccess);
            setTimeout(() => setSuccess(''), 3000);
            loadData();
        } catch (err) {
            setFormError(err.message);
        } finally {
            setSubmitting(false);
        }
    };

    const handleRefresh = async (id) => {
        setRefreshingId(id);
        try {
            const updated = await api.cfOptimizeRefresh(id);
            setConfigs(prev => prev.map(c => c.id === id ? { ...c, ...updated } : c));
        } catch (err) {
            setError(err.message);
        } finally {
            setRefreshingId(null);
        }
    };

    const handleDeleteClick = (config) => {
        setDeletingConfig(config);
        setShowDeleteConfirm(true);
    };

    const confirmDelete = async () => {
        if (!deletingConfig) return;
        setDeleting(true);
        try {
            await api.cfOptimizeDelete(deletingConfig.id, true);
            setConfigs(prev => prev.filter(c => c.id !== deletingConfig.id));
            setShowDeleteConfirm(false);
            setDeletingConfig(null);
        } catch (err) {
            setError(err.message);
        } finally {
            setDeleting(false);
        }
    };

    const getStatusBadge = (status) => {
        const s = (status || '').toLowerCase();
        if (s === 'active') return { className: 'badge badge-success', text: t.cfOptimize.status.active };
        if (s === 'pending' || s === 'initializing') return { className: 'badge badge-neutral', text: t.cfOptimize.status.pending };
        return { className: 'badge', text: status || t.cfOptimize.status.pending };
    };

    const getSSLBadge = (status) => {
        const s = (status || '').toLowerCase();
        if (s === 'active') return { className: 'badge badge-success', text: t.cfOptimize.sslStatus.active };
        if (s === 'pending' || s === 'initializing') return { className: 'badge badge-neutral', text: t.cfOptimize.sslStatus.pending };
        return { className: 'badge', text: status || t.cfOptimize.sslStatus.pending };
    };

    return (
        <div>
            {/* Header */}
            <div style={{ marginBottom: '1.5rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
                    <div style={{ minWidth: 0, flex: 1 }}>
                        <h2 style={{ fontSize: '1.5rem', fontWeight: 'bold', letterSpacing: '-0.02em', margin: 0 }}>
                            {t.cfOptimize.title}
                        </h2>
                        <p style={{ color: 'var(--text-secondary)', fontSize: '13px', marginTop: '0.25rem', marginBottom: 0 }}>
                            {t.cfOptimize.subtitle}
                        </p>
                    </div>
                    <div style={{ display: 'flex', gap: '0.75rem' }}>
                        <button onClick={() => loadData()} className="btn btn-secondary" style={{ height: '34px', padding: '0 10px' }} title={t.common.refresh}>
                            <RefreshCw size={15} style={{ animation: loading ? 'spin 1s linear infinite' : 'none' }} />
                        </button>
                        <button onClick={openModal} className="btn btn-primary" style={{ height: '34px', padding: '0 12px', display: 'flex', alignItems: 'center', gap: '6px' }}>
                            <Zap size={15} />
                            {t.cfOptimize.createButton}
                        </button>
                    </div>
                </div>
            </div>

            {/* Error */}
            {error && (
                <div style={{ color: 'var(--danger)', marginBottom: '1rem', padding: '0.75rem 1rem', backgroundColor: 'rgba(255, 0, 0, 0.05)', border: '1px solid rgba(255, 0, 0, 0.15)', borderRadius: 'var(--radius-sm)', fontSize: '14px', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                    <AlertCircle size={16} />
                    {error}
                </div>
            )}

            {/* Success */}
            {success && (
                <div style={{ color: 'var(--success)', marginBottom: '1rem', padding: '0.75rem 1rem', backgroundColor: 'rgba(0, 224, 84, 0.05)', border: '1px solid rgba(0, 224, 84, 0.15)', borderRadius: 'var(--radius-sm)', fontSize: '14px', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                    <CheckCircle size={16} />
                    {success}
                </div>
            )}

            {/* Loading */}
            {loading && !configs.length ? (
                <div style={{ textAlign: 'center', padding: '4rem' }}>
                    <div className="spinner" style={{ margin: '0 auto 1rem' }}></div>
                </div>
            ) : configs.length === 0 ? (
                /* Empty state */
                <div style={{ textAlign: 'center', padding: '4rem', color: 'var(--text-secondary)', border: '1px dashed var(--border-color)', borderRadius: 'var(--radius-md)' }}>
                    {t.cfOptimize.empty}
                </div>
            ) : isMobile ? (
                /* Mobile: Card layout */
                <div style={{ display: 'grid', gap: '0.75rem' }}>
                    {configs.map(config => {
                        const statusBadge = getStatusBadge(config.status);
                        const sslBadge = getSSLBadge(config.ssl_status);
                        return (
                            <div key={config.id} className="domain-list-card">
                                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: '1rem', marginBottom: '0.5rem' }}>
                                    <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center', flex: 1, minWidth: 0 }}>
                                        <Zap size={18} style={{ color: 'var(--text-secondary)', flexShrink: 0 }} />
                                        <div style={{ flex: 1, minWidth: 0 }}>
                                            <h3 className="font-mono" style={{ fontSize: '15px', fontWeight: '600', margin: 0, marginBottom: '0.25rem', color: 'var(--text-primary)', wordBreak: 'break-all' }}>
                                                {config.custom_hostname}
                                            </h3>
                                            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', flexWrap: 'wrap' }}>
                                                <span className="badge badge-neutral" style={{ gap: '0.25rem' }}>
                                                    <Server size={11} />
                                                    <span>{config.zone_name}</span>
                                                </span>
                                                <span className={statusBadge.className}>{statusBadge.text}</span>
                                                <span className={sslBadge.className}>{sslBadge.text}</span>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                                <div style={{ fontSize: '13px', color: 'var(--text-secondary)', display: 'flex', flexDirection: 'column', gap: '0.25rem', marginBottom: '0.75rem' }}>
                                    <span>{t.cfOptimize.table.originIP}: <span className="font-mono">{config.origin_ip}</span></span>
                                    <span>{t.cfOptimize.table.cnameTarget}: <span className="font-mono">{config.cname_target}</span></span>
                                </div>
                                <div style={{ display: 'flex', gap: '0.5rem' }}>
                                    <button
                                        className="btn btn-secondary"
                                        onClick={() => handleRefresh(config.id)}
                                        disabled={refreshingId === config.id}
                                        style={{ fontSize: '13px', padding: '0 10px', height: '32px' }}
                                    >
                                        <RefreshCw size={13} style={{ animation: refreshingId === config.id ? 'spin 1s linear infinite' : 'none' }} />
                                        {t.cfOptimize.actions.refresh}
                                    </button>
                                    <button
                                        className="btn btn-ghost"
                                        onClick={() => handleDeleteClick(config)}
                                        style={{ fontSize: '13px', padding: '0 10px', height: '32px', color: 'var(--danger)' }}
                                    >
                                        <Trash2 size={13} />
                                        {t.cfOptimize.actions.delete}
                                    </button>
                                </div>
                            </div>
                        );
                    })}
                </div>
            ) : (
                /* Desktop: Table layout */
                <div className="domain-list-card" style={{ padding: 0, overflow: 'hidden' }}>
                    <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                        <thead>
                            <tr style={{ borderBottom: '1px solid var(--border-color)' }}>
                                <th style={{ padding: '12px 16px', textAlign: 'left', fontWeight: '600', fontSize: '13px', color: 'var(--text-secondary)' }}>{t.cfOptimize.table.customHostname}</th>
                                <th style={{ padding: '12px 16px', textAlign: 'left', fontWeight: '600', fontSize: '13px', color: 'var(--text-secondary)' }}>{t.cfOptimize.table.zone}</th>
                                <th style={{ padding: '12px 16px', textAlign: 'left', fontWeight: '600', fontSize: '13px', color: 'var(--text-secondary)' }}>{t.cfOptimize.table.originIP}</th>
                                <th style={{ padding: '12px 16px', textAlign: 'left', fontWeight: '600', fontSize: '13px', color: 'var(--text-secondary)' }}>{t.cfOptimize.table.cnameTarget}</th>
                                <th style={{ padding: '12px 16px', textAlign: 'left', fontWeight: '600', fontSize: '13px', color: 'var(--text-secondary)' }}>{t.cfOptimize.table.status}</th>
                                <th style={{ padding: '12px 16px', textAlign: 'left', fontWeight: '600', fontSize: '13px', color: 'var(--text-secondary)' }}>{t.cfOptimize.table.actions}</th>
                            </tr>
                        </thead>
                        <tbody>
                            {configs.map(config => {
                                const statusBadge = getStatusBadge(config.status);
                                const sslBadge = getSSLBadge(config.ssl_status);
                                return (
                                    <tr key={config.id} style={{ borderBottom: '1px solid var(--border-color)' }}>
                                        <td style={{ padding: '12px 16px' }}>
                                            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                                <Zap size={14} style={{ color: 'var(--text-secondary)', flexShrink: 0 }} />
                                                <span className="font-mono" style={{ fontSize: '14px', fontWeight: '500' }}>{config.custom_hostname}</span>
                                            </div>
                                        </td>
                                        <td style={{ padding: '12px 16px', fontSize: '14px' }}>{config.zone_name}</td>
                                        <td style={{ padding: '12px 16px', fontFamily: 'var(--font-mono, monospace)', fontSize: '14px' }}>{config.origin_ip}</td>
                                        <td style={{ padding: '12px 16px', fontFamily: 'var(--font-mono, monospace)', fontSize: '14px' }}>{config.cname_target}</td>
                                        <td style={{ padding: '12px 16px' }}>
                                            <div style={{ display: 'flex', gap: '6px', flexWrap: 'wrap' }}>
                                                <span className={statusBadge.className}>{statusBadge.text}</span>
                                                <span className={sslBadge.className}>{sslBadge.text}</span>
                                            </div>
                                        </td>
                                        <td style={{ padding: '12px 16px' }}>
                                            <div style={{ display: 'flex', gap: '4px' }}>
                                                <button
                                                    className="btn btn-ghost"
                                                    onClick={() => handleRefresh(config.id)}
                                                    disabled={refreshingId === config.id}
                                                    title={t.cfOptimize.actions.refresh}
                                                    style={{ padding: '6px 8px' }}
                                                >
                                                    <RefreshCw size={14} style={{ animation: refreshingId === config.id ? 'spin 1s linear infinite' : 'none' }} />
                                                </button>
                                                <button
                                                    className="btn btn-ghost"
                                                    onClick={() => handleDeleteClick(config)}
                                                    title={t.cfOptimize.actions.delete}
                                                    style={{ padding: '6px 8px', color: 'var(--danger)' }}
                                                >
                                                    <Trash2 size={14} />
                                                </button>
                                            </div>
                                        </td>
                                    </tr>
                                );
                            })}
                        </tbody>
                    </table>
                </div>
            )}

            {/* Create Modal */}
            <Modal
                isOpen={isModalOpen}
                onClose={closeModal}
                title={t.cfOptimize.createButton}
            >
                <form onSubmit={handleSubmit}>
                    {formError && (
                        <div style={{ backgroundColor: 'rgba(239, 68, 68, 0.1)', color: 'var(--danger)', padding: '0.75rem', borderRadius: 'var(--radius-md)', marginBottom: '1rem', fontSize: '0.875rem' }}>
                            {formError}
                        </div>
                    )}

                    {/* Account */}
                    <div className="form-group">
                        <label className="form-label">{t.cfOptimize.form.account}</label>
                        <select
                            className="form-input"
                            value={formData.account_id}
                            onChange={e => setFormData(prev => ({ ...prev, account_id: e.target.value, zone_name: '' }))}
                        >
                            <option value="">{t.cfOptimize.form.accountPlaceholder}</option>
                            {cfAccounts.map(a => (
                                <option key={a.id} value={a.id}>{a.name}</option>
                            ))}
                        </select>
                        {cfAccounts.length === 0 && !loading && (
                            <span style={{ fontSize: '12px', color: 'var(--text-tertiary)', marginTop: '0.25rem', display: 'block' }}>
                                {t.cfOptimize.form.noCFAccount}
                            </span>
                        )}
                    </div>

                    {/* Zone */}
                    <div className="form-group">
                        <label className="form-label">{t.cfOptimize.form.zone}</label>
                        <select
                            className="form-input"
                            value={formData.zone_name}
                            onChange={e => setFormData(prev => ({ ...prev, zone_name: e.target.value }))}
                            disabled={!formData.account_id}
                        >
                            <option value="">{t.cfOptimize.form.zonePlaceholder}</option>
                            {zones.map(z => (
                                <option key={z.id} value={z.name}>{z.name}</option>
                            ))}
                        </select>
                    </div>

                    {/* Hostname */}
                    <div className="form-group">
                        <label className="form-label">{t.cfOptimize.form.hostname}</label>
                        <input
                            className="form-input"
                            type="text"
                            placeholder={t.cfOptimize.form.hostnamePlaceholder}
                            value={formData.hostname}
                            onChange={e => setFormData(prev => ({ ...prev, hostname: e.target.value }))}
                        />
                        <span style={{ fontSize: '12px', color: 'var(--text-tertiary)', marginTop: '0.25rem', display: 'block' }}>
                            {t.cfOptimize.form.hostnameHint}
                        </span>
                    </div>

                    {/* Origin IP */}
                    <div className="form-group">
                        <label className="form-label">{t.cfOptimize.form.originIP}</label>
                        <input
                            className="form-input"
                            type="text"
                            placeholder={t.cfOptimize.form.originIPPlaceholder}
                            value={formData.origin_ip}
                            onChange={e => setFormData(prev => ({ ...prev, origin_ip: e.target.value }))}
                        />
                    </div>

                    {/* CNAME Target */}
                    <div className="form-group">
                        <label className="form-label">{t.cfOptimize.form.cnameTarget}</label>
                        <select
                            className="form-input"
                            value={cnameMode === 'custom' ? '__custom__' : formData.cname_target}
                            onChange={e => {
                                if (e.target.value === '__custom__') {
                                    setCnameMode('custom');
                                    setFormData(prev => ({ ...prev, cname_target: '' }));
                                } else {
                                    setCnameMode('preset');
                                    setFormData(prev => ({ ...prev, cname_target: e.target.value }));
                                }
                            }}
                        >
                            {cnamePresets.map(p => (
                                <option key={p.value} value={p.value}>{p.label} — {p.desc}</option>
                            ))}
                            <option value="__custom__">{t.cfOptimize.form.cnameCustom}</option>
                        </select>
                        {cnameMode === 'custom' && (
                            <input
                                className="form-input"
                                type="text"
                                placeholder={t.cfOptimize.form.cnameTargetPlaceholder}
                                value={formData.cname_target}
                                onChange={e => setFormData(prev => ({ ...prev, cname_target: e.target.value }))}
                                style={{ marginTop: '0.5rem' }}
                            />
                        )}
                    </div>

                    {/* Actions */}
                    <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '2rem' }}>
                        <button type="button" onClick={closeModal} className="btn btn-ghost">{t.common.cancel}</button>
                        <button
                            type="submit"
                            className="btn btn-primary"
                            disabled={submitting || !formData.account_id || !formData.zone_name || !formData.hostname || !formData.origin_ip}
                            style={{ display: 'flex', alignItems: 'center', gap: '6px' }}
                        >
                            {submitting ? <div className="spinner" style={{ width: '1rem', height: '1rem', borderWidth: '2px' }}></div> : <Zap size={14} />}
                            {t.cfOptimize.createButton}
                        </button>
                    </div>
                </form>
            </Modal>

            {/* Delete Confirmation */}
            <ConfirmDialog
                isOpen={showDeleteConfirm}
                onClose={() => { setShowDeleteConfirm(false); setDeletingConfig(null); }}
                onConfirm={confirmDelete}
                title={t.cfOptimize.actions.deleteWithCleanup}
                message={deletingConfig ? `${t.cfOptimize.messages.confirmDelete}\n${t.cfOptimize.messages.confirmDeleteCleanup}\n\n${deletingConfig.custom_hostname}` : ''}
                confirmText={t.common.delete}
                loading={deleting}
                danger
            />
        </div>
    );
};

export default CFOptimize;
