import { useState } from 'react';
import { api } from '../api';
import { useLanguage } from '../LanguageContext';

export default function Profile() {
    const { t } = useLanguage();
    const [updating, setUpdating] = useState(false);
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');
    const [passwordForm, setPasswordForm] = useState({
        old_password: '',
        new_password: '',
        confirm_password: ''
    });

    const handlePasswordChange = async (e) => {
        e.preventDefault();
        setError('');
        setSuccess('');

        if (passwordForm.new_password !== passwordForm.confirm_password) {
            setError(t.common.password.mismatch);
            return;
        }

        if (passwordForm.new_password.length < 6) {
            setError(t.common.password.tooShort);
            return;
        }

        setUpdating(true);
        try {
            await api.updatePassword({
                old_password: passwordForm.old_password,
                new_password: passwordForm.new_password
            });
            setSuccess(t.common.password.updated);
            setPasswordForm({ old_password: '', new_password: '', confirm_password: '' });
        } catch (err) {
            setError(err.message);
        } finally {
            setUpdating(false);
        }
    };

    return (
        <div style={{ maxWidth: '500px', margin: '0 auto' }}>
            <div style={{ marginBottom: '1.5rem' }}>
                <h2 style={{ fontSize: '1.5rem', fontWeight: 'bold', letterSpacing: '-0.02em', margin: 0 }}>
                    {t.common.password.change}
                </h2>
                <p style={{ color: 'var(--text-secondary)', fontSize: '13px', marginTop: '0.25rem', marginBottom: 0 }}>
                    Update your account login password.
                </p>
            </div>

            {error && (
                <div style={{
                    padding: '0.75rem 1rem',
                    marginBottom: '1rem',
                    background: 'rgba(255, 0, 0, 0.05)',
                    border: '1px solid rgba(255, 0, 0, 0.15)',
                    borderRadius: 'var(--radius-sm)',
                    color: 'var(--danger)',
                    fontSize: '14px'
                }}>
                    {error}
                </div>
            )}

            {success && (
                <div style={{
                    padding: '0.75rem 1rem',
                    marginBottom: '1rem',
                    background: 'rgba(16, 185, 129, 0.05)',
                    border: '1px solid rgba(16, 185, 129, 0.15)',
                    borderRadius: 'var(--radius-sm)',
                    color: 'var(--success)',
                    fontSize: '14px'
                }}>
                    {success}
                </div>
            )}

            <div className="domain-list-card" style={{ padding: '1.25rem', cursor: 'default' }}>
                <form onSubmit={handlePasswordChange} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                    <div>
                        <label className="form-label">{t.common.password.current}</label>
                        <input
                            type="password"
                            value={passwordForm.old_password}
                            onChange={(e) => setPasswordForm({ ...passwordForm, old_password: e.target.value })}
                            className="form-input"
                            required
                        />
                    </div>

                    <div>
                        <label className="form-label">{t.common.password.new}</label>
                        <input
                            type="password"
                            value={passwordForm.new_password}
                            onChange={(e) => setPasswordForm({ ...passwordForm, new_password: e.target.value })}
                            className="form-input"
                            required
                            minLength={6}
                        />
                    </div>

                    <div>
                        <label className="form-label">{t.common.password.confirm}</label>
                        <input
                            type="password"
                            value={passwordForm.confirm_password}
                            onChange={(e) => setPasswordForm({ ...passwordForm, confirm_password: e.target.value })}
                            className="form-input"
                            required
                            minLength={6}
                        />
                    </div>

                    <button
                        type="submit"
                        disabled={updating}
                        className="btn btn-primary"
                        style={{ marginTop: '0.5rem', height: '34px', fontSize: '13px' }}
                    >
                        {updating ? t.common.updating : t.common.password.change}
                    </button>
                </form>
            </div>
        </div>
    );
}
