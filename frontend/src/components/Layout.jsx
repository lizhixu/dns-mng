import { useEffect, useMemo, useState } from 'react';
import { Outlet, Link, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../AuthContext';
import { useLanguage } from '../LanguageContext';
import { useTheme } from '../ThemeContext';
import { api } from '../api';
import { Sun, Moon, Monitor, FileText, Globe, Server, Settings, ChevronDown, X, Github, Menu } from 'lucide-react';
import BackToTop from './BackToTop';
import useMediaQuery from '../hooks/useMediaQuery';

const Layout = () => {
    const { user, logout } = useAuth();
    const navigate = useNavigate();
    const location = useLocation();
    const { themeMode, changeTheme } = useTheme();
    const [showSettings, setShowSettings] = useState(false);
    const [updating, setUpdating] = useState(false);
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');
    const [sidebarOpen, setSidebarOpen] = useState(false);
    const isMobile = useMediaQuery('(max-width: 768px)');
    const [passwordForm, setPasswordForm] = useState({
        old_password: '',
        new_password: '',
        confirm_password: ''
    });

    const { language, changeLanguage, languages, t } = useLanguage();

    useEffect(() => {
        if (!isMobile) {
            setSidebarOpen(false);
        }
    }, [isMobile]);

    useEffect(() => {
        if (isMobile) {
            setSidebarOpen(false);
        }
    }, [isMobile, location.pathname]);

    const handleLogout = () => {
        logout();
        navigate('/login');
    };

    const isActive = (path) => {
        if (location.pathname === path) return true;
        return location.pathname.startsWith(path + '/');
    };

    const menuItemStyle = (path) => ({
        display: 'block',
        padding: '8px',
        marginBottom: '4px',
        borderRadius: 'var(--radius-sm)',
        background: isActive(path) ? 'var(--bg-tertiary)' : 'transparent',
        color: isActive(path) ? 'var(--text-primary)' : 'var(--text-secondary)',
        fontSize: '14px',
        fontWeight: isActive(path) ? '500' : '400',
        transition: 'var(--transition)',
        textDecoration: 'none',
        cursor: 'pointer'
    });

    const handlePasswordChange = async (e) => {
        e.preventDefault();
        setError('');
        setSuccess('');

        if (passwordForm.new_password !== passwordForm.confirm_password) {
            setError(t.layout.passwordMismatch);
            return;
        }

        if (passwordForm.new_password.length < 6) {
            setError(t.layout.passwordTooShort);
            return;
        }

        setUpdating(true);
        try {
            await api.updatePassword({
                old_password: passwordForm.old_password,
                new_password: passwordForm.new_password
            });
            setSuccess(t.layout.passwordUpdated);
            setPasswordForm({ old_password: '', new_password: '', confirm_password: '' });
            setTimeout(() => setShowSettings(false), 1500);
        } catch (err) {
            setError(err.message);
        } finally {
            setUpdating(false);
        }
    };

    const closeSettings = () => {
        setShowSettings(false);
        setError('');
        setSuccess('');
        setPasswordForm({ old_password: '', new_password: '', confirm_password: '' });
    };

    const navigationItems = useMemo(() => ([
        { path: '/domains', icon: Globe, label: t.layout.domains },
        { path: '/accounts', icon: Server, label: t.layout.accounts },
        { path: '/logs', icon: FileText, label: t.layout.logsManagement || t.layout.logs },
        { path: '/email-settings', icon: Settings, label: t.layout.emailNotifications }
    ]), [t]);

    const sidebar = (
        <aside className={`app-sidebar ${isMobile ? 'mobile' : 'desktop'} ${sidebarOpen ? 'open' : ''}`}>
            <div className="app-sidebar-header">
                <h1 style={{
                    fontSize: '16px',
                    fontWeight: '600',
                    padding: '0 8px',
                    margin: 0
                }}>
                    {t.layout.title}
                </h1>
                {isMobile && (
                    <button
                        type="button"
                        onClick={() => setSidebarOpen(false)}
                        className="mobile-sidebar-close"
                        aria-label={t.common.close || 'Close'}
                    >
                        <X size={18} />
                    </button>
                )}
            </div>
            <nav style={{ flex: 1, minHeight: 0 }}>
                {navigationItems.map(({ path, icon: Icon, label }) => (
                    <Link key={path} to={path} style={menuItemStyle(path)}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                            <Icon size={14} />
                            {label}
                        </div>
                    </Link>
                ))}
            </nav>
        </aside>
    );

    return (
        <div className="app-shell" style={{ display: 'flex', height: '100vh', overflow: 'hidden' }}>
            {!isMobile && sidebar}

            {isMobile && (
                <>
                    <div
                        className={`sidebar-overlay ${sidebarOpen ? 'visible' : ''}`}
                        onClick={() => setSidebarOpen(false)}
                    />
                    {sidebar}
                </>
            )}

            <main style={{
                flex: 1,
                display: 'flex',
                flexDirection: 'column',
                overflow: 'hidden',
                minWidth: 0
            }}>
                <header className="app-header" style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    padding: isMobile ? '0 16px' : '0 24px',
                    height: '64px',
                    background: 'var(--bg-primary)',
                    borderBottom: '1px solid var(--border-color)',
                    flexShrink: 0,
                    gap: '12px'
                }}>
                    <div className="app-header-left" style={{ display: 'flex', alignItems: 'center', gap: '12px', minWidth: 0, flex: 1 }}>
                        {isMobile && (
                            <button
                                type="button"
                                onClick={() => setSidebarOpen(true)}
                                className="mobile-menu-btn"
                                aria-label={t.layout.title}
                            >
                                <Menu size={18} />
                            </button>
                        )}
                        {isMobile && (
                            <div className="app-header-title" style={{ minWidth: 0, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                                <span style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)', lineHeight: 1.1 }}>
                                    DNS Manager
                                </span>
                                <span style={{ fontSize: '15px', fontWeight: 600, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                                    {t.layout.title}
                                </span>
                            </div>
                        )}
                    </div>
                    <div className="app-header-actions" style={{ display: 'flex', alignItems: 'center', gap: isMobile ? '8px' : '12px', minWidth: 0, flexShrink: 0 }}>
                        <div className="header-theme-switcher" style={{ display: 'flex', alignItems: 'center', gap: '2px', background: 'var(--bg-secondary)', border: '1px solid var(--border-color)', borderRadius: 'var(--radius-sm)', padding: '2px', flexShrink: 0 }}>
                            <button
                                onClick={() => changeTheme('light')}
                                title={t.layout.themeLight}
                                style={{
                                    padding: '6px',
                                    borderRadius: '3px',
                                    background: themeMode === 'light' ? 'var(--bg-primary)' : 'transparent',
                                    color: themeMode === 'light' ? 'var(--text-primary)' : 'var(--text-tertiary)',
                                    border: themeMode === 'light' ? '1px solid var(--border-color)' : '1px solid transparent',
                                    transition: 'var(--transition)',
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'center'
                                }}
                            >
                                <Sun size={14} />
                            </button>
                            <button
                                onClick={() => changeTheme('system')}
                                title={t.layout.themeSystem}
                                style={{
                                    padding: '6px',
                                    borderRadius: '3px',
                                    background: themeMode === 'system' ? 'var(--bg-primary)' : 'transparent',
                                    color: themeMode === 'system' ? 'var(--text-primary)' : 'var(--text-tertiary)',
                                    border: themeMode === 'system' ? '1px solid var(--border-color)' : '1px solid transparent',
                                    transition: 'var(--transition)',
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'center'
                                }}
                            >
                                <Monitor size={14} />
                            </button>
                            <button
                                onClick={() => changeTheme('dark')}
                                title={t.layout.themeDark}
                                style={{
                                    padding: '6px',
                                    borderRadius: '3px',
                                    background: themeMode === 'dark' ? 'var(--bg-primary)' : 'transparent',
                                    color: themeMode === 'dark' ? 'var(--text-primary)' : 'var(--text-tertiary)',
                                    border: themeMode === 'dark' ? '1px solid var(--border-color)' : '1px solid transparent',
                                    transition: 'var(--transition)',
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'center'
                                }}
                            >
                                <Moon size={14} />
                            </button>
                        </div>

                        <select
                            value={language}
                            onChange={(e) => changeLanguage(e.target.value)}
                            className={`form-input header-language-select ${isMobile ? 'mobile' : 'desktop'}`}
                            style={{
                                width: isMobile ? '72px' : 'auto',
                                height: '32px',
                                padding: isMobile ? '0 6px' : '0 8px',
                                fontSize: '13px'
                            }}
                        >
                            {Object.entries(languages).map(([code, lang]) => (
                                <option key={code} value={code}>{isMobile ? code.toUpperCase() : lang.name}</option>
                            ))}
                        </select>

                        {!isMobile && (
                            <a
                                href="https://github.com/lizhixu/dns-mng"
                                target="_blank"
                                rel="noopener noreferrer"
                                title="GitHub"
                                style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'center',
                                    padding: '6px',
                                    background: 'var(--bg-secondary)',
                                    border: '1px solid var(--border-color)',
                                    borderRadius: 'var(--radius-sm)',
                                    color: 'var(--text-secondary)',
                                    transition: 'var(--transition)',
                                    textDecoration: 'none'
                                }}
                                onMouseEnter={(e) => {
                                    e.currentTarget.style.color = 'var(--text-primary)';
                                    e.currentTarget.style.borderColor = 'var(--text-tertiary)';
                                }}
                                onMouseLeave={(e) => {
                                    e.currentTarget.style.color = 'var(--text-secondary)';
                                    e.currentTarget.style.borderColor = 'var(--border-color)';
                                }}
                            >
                                <Github size={16} />
                            </a>
                        )}

                        {!isMobile && (
                            <div style={{
                                height: '20px',
                                width: '1px',
                                background: 'var(--border-color)'
                            }}></div>
                        )}

                        <div style={{ position: 'relative', minWidth: 0 }}>
                            <button
                                className="header-user-btn"
                                onClick={() => setShowSettings(!showSettings)}
                                style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: isMobile ? '0.35rem' : '0.5rem',
                                    padding: isMobile ? '6px 8px' : '6px 10px',
                                    background: 'var(--bg-secondary)',
                                    border: '1px solid var(--border-color)',
                                    borderRadius: 'var(--radius-sm)',
                                    color: 'var(--text-primary)',
                                    fontSize: '14px',
                                    cursor: 'pointer',
                                    transition: 'var(--transition)',
                                    maxWidth: isMobile ? '120px' : 'none'
                                }}
                            >
                                <Settings size={14} />
                                <span className="header-user-name" style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                                    {user?.username}
                                </span>
                                <ChevronDown size={14} style={{
                                    transform: showSettings ? 'rotate(180deg)' : 'rotate(0deg)',
                                    transition: 'transform 0.2s',
                                    flexShrink: 0
                                }} />
                            </button>

                            {showSettings && (
                                <>
                                    <div
                                        style={{
                                            position: 'fixed',
                                            top: 0,
                                            left: 0,
                                            right: 0,
                                            bottom: 0,
                                            zIndex: 40
                                        }}
                                        onClick={closeSettings}
                                    />
                                    <div className="settings-dropdown" style={{
                                        position: 'absolute',
                                        top: '100%',
                                        right: 0,
                                        marginTop: '8px',
                                        width: isMobile ? 'min(320px, calc(100vw - 32px))' : '320px',
                                        maxWidth: 'calc(100vw - 32px)',
                                        background: 'var(--bg-primary)',
                                        border: '1px solid var(--border-color)',
                                        borderRadius: 'var(--radius-md)',
                                        boxShadow: 'var(--shadow-lg)',
                                        zIndex: 60,
                                        overflow: 'hidden'
                                    }}>
                                        <div style={{
                                            padding: '16px',
                                            borderBottom: '1px solid var(--border-color)',
                                            display: 'flex',
                                            justifyContent: 'space-between',
                                            alignItems: 'center'
                                        }}>
                                            <span style={{ fontWeight: '500', fontSize: '14px' }}>{t.layout.settings}</span>
                                            <button
                                                onClick={closeSettings}
                                                style={{
                                                    padding: '4px',
                                                    borderRadius: '4px',
                                                    color: 'var(--text-tertiary)',
                                                    cursor: 'pointer'
                                                }}
                                            >
                                                <X size={16} />
                                            </button>
                                        </div>

                                        <div style={{ padding: '16px' }}>
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
                                                    <label className="form-label" style={{ fontSize: '13px' }}>{t.layout.currentPassword}</label>
                                                    <input
                                                        type="password"
                                                        value={passwordForm.old_password}
                                                        onChange={(e) => setPasswordForm({ ...passwordForm, old_password: e.target.value })}
                                                        className="form-input"
                                                        style={{ height: '36px', fontSize: '13px' }}
                                                        required
                                                    />
                                                </div>

                                                <div>
                                                    <label className="form-label" style={{ fontSize: '13px' }}>{t.layout.newPassword}</label>
                                                    <input
                                                        type="password"
                                                        value={passwordForm.new_password}
                                                        onChange={(e) => setPasswordForm({ ...passwordForm, new_password: e.target.value })}
                                                        className="form-input"
                                                        style={{ height: '36px', fontSize: '13px' }}
                                                        required
                                                        minLength={6}
                                                    />
                                                </div>

                                                <div>
                                                    <label className="form-label" style={{ fontSize: '13px' }}>{t.layout.confirmPassword}</label>
                                                    <input
                                                        type="password"
                                                        value={passwordForm.confirm_password}
                                                        onChange={(e) => setPasswordForm({ ...passwordForm, confirm_password: e.target.value })}
                                                        className="form-input"
                                                        style={{ height: '36px', fontSize: '13px' }}
                                                        required
                                                        minLength={6}
                                                    />
                                                </div>

                                                <button
                                                    type="submit"
                                                    disabled={updating}
                                                    className="btn btn-primary"
                                                    style={{ marginTop: '0.5rem', height: '36px' }}
                                                >
                                                    {updating ? t.layout.updating : t.layout.changePassword}
                                                </button>
                                            </form>
                                        </div>

                                        <div style={{
                                            padding: '12px 16px',
                                            borderTop: '1px solid var(--border-color)',
                                            background: 'var(--bg-secondary)'
                                        }}>
                                            <button
                                                onClick={handleLogout}
                                                style={{
                                                    width: '100%',
                                                    padding: '8px 12px',
                                                    background: 'transparent',
                                                    border: '1px solid var(--border-color)',
                                                    borderRadius: 'var(--radius-sm)',
                                                    color: 'var(--text-secondary)',
                                                    fontSize: '13px',
                                                    cursor: 'pointer',
                                                    transition: 'var(--transition)'
                                                }}
                                            >
                                                {t.layout.logout}
                                            </button>
                                        </div>
                                    </div>
                                </>
                            )}
                        </div>
                    </div>
                </header>

                <div className="app-content" style={{
                    flex: 1,
                    overflow: 'auto',
                    padding: isMobile ? '16px' : '24px',
                    background: 'var(--bg-primary)'
                }}>
                    <Outlet />
                </div>
            </main>
            <BackToTop />
        </div>
    );
};

export default Layout;
