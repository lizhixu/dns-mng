import { useState, useEffect, useRef } from 'react';
import { api } from '../api';
import { Mail, Send, AlertCircle, CheckCircle } from 'lucide-react';
import { useLanguage } from '../LanguageContext';

const EmailSettings = () => {
    const { t } = useLanguage();
    const [config, setConfig] = useState({
        smtp_host: '',
        smtp_port: 587,
        smtp_username: '',
        smtp_password: '',
        from_email: '',
        from_name: '',
        to_email: '',
        language: 'zh',
        enabled: true
    });
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [testing, setTesting] = useState(false);
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');
    const fetchedRef = useRef(false);

    useEffect(() => {
        // Prevent duplicate requests in React StrictMode
        if (fetchedRef.current) return;
        fetchedRef.current = true;
        loadConfig();
    }, []);

    const loadConfig = async () => {
        setLoading(true);
        try {
            const data = await api.getEmailConfig();
            if (data && data.id) {
                // Config exists, load it
                setConfig({
                    smtp_host: data.smtp_host || '',
                    smtp_port: data.smtp_port || 587,
                    smtp_username: data.smtp_username || '',
                    smtp_password: '', // Don't show password
                    from_email: data.from_email || '',
                    from_name: data.from_name || '',
                    to_email: data.to_email || '',
                    language: data.language || 'zh',
                    enabled: data.enabled !== undefined ? data.enabled : true
                });
            } else {
                // No config yet, use defaults
                setConfig({
                    smtp_host: '',
                    smtp_port: 587,
                    smtp_username: '',
                    smtp_password: '',
                    from_email: '',
                    from_name: 'DNS Manager',
                    to_email: '',
                    language: 'zh',
                    enabled: true
                });
            }
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setSaving(true);
        setError('');
        setSuccess('');

        try {
            await api.updateEmailConfig(config);
            setSuccess(t.emailSettings.configSaved);
            setTimeout(() => setSuccess(''), 3000);
        } catch (err) {
            setError(err.message);
        } finally {
            setSaving(false);
        }
    };

    const handleTest = async () => {
        setTesting(true);
        setError('');
        setSuccess('');

        try {
            await api.testEmailConfig();
            setSuccess(t.emailSettings.testSent.replace('{email}', config.to_email));
            setTimeout(() => setSuccess(''), 5000);
        } catch (err) {
            setError(err.message);
        } finally {
            setTesting(false);
        }
    };

    if (loading) {
        return (
            <div style={{ textAlign: 'center', padding: '4rem' }}>
                <div className="spinner" style={{ margin: '0 auto' }}></div>
            </div>
        );
    }

    return (
        <div>
            <div style={{ marginBottom: '2rem' }}>
                <h2 style={{ fontSize: '1.5rem', fontWeight: 'bold', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                    <Mail size={24} />
                    {t.emailSettings.title}
                </h2>
                <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', marginTop: '0.5rem' }}>
                    {t.emailSettings.subtitle}
                </p>
            </div>

            {error && (
                <div style={{ 
                    display: 'flex', 
                    alignItems: 'center', 
                    gap: '0.5rem',
                    color: 'var(--danger)', 
                    marginBottom: '1rem', 
                    padding: '1rem', 
                    backgroundColor: 'rgba(239, 68, 68, 0.1)', 
                    borderRadius: 'var(--radius-md)' 
                }}>
                    <AlertCircle size={18} />
                    {error}
                </div>
            )}

            {success && (
                <div style={{ 
                    display: 'flex', 
                    alignItems: 'center', 
                    gap: '0.5rem',
                    color: 'var(--success)', 
                    marginBottom: '1rem', 
                    padding: '1rem', 
                    backgroundColor: 'rgba(16, 185, 129, 0.1)', 
                    borderRadius: 'var(--radius-md)' 
                }}>
                    <CheckCircle size={18} />
                    {success}
                </div>
            )}

            <div className="glass-panel" style={{ padding: '1.5rem' }}>
                <form onSubmit={handleSubmit}>
                    <div className="form-group">
                        <label className="form-label" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                            <input
                                type="checkbox"
                                checked={config.enabled}
                                onChange={e => setConfig({ ...config, enabled: e.target.checked })}
                                style={{ width: 'auto' }}
                            />
                            {t.emailSettings.enabled}
                        </label>
                    </div>

                    {config.enabled && (
                        <>
                            <div className="email-smtp-row" style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: '1rem' }}>
                                <div className="form-group">
                                    <label className="form-label">{t.emailSettings.smtpServer}</label>
                                    <input
                                        type="text"
                                        className="form-input"
                                        value={config.smtp_host}
                                        onChange={e => setConfig({ ...config, smtp_host: e.target.value })}
                                        placeholder="smtp.gmail.com"
                                        required
                                    />
                                </div>
                                <div className="form-group">
                                    <label className="form-label">{t.emailSettings.port}</label>
                                    <input
                                        type="number"
                                        className="form-input"
                                        value={config.smtp_port}
                                        onChange={e => setConfig({ ...config, smtp_port: parseInt(e.target.value) || 587 })}
                                        placeholder="587"
                                        required
                                    />
                                </div>
                            </div>

                            <div className="form-group">
                                <label className="form-label">{t.emailSettings.username}</label>
                                <input
                                    type="text"
                                    className="form-input"
                                    value={config.smtp_username}
                                    onChange={e => setConfig({ ...config, smtp_username: e.target.value })}
                                    placeholder="your-email@example.com"
                                    required
                                />
                            </div>

                            <div className="form-group">
                                <label className="form-label">{t.emailSettings.password}</label>
                                <input
                                    type="password"
                                    className="form-input"
                                    value={config.smtp_password}
                                    onChange={e => setConfig({ ...config, smtp_password: e.target.value })}
                                    placeholder={t.emailSettings.passwordKeepBlank}
                                />
                            </div>

                            <div className="form-group">
                                <label className="form-label">{t.emailSettings.fromEmail}</label>
                                <input
                                    type="email"
                                    className="form-input"
                                    value={config.from_email}
                                    onChange={e => setConfig({ ...config, from_email: e.target.value })}
                                    placeholder="noreply@example.com"
                                    required
                                />
                            </div>

                            <div className="form-group">
                                <label className="form-label">{t.emailSettings.fromName}</label>
                                <input
                                    type="text"
                                    className="form-input"
                                    value={config.from_name}
                                    onChange={e => setConfig({ ...config, from_name: e.target.value })}
                                    placeholder="DNS Manager"
                                />
                            </div>

                            <div className="form-group">
                                <label className="form-label">{t.emailSettings.toEmail}</label>
                                <input
                                    type="email"
                                    className="form-input"
                                    value={config.to_email}
                                    onChange={e => setConfig({ ...config, to_email: e.target.value })}
                                    placeholder="your-email@example.com"
                                    required
                                />
                                <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)', marginTop: '0.25rem' }}>
                                    {t.emailSettings.toEmailHint}
                                </div>
                            </div>

                            <div className="form-group">
                                <label className="form-label">{t.emailSettings.language}</label>
                                <select
                                    className="form-input"
                                    value={config.language || 'zh'}
                                    onChange={e => setConfig({ ...config, language: e.target.value })}
                                >
                                    <option value="zh">{t.emailSettings.languageZh}</option>
                                    <option value="en">{t.emailSettings.languageEn}</option>
                                </select>
                                <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)', marginTop: '0.25rem' }}>
                                    {t.emailSettings.languageHint}
                                </div>
                            </div>

                            <div style={{ 
                                backgroundColor: 'var(--bg-secondary)', 
                                padding: '1rem', 
                                borderRadius: 'var(--radius-md)', 
                                marginTop: '1rem',
                                fontSize: '0.875rem',
                                color: 'var(--text-secondary)'
                            }}>
                                <strong>{t.emailSettings.commonSmtp}</strong>
                                <ul style={{ marginTop: '0.5rem', marginLeft: '1.5rem' }}>
                                    <li>{t.emailSettings.providers.gmail}</li>
                                    <li>{t.emailSettings.providers.outlook}</li>
                                    <li>{t.emailSettings.providers.qq}</li>
                                    <li>{t.emailSettings.providers.mail163}</li>
                                </ul>
                            </div>
                        </>
                    )}

                    <div className="form-actions-row" style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '2rem' }}>
                        <button type="submit" className="btn btn-primary" disabled={saving}>
                            {saving ? <div className="spinner" style={{ width: '1rem', height: '1rem', borderWidth: '2px' }}></div> : t.emailSettings.saveConfig}
                        </button>
                    </div>
                </form>
            </div>

            {config.enabled && config.smtp_host && config.to_email && (
                <div className="glass-panel" style={{ padding: '1.5rem', marginTop: '1.5rem' }}>
                    <h3 style={{ fontSize: '1.125rem', fontWeight: '600', marginBottom: '1rem' }}>{t.emailSettings.testTitle}</h3>
                    <p style={{ fontSize: '0.875rem', color: 'var(--text-secondary)', marginBottom: '1rem' }}>
                        {t.emailSettings.testDesc.replace('{email}', config.to_email).split(config.to_email).map((part, idx, arr) => (
                            <span key={idx}>
                                {part}
                                {idx < arr.length - 1 ? <strong>{config.to_email}</strong> : null}
                            </span>
                        ))}
                    </p>
                    <button 
                        type="button" 
                        onClick={handleTest} 
                        className="btn btn-secondary"
                        disabled={testing}
                    >
                        {testing ? (
                            <div className="spinner" style={{ width: '1rem', height: '1rem', borderWidth: '2px' }}></div>
                        ) : (
                            <>
                                <Send size={16} />
                                {t.emailSettings.sendTest}
                            </>
                        )}
                    </button>
                </div>
            )}
        </div>
    );
};

export default EmailSettings;
