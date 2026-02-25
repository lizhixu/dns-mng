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
            setError(t.profile.passwords_not_match);
            return;
        }

        if (passwordForm.new_password.length < 6) {
            setError(t.profile.password_min_length);
            return;
        }

        setUpdating(true);
        try {
            await api.updatePassword({
                old_password: passwordForm.old_password,
                new_password: passwordForm.new_password
            });
            setSuccess(t.profile.password_updated);
            setPasswordForm({ old_password: '', new_password: '', confirm_password: '' });
        } catch (err) {
            setError(err.message);
        } finally {
            setUpdating(false);
        }
    };

    return (
        <div style={{ maxWidth: '500px', margin: '0 auto' }}>
            <h1 style={{ fontSize: '1.5rem', fontWeight: '600', marginBottom: '1.5rem' }}>
                {t.profile.change_password}
            </h1>

            {error && (
                <div style={{
                    padding: '0.75rem',
                    marginBottom: '1rem',
                    background: 'rgba(239, 68, 68, 0.1)',
                    border: '1px solid rgba(239, 68, 68, 0.2)',
                    borderRadius: 'var(--radius-sm)',
                    color: 'var(--danger)',
                    fontSize: '0.875rem'
                }}>
                    {error}
                </div>
            )}

            {success && (
                <div style={{
                    padding: '0.75rem',
                    marginBottom: '1rem',
                    background: 'rgba(0, 112, 243, 0.1)',
                    border: '1px solid rgba(0, 112, 243, 0.2)',
                    borderRadius: 'var(--radius-sm)',
                    color: 'var(--success)',
                    fontSize: '0.875rem'
                }}>
                    {success}
                </div>
            )}

            <form onSubmit={handlePasswordChange} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                <div>
                    <label className="form-label">{t.profile.old_password}</label>
                    <input
                        type="password"
                        value={passwordForm.old_password}
                        onChange={(e) => setPasswordForm({ ...passwordForm, old_password: e.target.value })}
                        className="form-input"
                        required
                    />
                </div>

                <div>
                    <label className="form-label">{t.profile.new_password}</label>
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
                    <label className="form-label">{t.profile.confirm_password}</label>
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
                    style={{ marginTop: '0.5rem' }}
                >
                    {updating ? t.profile.updating : t.profile.update_password}
                </button>
            </form>
        </div>
    );
}
