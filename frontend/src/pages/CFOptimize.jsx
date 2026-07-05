import { useState, useEffect, useCallback } from 'react';
import { api } from '../api';
import { Zap, RefreshCw, Trash2, AlertCircle, CheckCircle, Clock, ExternalLink } from 'lucide-react';
import ConfirmDialog from '../components/ConfirmDialog';
import { useLanguage } from '../LanguageContext';
import useMediaQuery from '../hooks/useMediaQuery';

const CFOptimize = () => {
    const { t } = useLanguage();
    const isMobile = useMediaQuery('(max-width: 768px)');
    const [configs, setConfigs] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    // Accounts state
    const [cfAccounts, setCfAccounts] = useState([]);
    const [zones, setZones] = useState([]);
    const [zonesLoading, setZonesLoading] = useState(false);

    // Form state
    const [formData, setFormData] = useState({
        account_id: '',
        zone_name: '',
        hostname: '',
        origin_ip: '',
        cname_target: '',
    });
    const [formError, setFormError] = useState('');
    const [submitting, setSubmitting] = useState(false);

    // Delete confirmation
    const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
    const [deletingConfig, setDeletingConfig] = useState(null);
    const [deleting, setDeleting] = useState(false);

    // Refresh state
    const [refreshingId, setRefreshingId] = useState(null);

    // Load configs and CF accounts
    const loadData = useCallback(async () => {
        try {
            setLoading(true);
            const [configsData, accountsData] = await Promise.all([
                api.cfOptimizeList(),
                api.getAccounts(),
            ]);
            setConfigs(configsData);
            const cfOnly = accountsData.filter(a => a.provider_type === 'cloudflare');
            setCfAccounts(cfOnly);
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        loadData();
    }, [loadData]);

    // Load zones when account changes
    useEffect(() => {
        const loadZones = async () => {
            if (!formData.account_id) {
                setZones([]);
                return;
            }
            setZonesLoading(true);
            try {
                const domainsData = await api.getDomains(formData.account_id);
                setZones(domainsData);
            } catch (err) {
                console.error('Failed to load zones:', err);
                setZones([]);
            } finally {
                setZonesLoading(false);
            }
        };
        loadZones();
    }, [formData.account_id]);

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
            setFormData({ account_id: '', zone_name: '', hostname: '', origin_ip: '', cname_target: '' });
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
        <div style={{ padding: isMobile ? '16px' : '24px', maxWidth: '1200px', margin: '0 auto' }}>
            {/* Header */}
            <div style={{ marginBottom: '24px' }}>
                <h1 style={{ fontSize: '24px', fontWeight: '700', marginBottom: '4px', display: 'flex', alignItems: 'center', gap: '8px' }}>
                    <Zap size={24} />
                    {t.cfOptimize.title}
                </h1>
                <p style={{ color: 'var(--text-secondary)', fontSize: '14px' }}>{t.cfOptimize.subtitle}</p>
            </div>

            {/* Error */}
            {error && (
                <div style={{ padding: '12px 16px', marginBottom: '16px', background: 'var(--danger-bg, #fef2f2)', border: '1px solid var(--danger-border, #fecaca)', borderRadius: '8px', color: 'var(--danger-text, #dc2626)', display: 'flex', alignItems: 'center', gap: '8px' }}>
                    <AlertCircle size={16} />
                    {error}
                    <button onClick={() => setError('')} style={{ marginLeft: 'auto', background: 'none', border: 'none', cursor: 'pointer', color: 'inherit' }}>×</button>
                </div>
            )}

            {/* Create Form */}
            <div className="card" style={{ padding: '20px', marginBottom: '24px' }}>
                <form onSubmit={handleSubmit} style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : 'repeat(auto-fit, minmax(200px, 1fr))', gap: '12px', alignItems: 'end' }}>
                    {/* Account */}
                    <div className="form-group" style={{ marginBottom: 0 }}>
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
                            <span style={{ fontSize: '12px', color: 'var(--text-secondary)', marginTop: '4px', display: 'block' }}>
                                {t.cfOptimize.form.noCFAccount}
                            </span>
                        )}
                    </div>

                    {/* Zone */}
                    <div className="form-group" style={{ marginBottom: 0 }}>
                        <label className="form-label">{t.cfOptimize.form.zone}</label>
                        <select
                            className="form-input"
                            value={formData.zone_name}
                            onChange={e => setFormData(prev => ({ ...prev, zone_name: e.target.value }))}
                            disabled={!formData.account_id || zonesLoading}
                        >
                            <option value="">{zonesLoading ? t.cfOptimize.form.zoneLoading : t.cfOptimize.form.zonePlaceholder}</option>
                            {zones.map(z => (
                                <option key={z.id} value={z.name}>{z.name}</option>
                            ))}
                        </select>
                    </div>

                    {/* Hostname */}
                    <div className="form-group" style={{ marginBottom: 0 }}>
                        <label className="form-label">{t.cfOptimize.form.hostname}</label>
                        <input
                            className="form-input"
                            type="text"
                            placeholder={t.cfOptimize.form.hostnamePlaceholder}
                            value={formData.hostname}
                            onChange={e => setFormData(prev => ({ ...prev, hostname: e.target.value }))}
                        />
                    </div>

                    {/* Origin IP */}
                    <div className="form-group" style={{ marginBottom: 0 }}>
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
                    <div className="form-group" style={{ marginBottom: 0 }}>
                        <label className="form-label">{t.cfOptimize.form.cnameTarget}</label>
                        <input
                            className="form-input"
                            type="text"
                            placeholder={t.cfOptimize.form.cnameTargetPlaceholder}
                            value={formData.cname_target}
                            onChange={e => setFormData(prev => ({ ...prev, cname_target: e.target.value }))}
                        />
                    </div>

                    {/* Submit */}
                    <div className="form-group" style={{ marginBottom: 0 }}>
                        <button
                            type="submit"
                            className="btn btn-primary"
                            disabled={submitting || !formData.account_id || !formData.zone_name || !formData.hostname || !formData.origin_ip}
                            style={{ width: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '6px' }}
                        >
                            {submitting ? (
                                <span className="spinner" />
                            ) : (
                                <Zap size={16} />
                            )}
                            {t.cfOptimize.createButton}
                        </button>
                    </div>
                </form>
                {formError && (
                    <div style={{ marginTop: '8px', color: 'var(--danger-text, #dc2626)', fontSize: '13px' }}>
                        {formError}
                    </div>
                )}
            </div>

            {/* Configs Table */}
            {loading ? (
                <div style={{ textAlign: 'center', padding: '40px', color: 'var(--text-secondary)' }}>
                    <span className="spinner" />
                </div>
            ) : configs.length === 0 ? (
                <div style={{ textAlign: 'center', padding: '40px', color: 'var(--text-secondary)' }}>
                    {t.cfOptimize.empty}
                </div>
            ) : isMobile ? (
                /* Mobile: Card layout */
                <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                    {configs.map(config => {
                        const statusBadge = getStatusBadge(config.status);
                        const sslBadge = getSSLBadge(config.ssl_status);
                        return (
                            <div key={config.id} className="domain-list-card" style={{ padding: '16px' }}>
                                <div style={{ fontWeight: '600', fontSize: '15px', marginBottom: '8px', fontFamily: 'var(--font-mono, monospace)' }}>
                                    {config.custom_hostname}
                                </div>
                                <div style={{ fontSize: '13px', color: 'var(--text-secondary)', display: 'flex', flexDirection: 'column', gap: '4px' }}>
                                    <span>{t.cfOptimize.table.zone}: {config.zone_name}</span>
                                    <span>{t.cfOptimize.table.originIP}: {config.origin_ip}</span>
                                    <span>{t.cfOptimize.table.cnameTarget}: <span style={{ fontFamily: 'var(--font-mono, monospace)' }}>{config.cname_target}</span></span>
                                </div>
                                <div style={{ display: 'flex', gap: '8px', marginTop: '8px', flexWrap: 'wrap' }}>
                                    <span className={statusBadge.className}>{statusBadge.text}</span>
                                    <span className={sslBadge.className}>{sslBadge.text}</span>
                                </div>
                                <div style={{ display: 'flex', gap: '8px', marginTop: '12px' }}>
                                    <button
                                        className="btn btn-ghost"
                                        onClick={() => handleRefresh(config.id)}
                                        disabled={refreshingId === config.id}
                                        style={{ fontSize: '13px' }}
                                    >
                                        <RefreshCw size={14} className={refreshingId === config.id ? 'spinner' : ''} />
                                        {t.cfOptimize.actions.refresh}
                                    </button>
                                    <button
                                        className="btn btn-ghost"
                                        onClick={() => handleDeleteClick(config)}
                                        style={{ fontSize: '13px', color: 'var(--danger-text, #dc2626)' }}
                                    >
                                        <Trash2 size={14} />
                                        {t.cfOptimize.actions.delete}
                                    </button>
                                </div>
                            </div>
                        );
                    })}
                </div>
            ) : (
                /* Desktop: Table layout */
                <div style={{ overflowX: 'auto' }}>
                    <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                        <thead>
                            <tr style={{ borderBottom: '1px solid var(--border-color)' }}>
                                <th style={{ padding: '12px 16px', textAlign: 'left', fontWeight: '600', fontSize: '13px' }}>{t.cfOptimize.table.customHostname}</th>
                                <th style={{ padding: '12px 16px', textAlign: 'left', fontWeight: '600', fontSize: '13px' }}>{t.cfOptimize.table.zone}</th>
                                <th style={{ padding: '12px 16px', textAlign: 'left', fontWeight: '600', fontSize: '13px' }}>{t.cfOptimize.table.originIP}</th>
                                <th style={{ padding: '12px 16px', textAlign: 'left', fontWeight: '600', fontSize: '13px' }}>{t.cfOptimize.table.cnameTarget}</th>
                                <th style={{ padding: '12px 16px', textAlign: 'left', fontWeight: '600', fontSize: '13px' }}>{t.cfOptimize.table.status}</th>
                                <th style={{ padding: '12px 16px', textAlign: 'left', fontWeight: '600', fontSize: '13px' }}>{t.cfOptimize.table.actions}</th>
                            </tr>
                        </thead>
                        <tbody>
                            {configs.map(config => {
                                const statusBadge = getStatusBadge(config.status);
                                const sslBadge = getSSLBadge(config.ssl_status);
                                return (
                                    <tr key={config.id} style={{ borderBottom: '1px solid var(--border-color-light, #f0f0f0)' }}>
                                        <td style={{ padding: '12px 16px', fontFamily: 'var(--font-mono, monospace)', fontSize: '14px' }}>
                                            {config.custom_hostname}
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
                                                    <RefreshCw size={14} className={refreshingId === config.id ? 'spinner' : ''} />
                                                </button>
                                                <button
                                                    className="btn btn-ghost"
                                                    onClick={() => handleDeleteClick(config)}
                                                    title={t.cfOptimize.actions.delete}
                                                    style={{ padding: '6px 8px', color: 'var(--danger-text, #dc2626)' }}
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
