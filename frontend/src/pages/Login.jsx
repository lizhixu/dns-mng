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
            padding: '24px',
            position: 'relative',
            overflow: 'hidden'
        }}>
            {/* 背景装饰 */}
            <div style={{
                position: 'absolute',
                top: '-50%',
                left: '-50%',
                width: '200%',
                height: '200%',
                background: 'radial-gradient(circle at 30% 50%, rgba(0, 112, 243, 0.08) 0%, transparent 50%), radial-gradient(circle at 70% 50%, rgba(99, 102, 241, 0.08) 0%, transparent 50%)',
                animation: 'rotate 30s linear infinite',
                pointerEvents: 'none'
            }} />
            
            {/* 网格背景 */}
            <div style={{
                position: 'absolute',
                top: 0,
                left: 0,
                right: 0,
                bottom: 0,
                backgroundImage: `
                    linear-gradient(var(--border-color) 1px, transparent 1px),
                    linear-gradient(90deg, var(--border-color) 1px, transparent 1px)
                `,
                backgroundSize: '50px 50px',
                opacity: 0.3,
                pointerEvents: 'none'
            }} />

            {/* 右上角工具栏 */}
            <div style={{ 
                position: 'fixed', 
                top: '24px', 
                right: '24px',
                display: 'flex',
                gap: '12px',
                alignItems: 'center',
                zIndex: 10
            }}>
                {/* Theme Switcher */}
                <div style={{ 
                    display: 'flex', 
                    alignItems: 'center', 
                    gap: '2px', 
                    background: 'var(--bg-secondary)', 
                    border: '1px solid var(--border-color)', 
                    borderRadius: 'var(--radius-sm)', 
                    padding: '2px',
                    backdropFilter: 'blur(10px)'
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
                        fontSize: '13px',
                        backdropFilter: 'blur(10px)'
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
                maxWidth: '400px',
                position: 'relative',
                zIndex: 1,
                backdropFilter: 'blur(10px)',
                boxShadow: '0 8px 32px rgba(0, 0, 0, 0.1)'
            }}>
                <div style={{ textAlign: 'center', marginBottom: '32px' }}>
                    <div style={{
                        width: '64px',
                        height: '64px',
                        margin: '0 auto 16px',
                        background: 'linear-gradient(135deg, var(--accent-primary) 0%, rgba(99, 102, 241, 0.8) 100%)',
                        borderRadius: '16px',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        fontSize: '28px',
                        fontWeight: '700',
                        color: '#fff',
                        boxShadow: '0 4px 16px rgba(0, 112, 243, 0.3)'
                    }}>
                        DNS
                    </div>
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
