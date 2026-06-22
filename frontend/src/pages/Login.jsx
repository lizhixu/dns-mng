import { useState } from 'react';
import { useAuth } from '../AuthContext';
import { useLanguage } from '../LanguageContext';
import { useNavigate } from 'react-router-dom';
import { Eye, EyeOff, Github } from 'lucide-react';
import ThemeSwitcher from '../components/ThemeSwitcher';
import LanguageSelect from '../components/LanguageSelect';

const Login = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [showPassword, setShowPassword] = useState(false);
    const { login } = useAuth();
    const { t } = useLanguage();
    const navigate = useNavigate();
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);
    const [touched, setTouched] = useState({ username: false, password: false });

    // 实时验证
    const validateUsername = (value) => {
        if (!value) return t.login.usernamePlaceholder;
        if (value.length < 3) return t.login.usernameTooShort;
        if (value.length > 50) return t.login.usernameTooLong;
        if (!/^[a-zA-Z0-9_-]+$/.test(value)) return t.login.usernameInvalid;
        return '';
    };

    const validatePassword = (value) => {
        if (!value) return t.login.passwordPlaceholder;
        if (value.length < 6) return t.login.passwordTooShort;
        return '';
    };

    const usernameError = touched.username ? validateUsername(username) : '';
    const passwordError = touched.password ? validatePassword(password) : '';
    const isFormValid = username && password && !validateUsername(username) && !validatePassword(password);

    const handleUsernameChange = (e) => {
        setUsername(e.target.value);
        setError('');
    };

    const handlePasswordChange = (e) => {
        setPassword(e.target.value);
        setError('');
    };

    const handleUsernameBlur = () => {
        setTouched({ ...touched, username: true });
    };

    const handlePasswordBlur = () => {
        setTouched({ ...touched, password: true });
    };

    const handleSubmit = async (e) => {
        e.preventDefault();

        setTouched({ username: true, password: true });

        const usernameErr = validateUsername(username);
        const passwordErr = validatePassword(password);

        if (usernameErr || passwordErr) {
            return;
        }

        setLoading(true);
        setError('');

        try {
            await login(username.trim(), password);
            navigate('/domains');
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
                top: '20px',
                right: '20px',
                display: 'flex',
                gap: '8px',
                alignItems: 'center',
                zIndex: 10
            }}>
                {/* GitHub Link */}
                <a
                    href="https://github.com/lizhixu/dns-mng"
                    target="_blank"
                    rel="noopener noreferrer"
                    title={t.login.viewOnGitHub}
                    className="login-toolbar-btn"
                >
                    <Github size={15} />
                </a>

                {/* Theme Switcher */}
                <ThemeSwitcher />

                {/* Language Switcher */}
                <LanguageSelect className="form-input login-lang-select" />
            </div>

            {/* 登录卡片 */}
            <div className="login-card">
                {/* Logo */}
                <div style={{ textAlign: 'center', marginBottom: '36px' }}>
                    <div className="login-logo">
                        <img src="/DNS.svg" alt="DNS Logo" style={{ width: '50px', height: '50px' }} />
                    </div>
                    <h1 style={{
                        fontSize: '22px',
                        fontWeight: '600',
                        marginBottom: '6px',
                        margin: 0,
                        letterSpacing: '-0.02em'
                    }}>
                        {t.layout.title}
                    </h1>
                    <p style={{
                        fontSize: '14px',
                        color: 'var(--text-secondary)',
                        margin: '8px 0 0',
                        fontWeight: '400'
                    }}>
                        {t.login.title}
                    </p>
                </div>

                {/* 错误提示 */}
                {error && (
                    <div className="login-error">
                        {error}
                    </div>
                )}

                {/* 表单 */}
                <form onSubmit={handleSubmit}>
                    <div className="form-group">
                        <label className="form-label">{t.login.username}</label>
                        <input
                            type="text"
                            className={`form-input ${usernameError ? 'input-error' : ''}`}
                            value={username}
                            onChange={handleUsernameChange}
                            onBlur={handleUsernameBlur}
                            autoComplete="username"
                            autoFocus
                            placeholder={t.login.usernamePlaceholder}
                        />
                        {usernameError && (
                            <div className="login-field-error">{usernameError}</div>
                        )}
                    </div>
                    <div className="form-group">
                        <label className="form-label">{t.login.password}</label>
                        <div style={{ position: 'relative' }}>
                            <input
                                type={showPassword ? "text" : "password"}
                                className={`form-input ${passwordError ? 'input-error' : ''}`}
                                value={password}
                                onChange={handlePasswordChange}
                                onBlur={handlePasswordBlur}
                                autoComplete="current-password"
                                placeholder={t.login.passwordPlaceholder}
                                style={{ paddingRight: '44px' }}
                            />
                            <button
                                type="button"
                                onClick={() => setShowPassword(!showPassword)}
                                className="login-password-toggle"
                            >
                                {showPassword ? <EyeOff size={16} /> : <Eye size={16} />}
                            </button>
                        </div>
                        {passwordError && (
                            <div className="login-field-error">{passwordError}</div>
                        )}
                    </div>
                    <button
                        type="submit"
                        className="btn btn-primary login-submit-btn"
                        disabled={loading || !isFormValid}
                    >
                        {loading ? <div className="spinner"></div> : t.login.title}
                    </button>
                </form>
            </div>
        </div>
    );
};

export default Login;
