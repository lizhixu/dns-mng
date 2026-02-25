import { useState } from 'react';
import { Lock } from 'lucide-react';
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
        <div style={{ maxWidth: '600px', margin: '0 auto' }}>
            <h1 style={{ fontSize: '24px', fontWeight: '600', color: 'var(--text-primary)', marginBottom: '24px' }}>
                {t.profile.user_profile}
            </h1>

            {/* Change Password Card */}
            <div style={{
                background: 'var(--bg-secondary)',
                border: '1px solid var(--border-color)',
                borderRadius: 'var(--radius-md)',
                padding: '24px'
            }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '20px' }}>
                    <Lock size={20} style={{ color: 'var(--text-secondary)' }} />
                    <h2 style={{ fontSize: '18px', fontWeight: '500', color: 'var(--text-primary)' }}>
                        {t.profile.change_password}
                    </h2>
                </div>

                {error && (
                    <div style={{
                        padding: '12px',
                        marginBottom: '16px',
                        background: 'rgba(238, 0, 0, 0.1)',
                        border: '1px solid var(--danger)',
                        borderRadius: 'var(--radius-sm)',
                        color: 'var(--danger)',
                        fontSize: '14px'
                    }}>
                        {error}
                    </div>
                )}

                {success && (
                    <div style={{
                        padding: '12px',
                        marginBottom: '16px',
                        background: 'rgba(0, 112, 243, 0.1)',
                        border: '1px solid var(--success)',
                        borderRadius: 'var(--radius-sm)',
                        color: 'var(--success)',
                        fontSize: '14px'
                    }}>
                        {success}
                    </div>
                )}

                <form onSubmit={handlePasswordChange} style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                    <div>
                        <label style={{ display: 'block', fontSize: '13px', fontWeight: '500', color: 'var(--text-primary)', marginBottom: '6px' }}>
                            {t.profile.old_password}
                        </label>
                        <input
                            type="password"
                            value={passwordForm.old_password}
                            onChange={(e) => setPasswordForm({ ...passwordForm, old_password: e.target.value })}
                            className="form-input"
                            required
                        />
                    </div>
                    <div>
                        <label style={{ display: 'block', fontSize: '13px', fontWeight: '500', color: 'var(--text-primary)', marginBottom: '6px' }}>
                            {t.profile.new_password}
                        </label>
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
                        <label style={{ display: 'block', fontSize: '13px', fontWeight: '500', color: 'var(--text-primary)', marginBottom: '6px' }}>
                            {t.profile.confirm_password}
                        </label>
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
                        style={{ marginTop: '8px' }}
                    >
                        {updating ? t.profile.updating : t.profile.update_password}
                    </button>
                </form>
            </div>
        </div>
    );
}
