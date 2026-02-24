import { useState } from 'react';
import { useAuth } from '../AuthContext';
import { useLanguage } from '../LanguageContext';
import { useTheme } from '../ThemeContext';
import { useNavigate } from 'react-router-dom';
import { Sun, Moon, Monitor } from 'lucide-react';

const Login = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const { login } = useAuth();
    const { t, language, changeLanguage, languages } = useLanguage();
    const { themeMode, changeTheme } = useTheme();
    const navigate = useNavigate();
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);

    const handleSubmit = async (e) => {
        e.preventDefault();
        setLoading(true);
        setError('');
        
        try {
            await login(username, password);
            navigate('/domains'); // 登录后跳转到域名列表
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div style={{ 
            display: 'flex', 
            justifyContent: 'center', 
            alignItems: 'center', 
            minHeight: '100vh',
            background: 'var(--bg-primary)',
            padding: '24px'
        }}>
            {/* 右上角工具栏 */}
            <div style={{ 
                position: 'fixed', 
                top: '24px', 
                right: '24px',
                display: 'flex',
                gap: '12px',
                alignItems: 'center'
            }}>
                {/* Theme Switcher */}
                <div style={{ 
                    display: 'flex', 
                    alignItems: 'center', 
                    gap: '2px', 
                    background: 'var(--bg-secondary)', 
                    border: '1px solid var(--border-color)', 
                    borderRadius: 'var(--radius-sm)', 
                    padding: '2px' 
                }}>
                    <button
                        onClick={() => changeTheme('light')}
                        title="Light Mode"
                        style={{
                            padding: '6px',
                            borderRadius: '3px',
                            background: themeMode === 'light' ? 'var(--bg-primary)' : 'transparent',
                            color: themeMode === 'light' ? 'var(--text-primary)' : 'var(--text-tertiary)',
                            border: themeMode === 'light' ? '1px solid var(--border-color)' : '1px solid transparent',
                            transition: 'var(--transition)',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            cursor: 'pointer'
                        }}
                    >
                        <Sun size={14} />
                    </button>
                    <button
                        onClick={() => changeTheme('system')}
                        title="System Theme"
                        style={{
                            padding: '6px',
                            borderRadius: '3px',
                            background: themeMode === 'system' ? 'var(--bg-primary)' : 'transparent',
                            color: themeMode === 'system' ? 'var(--text-primary)' : 'var(--text-tertiary)',
                            border: themeMode === 'system' ? '1px solid var(--border-color)' : '1px solid transparent',
                            transition: 'var(--transition)',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            cursor: 'pointer'
                        }}
                    >
                        <Monitor size={14} />
                    </button>
                    <button
                        onClick={() => changeTheme('dark')}
                        title="Dark Mode"
                        style={{
                            padding: '6px',
                            borderRadius: '3px',
                            background: themeMode === 'dark' ? 'var(--bg-primary)' : 'transparent',
                            color: themeMode === 'dark' ? 'var(--text-primary)' : 'var(--text-tertiary)',
                            border: themeMode === 'dark' ? '1px solid var(--border-color)' : '1px solid transparent',
                            transition: 'var(--transition)',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            cursor: 'pointer'
                        }}
                    >
                        <Moon size={14} />
                    </button>
                </div>

                {/* Language Switcher */}
                <select
                    value={language}
                    onChange={(e) => changeLanguage(e.target.value)}
                    className="form-input"
                    style={{
                        width: 'auto',
                        height: '32px',
                        padding: '0 8px',
                        fontSize: '13px'
                    }}
                >
                    {Object.entries(languages).map(([code, lang]) => (
                        <option key={code} value={code}>{lang.name}</option>
                    ))}
                </select>
            </div>

            <div className="glass-panel" style={{ 
                padding: '40px', 
                width: '100%', 
                maxWidth: '400px'
            }}>
                <div style={{ textAlign: 'center', marginBottom: '32px' }}>
                    <h1 style={{ 
                        fontSize: '24px', 
                        fontWeight: '600', 
                        marginBottom: '8px',
                        margin: 0
                    }}>
                        {t.layout.title}
                    </h1>
                    <p style={{ 
                        fontSize: '14px', 
                        color: 'var(--text-secondary)',
                        margin: 0
                    }}>
                        {t.login.title}
                    </p>
                </div>

                {error && (
                    <div style={{ 
                        color: 'var(--danger)', 
                        marginBottom: '20px',
                        padding: '12px',
                        background: 'rgba(238, 0, 0, 0.1)',
                        border: '1px solid rgba(238, 0, 0, 0.2)',
                        borderRadius: 'var(--radius-sm)',
                        fontSize: '13px'
                    }}>
                        {error}
                    </div>
                )}

                <form onSubmit={handleSubmit}>
                    <div className="form-group">
                        <label className="form-label">{t.login.username}</label>
                        <input
                            type="text"
                            className="form-input"
                            value={username}
                            onChange={(e) => setUsername(e.target.value)}
                            autoComplete="username"
                            autoFocus
                            required
                        />
                    </div>
                    <div className="form-group">
                        <label className="form-label">{t.login.password}</label>
                        <input
                            type="password"
                            className="form-input"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            autoComplete="current-password"
                            required
                        />
                    </div>
                    <button 
                        type="submit" 
                        className="btn btn-primary" 
                        style={{ width: '100%', marginTop: '8px' }}
                        disabled={loading}
                    >
                        {loading ? <div className="spinner"></div> : t.login.loginBtn}
                    </button>
                </form>
            </div>
        </div>
    );
};

export default Login;
