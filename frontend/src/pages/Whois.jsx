import { useState, useEffect, useRef } from 'react';
import { api } from '../api';
import {
    Search,
    ChevronDown,
    ChevronRight,
    AlertCircle,
    CheckCircle,
    XCircle,
    Info,
    Eye,
    EyeOff,
} from 'lucide-react';
import { useLanguage } from '../LanguageContext';

// InfoCell renders a single grid info cell. Defined at module scope (not inside
// Whois) so it doesn't become a fresh component type on every render, which
// would unmount/remount the whole result subtree on each keystroke.
const InfoCell = ({ label, value, full = false, mono = false }) => {
    if (!value && value !== 0 && value !== false) return null;
    return (
        <div style={full ? { gridColumn: '1 / -1' } : undefined}>
            <label style={{ fontSize: '0.875rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '0.5rem' }}>
                {label}
            </label>
            <div style={{
                padding: '0.75rem',
                background: 'var(--bg-secondary)',
                borderRadius: 'var(--radius-sm)',
                fontFamily: mono ? 'monospace' : 'inherit',
                fontSize: '0.875rem',
                wordBreak: 'break-all',
                whiteSpace: mono ? 'pre-wrap' : 'normal',
            }}>
                {value}
            </div>
        </div>
    );
};

// ContactList renders an owner/admin/tech contact list, or "no contact info"
// when empty. `t` is passed in so this can live at module scope.
const ContactList = ({ title, contacts, t }) => {
    const items = contacts && contacts.length > 0 ? contacts : null;
    return (
        <div>
            <label style={{ fontSize: '0.875rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '0.5rem' }}>
                {title}
            </label>
            <div style={{
                padding: '0.75rem',
                background: 'var(--bg-secondary)',
                borderRadius: 'var(--radius-sm)',
                fontSize: '0.875rem',
            }}>
                {items ? (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                        {items.map((c, idx) => (
                            <div key={idx} style={{
                                padding: '0.5rem',
                                background: 'var(--bg-tertiary)',
                                borderRadius: 'var(--radius-sm)',
                                wordBreak: 'break-all',
                            }}>
                                {[
                                    c.name,
                                    c.organization,
                                    c.email,
                                    c.country,
                                    c.handle && `handle: ${c.handle}`,
                                ].filter(Boolean).join(' · ')}
                            </div>
                        ))}
                    </div>
                ) : (
                    <span style={{ color: 'var(--text-tertiary)' }}>{t.whois.noContact}</span>
                )}
            </div>
        </div>
    );
};

const Whois = () => {
    const { t } = useLanguage();

    // Query state
    const [domain, setDomain] = useState('');
    const [result, setResult] = useState(null);
    const [loading, setLoading] = useState(false);
    const [queryError, setQueryError] = useState('');
    const [showRaw, setShowRaw] = useState(false);

    // Config state
    const [config, setConfig] = useState({
        api_key: '',
        enabled: true,
    });
    const [configured, setConfigured] = useState(false);
    const [configLoading, setConfigLoading] = useState(true);
    const [configSaving, setConfigSaving] = useState(false);
    const [configError, setConfigError] = useState('');
    const [configSuccess, setConfigSuccess] = useState('');
    const [showConfig, setShowConfig] = useState(false);
    const [showKey, setShowKey] = useState(false); // toggle API key input between password/text
    const fetchedRef = useRef(false);

    useEffect(() => {
        // Prevent duplicate requests in React StrictMode
        if (fetchedRef.current) return;
        fetchedRef.current = true;
        loadConfig();
    }, []);

    const loadConfig = async () => {
        setConfigLoading(true);
        try {
            const data = await api.getWHOISConfig();
            if (data && data.id) {
                // Config exists — server returns the key in plaintext (same as
                // the DDNS token pattern), so it can be shown/edited directly.
                setConfigured(true);
                setConfig({
                    api_key: data.api_key || '',
                    enabled: data.enabled !== undefined ? data.enabled : true,
                });
            } else {
                // Not configured yet
                setConfigured(false);
                setConfig({ api_key: '', enabled: true });
                // Auto-expand config when no key has been set
                setShowConfig(true);
            }
        } catch (err) {
            setConfigError(err.message);
        } finally {
            setConfigLoading(false);
        }
    };

    const handleQuery = async (e) => {
        e.preventDefault();
        const trimmed = domain.trim();
        if (!trimmed) return;

        // Normalize the input so users can paste a URL or a "www."-prefixed
        // host. We strip the protocol/port/path/query/anchor and keep only the
        // hostname, which is what WhoisJSON.com expects. `new URL` requires a
        // scheme, so we prepend one when the input doesn't already have it.
        let normalized = trimmed;
        try {
            const withScheme = /^[a-z][a-z0-9+.-]*:\/\//i.test(trimmed) ? trimmed : `http://${trimmed}`;
            normalized = new URL(withScheme).hostname;
        } catch {
            // Fall back to the raw input if URL parsing fails
        }
        const query = normalized || trimmed;

        setLoading(true);
        setQueryError('');
        setResult(null);
        setShowRaw(false);

        try {
            const data = await api.whoisQuery(query);
            setResult(data);
            // Collapse the config panel after a successful query (when already
            // configured) so the result isn't dwarfed by the API key form.
            if (configured) setShowConfig(false);
        } catch (err) {
            setQueryError(err.message || t.whois.queryFailed);
            // Only auto-expand the config panel on the specific "key not
            // configured" case. A loose `includes('api key')` match would
            // misfire on upstream 403/429 messages that also mention "api key".
            if ((err.message || '') === 'WHOIS API Key is not configured') {
                setShowConfig(true);
            }
        } finally {
            setLoading(false);
        }
    };

    const handleSaveConfig = async (e) => {
        e.preventDefault();
        setConfigSaving(true);
        setConfigError('');
        setConfigSuccess('');

        try {
            const data = await api.updateWHOISConfig(config);
            setConfigured(true);
            // Server returns the stored key in plaintext; fill it back so the
            // input keeps showing the current value (the eye toggle still works).
            setConfig({
                api_key: data.api_key || '',
                enabled: data && data.enabled !== undefined ? data.enabled : config.enabled,
            });
            setConfigSuccess(t.whois.configSaved);
            setTimeout(() => setConfigSuccess(''), 3000);
        } catch (err) {
            setConfigError(err.message);
        } finally {
            setConfigSaving(false);
        }
    };

    if (configLoading) {
        return (
            <div style={{ textAlign: 'center', padding: '4rem' }}>
                <div className="spinner" style={{ margin: '0 auto' }}></div>
            </div>
        );
    }

    return (
        <div>
            <div style={{ marginBottom: '1.5rem' }}>
                <h2 style={{ fontSize: '1.5rem', fontWeight: 'bold', letterSpacing: '-0.02em', margin: 0 }}>
                    {t.whois.title}
                </h2>
                <p style={{ color: 'var(--text-secondary)', fontSize: '13px', marginTop: '0.25rem', marginBottom: 0 }}>
                    {t.whois.subtitle}
                </p>
            </div>

            {/* Hint when no API key configured */}
            {!configured && (
                <div style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.5rem',
                    color: 'var(--text-primary)',
                    marginBottom: '1rem',
                    padding: '0.75rem 1rem',
                    backgroundColor: 'rgba(59, 130, 246, 0.08)',
                    border: '1px solid rgba(59, 130, 246, 0.25)',
                    borderRadius: 'var(--radius-sm)',
                    fontSize: '14px',
                }}>
                    <Info size={16} style={{ color: '#3b82f6', flexShrink: 0 }} />
                    {t.whois.keyMissing}
                </div>
            )}

            {/* Query form */}
            <div className="domain-list-card" style={{ padding: '1.25rem', cursor: 'default' }}>
                <form onSubmit={handleQuery} style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap' }}>
                    <input
                        type="text"
                        className="form-input"
                        value={domain}
                        onChange={(e) => setDomain(e.target.value)}
                        placeholder={t.whois.placeholder}
                        style={{ flex: '1', minWidth: '200px' }}
                    />
                    <button
                        type="submit"
                        className="btn btn-primary"
                        style={{ height: '34px', fontSize: '13px', display: 'inline-flex', alignItems: 'center', gap: '6px' }}
                        disabled={loading || !domain.trim()}
                    >
                        {loading ? (
                            <div className="spinner" style={{ width: '1rem', height: '1rem', borderWidth: '2px' }}></div>
                        ) : (
                            <>
                                <Search size={14} />
                                {t.whois.button}
                            </>
                        )}
                    </button>
                </form>
            </div>

            {/* Query error */}
            {queryError && (
                <div style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.5rem',
                    color: 'var(--danger)',
                    marginTop: '1rem',
                    padding: '0.75rem 1rem',
                    backgroundColor: 'rgba(255, 0, 0, 0.05)',
                    border: '1px solid rgba(255, 0, 0, 0.15)',
                    borderRadius: 'var(--radius-sm)',
                    fontSize: '14px',
                }}>
                    <AlertCircle size={16} />
                    {queryError}
                </div>
            )}

            {/* Result */}
            {result && (
                <div className="domain-list-card" style={{ padding: '1.25rem', marginTop: '1.25rem', cursor: 'default' }}>
                    {/* Status banner */}
                    <div style={{
                        padding: '1rem',
                        borderRadius: 'var(--radius-md)',
                        background: result.registered ? 'rgba(16, 185, 129, 0.1)' : 'rgba(239, 68, 68, 0.1)',
                        border: `1px solid ${result.registered ? 'rgba(16, 185, 129, 0.3)' : 'rgba(239, 68, 68, 0.3)'}`,
                        display: 'flex',
                        alignItems: 'center',
                        gap: '0.75rem',
                        marginBottom: '1.25rem',
                    }}>
                        {result.registered ? (
                            <CheckCircle size={24} style={{ color: '#10b981', flexShrink: 0 }} />
                        ) : (
                            <XCircle size={24} style={{ color: '#ef4444', flexShrink: 0 }} />
                        )}
                        <div style={{ flex: 1 }}>
                            <div style={{
                                fontWeight: '600',
                                marginBottom: '0.25rem',
                                color: result.registered ? '#10b981' : '#ef4444',
                            }}>
                                {result.registered
                                    ? `✓ ${t.whois.registered}`
                                    : `✗ ${t.whois.unregistered}`}
                            </div>
                            {result.message && (
                                <div style={{ fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
                                    {result.message === 'Domain may be unregistered or privacy-protected'
                                        ? t.whois.unregistered
                                        : result.message}
                                </div>
                            )}
                        </div>
                    </div>

                    {/* Info grid */}
                    <div style={{
                        display: 'grid',
                        gap: '1rem',
                        gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))',
                    }}>
                        <InfoCell label={t.whois.domain} value={result.domain || result.raw?.name} mono />
                        <InfoCell label={t.whois.created} value={result.created} mono />
                        <InfoCell label={t.whois.expires} value={result.expires} mono />
                        <InfoCell label={t.whois.changed} value={result.changed} mono />

                        {result.registrar && (
                            <>
                                <InfoCell label={`${t.whois.registrar} · ${t.whois.registrarName}`} value={result.registrar.name} />
                                <InfoCell label={`${t.whois.registrar} · ${t.whois.registrarUrl}`} value={result.registrar.url} mono />
                                <InfoCell label={`${t.whois.registrar} · ${t.whois.registrarEmail}`} value={result.registrar.email} mono />
                                <InfoCell label={`${t.whois.registrar} · ${t.whois.registrarPhone}`} value={result.registrar.phone} mono />
                            </>
                        )}

                        <InfoCell label={t.whois.whoisServer} value={result.whois_server} mono />
                        <InfoCell label={t.whois.dnssec} value={result.dnssec} mono />

                        {result.nameservers && result.nameservers.length > 0 && (
                            <div style={{ gridColumn: '1 / -1' }}>
                                <label style={{ fontSize: '0.875rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '0.5rem' }}>
                                    {t.whois.nameservers}
                                </label>
                                <div style={{
                                    padding: '0.75rem',
                                    background: 'var(--bg-secondary)',
                                    borderRadius: 'var(--radius-sm)',
                                    display: 'flex',
                                    flexDirection: 'column',
                                    gap: '0.5rem',
                                }}>
                                    {result.nameservers.map((ns, idx) => (
                                        <div key={idx} style={{
                                            padding: '0.5rem',
                                            background: 'var(--bg-tertiary)',
                                            borderRadius: 'var(--radius-sm)',
                                            fontFamily: 'monospace',
                                            fontSize: '0.875rem',
                                            wordBreak: 'break-all',
                                        }}>
                                            {ns}
                                        </div>
                                    ))}
                                </div>
                            </div>
                        )}

                        {result.ips && result.ips.length > 0 && (
                            <div style={{ gridColumn: '1 / -1' }}>
                                <label style={{ fontSize: '0.875rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '0.5rem' }}>
                                    {t.whois.ips}
                                </label>
                                <div style={{
                                    padding: '0.75rem',
                                    background: 'var(--bg-secondary)',
                                    borderRadius: 'var(--radius-sm)',
                                    display: 'flex',
                                    flexWrap: 'wrap',
                                    gap: '0.5rem',
                                }}>
                                    {result.ips.map((ip, idx) => (
                                        <span key={idx} style={{
                                            padding: '0.25rem 0.6rem',
                                            background: 'var(--bg-tertiary)',
                                            borderRadius: 'var(--radius-sm)',
                                            fontFamily: 'monospace',
                                            fontSize: '0.8125rem',
                                            wordBreak: 'break-all',
                                        }}>
                                            {ip}
                                        </span>
                                    ))}
                                </div>
                            </div>
                        )}

                        {result.status && result.status.length > 0 && (
                            <div style={{ gridColumn: '1 / -1' }}>
                                <label style={{ fontSize: '0.875rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '0.5rem' }}>
                                    {t.whois.status}
                                </label>
                                <div style={{
                                    padding: '0.75rem',
                                    background: 'var(--bg-secondary)',
                                    borderRadius: 'var(--radius-sm)',
                                    display: 'flex',
                                    flexWrap: 'wrap',
                                    gap: '0.5rem',
                                }}>
                                    {result.status.map((st, idx) => (
                                        <span key={idx} style={{
                                            padding: '0.25rem 0.6rem',
                                            background: 'var(--bg-tertiary)',
                                            borderRadius: 'var(--radius-sm)',
                                            fontFamily: 'monospace',
                                            fontSize: '0.8125rem',
                                        }}>
                                            {st}
                                        </span>
                                    ))}
                                </div>
                            </div>
                        )}

                        {result.contacts && (
                            <div style={{ gridColumn: '1 / -1', display: 'grid', gap: '1rem', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))' }}>
                                <ContactList t={t} title={`${t.whois.contacts} · ${t.whois.owner}`} contacts={result.contacts.owner} />
                                <ContactList t={t} title={`${t.whois.contacts} · ${t.whois.admin}`} contacts={result.contacts.admin} />
                                <ContactList t={t} title={`${t.whois.contacts} · ${t.whois.tech}`} contacts={result.contacts.tech} />
                            </div>
                        )}

                        {result.raw && (
                            <div style={{ gridColumn: '1 / -1' }}>
                                <button
                                    type="button"
                                    onClick={() => setShowRaw((v) => !v)}
                                    style={{
                                        background: 'none',
                                        border: 'none',
                                        color: 'var(--text-secondary)',
                                        cursor: 'pointer',
                                        fontSize: '0.875rem',
                                        display: 'inline-flex',
                                        alignItems: 'center',
                                        gap: '0.25rem',
                                        padding: 0,
                                    }}
                                >
                                    {showRaw ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
                                    {showRaw ? t.whois.hideRaw : t.whois.showRaw}
                                </button>
                                {showRaw && (
                                    <pre style={{
                                        marginTop: '0.5rem',
                                        padding: '0.75rem',
                                        background: 'var(--bg-secondary)',
                                        borderRadius: 'var(--radius-sm)',
                                        fontFamily: 'monospace',
                                        fontSize: '0.75rem',
                                        overflow: 'auto',
                                        maxHeight: '400px',
                                        whiteSpace: 'pre-wrap',
                                        wordBreak: 'break-all',
                                    }}>
                                        {JSON.stringify(result.raw, null, 2)}
                                    </pre>
                                )}
                            </div>
                        )}
                    </div>
                </div>
            )}

            {/* Collapsible API key configuration */}
            <div className="domain-list-card" style={{ padding: '1.25rem', marginTop: '1.25rem', cursor: 'default' }}>
                <button
                    type="button"
                    onClick={() => setShowConfig((v) => !v)}
                    style={{
                        background: 'none',
                        border: 'none',
                        color: 'var(--text-primary)',
                        cursor: 'pointer',
                        fontSize: '15px',
                        fontWeight: '600',
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: '0.5rem',
                        padding: 0,
                        width: '100%',
                    }}
                >
                    {showConfig ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
                    {t.whois.configExpand}
                </button>

                {showConfig && (
                    <form onSubmit={handleSaveConfig} style={{ marginTop: '1rem' }}>
                        <p style={{ fontSize: '13px', color: 'var(--text-secondary)', marginTop: 0, marginBottom: '1rem' }}>
                            {t.whois.configHint}
                        </p>

                        {configError && (
                            <div style={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: '0.5rem',
                                color: 'var(--danger)',
                                marginBottom: '1rem',
                                padding: '0.75rem 1rem',
                                backgroundColor: 'rgba(255, 0, 0, 0.05)',
                                border: '1px solid rgba(255, 0, 0, 0.15)',
                                borderRadius: 'var(--radius-sm)',
                                fontSize: '14px',
                            }}>
                                <AlertCircle size={16} />
                                {configError}
                            </div>
                        )}

                        {configSuccess && (
                            <div style={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: '0.5rem',
                                color: 'var(--success)',
                                marginBottom: '1rem',
                                padding: '0.75rem 1rem',
                                backgroundColor: 'rgba(16, 185, 129, 0.05)',
                                border: '1px solid rgba(16, 185, 129, 0.15)',
                                borderRadius: 'var(--radius-sm)',
                                fontSize: '14px',
                            }}>
                                <CheckCircle size={16} />
                                {configSuccess}
                            </div>
                        )}

                        <div className="form-group" style={{ marginBottom: '1rem' }}>
                            <label className="form-label" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', cursor: 'pointer', fontSize: '13px', fontWeight: '500', color: 'var(--text-primary)' }}>
                                <input
                                    type="checkbox"
                                    checked={config.enabled}
                                    onChange={(e) => setConfig({ ...config, enabled: e.target.checked })}
                                    style={{ width: 'auto', cursor: 'pointer' }}
                                />
                                {t.whois.configEnabled}
                            </label>
                        </div>

                        <div className="form-group">
                            <label className="form-label">{t.whois.configApiKey}</label>
                            <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'stretch' }}>
                                <input
                                    type={showKey ? 'text' : 'password'}
                                    className="form-input"
                                    value={config.api_key}
                                    onChange={(e) => setConfig({ ...config, api_key: e.target.value })}
                                    placeholder={configured
                                        ? t.common.keepCurrentIfBlank.replace('{field}', t.whois.configApiKey)
                                        : t.whois.configKeyPlaceholder}
                                    autoComplete="off"
                                    style={{ flex: 1 }}
                                />
                                <button
                                    type="button"
                                    onClick={() => setShowKey((v) => !v)}
                                    className="btn btn-secondary"
                                    title={showKey ? t.whois.hideKey : t.whois.previewKey}
                                    aria-label={showKey ? t.whois.hideKey : t.whois.previewKey}
                                    style={{ height: '34px', width: '34px', padding: 0, display: 'inline-flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}
                                >
                                    {showKey ? <EyeOff size={15} /> : <Eye size={15} />}
                                </button>
                            </div>
                            {configured && (
                                <div style={{ fontSize: '12px', color: 'var(--text-tertiary)', marginTop: '0.375rem' }}>
                                    {t.whois.configEditHint}
                                </div>
                            )}
                        </div>

                        <div className="form-actions-row" style={{ display: 'flex', justifyContent: 'flex-end', marginTop: '1.25rem' }}>
                            <button type="submit" className="btn btn-primary" style={{ height: '34px', fontSize: '13px' }} disabled={configSaving}>
                                {configSaving ? <div className="spinner" style={{ width: '1rem', height: '1rem', borderWidth: '2px' }}></div> : t.whois.configSave}
                            </button>
                        </div>
                    </form>
                )}
            </div>
        </div>
    );
};

export default Whois;
