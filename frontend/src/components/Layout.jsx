import { useEffect, useMemo, useRef, useState } from 'react';
import { Outlet, Link, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../AuthContext';
import { useLanguage } from '../LanguageContext';
import { api } from '../api';
import { FileText, Globe, Server, Settings, ChevronDown, X, Github, Menu, DatabaseBackup } from 'lucide-react';
import ThemeSwitcher from './ThemeSwitcher';
import LanguageSelect from './LanguageSelect';
import BackToTop from './BackToTop';
import useMediaQuery from '../hooks/useMediaQuery';

const Layout = () => {
    const { user, logout } = useAuth();
    const navigate = useNavigate();
    const location = useLocation();

    const [showSettings, setShowSettings] = useState(false);
    const [updating, setUpdating] = useState(false);
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');
    const [sidebarOpen, setSidebarOpen] = useState(false);
    const isMobile = useMediaQuery('(max-width: 768px)');
    const settingsRef = useRef(null);
    const [passwordForm, setPasswordForm] = useState({
        old_password: '',
        new_password: '',
        confirm_password: ''
    });

    const { t } = useLanguage();

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

    useEffect(() => {
        if (!showSettings) {
            return undefined;
        }

        const handlePointerDown = (event) => {
            if (settingsRef.current && !settingsRef.current.contains(event.target)) {
                closeSettings();
            }
        };

        document.addEventListener('mousedown', handlePointerDown);
        document.addEventListener('touchstart', handlePointerDown);

        return () => {
            document.removeEventListener('mousedown', handlePointerDown);
            document.removeEventListener('touchstart', handlePointerDown);
        };
    }, [showSettings]);

    const handleLogout = () => {
        logout();
        navigate('/login');
    };

    const isActive = (path) => {
        if (location.pathname === path) return true;
        return location.pathname.startsWith(path + '/');
    };

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
            setTimeout(() => resetAndCloseSettings(), 1500);
        } catch (err) {
            setError(err.message);
        } finally {
            setUpdating(false);
        }
    };

    const closeSettings = () => {
        setShowSettings(false);
    };

    const resetAndCloseSettings = () => {
        setShowSettings(false);
        setError('');
        setSuccess('');
        setPasswordForm({ old_password: '', new_password: '', confirm_password: '' });
    };

    const navigationItems = useMemo(() => ([
        { path: '/domains', icon: Globe, label: t.layout.domains },
        { path: '/accounts', icon: Server, label: t.accounts.title },
        { path: '/logs', icon: FileText, label: t.layout.logsManagement },
        { path: '/email-settings', icon: Settings, label: t.layout.emailNotifications },
        { path: '/backup', icon: DatabaseBackup, label: t.backup.title }
    ]), [t]);

    const sidebar = (
        <aside className={`app-sidebar ${isMobile ? 'mobile' : 'desktop'} ${sidebarOpen ? 'open' : ''}`}>
            <div className="app-sidebar-header" style={{ paddingLeft: '12px', paddingRight: '12px', marginBottom: '28px' }}>
                <span className="app-sidebar-title" style={{ 
                    fontSize: '16px', 
                    fontWeight: '700', 
                    letterSpacing: '-0.02em',
                    color: 'var(--text-primary)'
                }}>
                    {t.layout.title}
                </span>
                {isMobile && (
                    <button
                        type="button"
                        onClick={() => setSidebarOpen(false)}
                        className="mobile-sidebar-close"
                        aria-label={t.common.close}
                    >
                        <X size={16} />
                    </button>
                )}
            </div>
            <nav className="sidebar-nav">
                {navigationItems.map(({ path, icon: Icon, label }) => (
                    <Link key={path} to={path} className={`nav-link ${isActive(path) ? 'active' : ''}`}>
                        <Icon size={15} />
                        <span>{label}</span>
                    </Link>
                ))}
            </nav>
            {isMobile && (
                <div className="sidebar-footer">
                    <div className="sidebar-footer-row">
                        <span className="sidebar-footer-label">{t.layout.theme}</span>
                        <ThemeSwitcher style={{ margin: 0 }} />
                    </div>
                    <div className="sidebar-footer-row">
                        <span className="sidebar-footer-label">{t.layout.language}</span>
                        <LanguageSelect style={{ height: '30px' }} />
                    </div>
                </div>
            )}
        </aside>
    );

    return (
        <div className="app-shell">
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

            <main className="app-main">
                <header className="app-header">
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
                                <span style={{ fontSize: '0.7rem', color: 'var(--text-tertiary)', lineHeight: 1.1 }}>
                                    DNS Manager
                                </span>
                                <span style={{ fontSize: '14px', fontWeight: 600, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                                    {t.layout.title}
                                </span>
                            </div>
                        )}
                    </div>
                    <div className="app-header-actions" style={{ display: 'flex', alignItems: 'center', gap: '12px', minWidth: 0, flexShrink: 0 }}>
                        {/* Theme Switcher */}
                        {!isMobile && <ThemeSwitcher />}

                        {/* Language Select */}
                        {!isMobile && <LanguageSelect />}

                        {!isMobile && (
                            <a
                                href="https://github.com/lizhixu/dns-mng"
                                target="_blank"
                                rel="noopener noreferrer"
                                title="GitHub"
                                className="login-toolbar-btn"
                                style={{ width: '32px', height: '32px' }}
                            >
                                <Github size={14} />
                            </a>
                        )}

                        {!isMobile && (
                            <div style={{
                                height: '16px',
                                width: '1px',
                                background: 'var(--border-color)'
                            }}></div>
                        )}

                        {/* User Settings Dropdown */}
                        <div ref={settingsRef} style={{ position: 'relative', minWidth: 0 }}>
                            <button
                                className="header-user-btn"
                                onClick={() => setShowSettings(!showSettings)}
                                style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: '6px',
                                    padding: '6px 10px',
                                    background: 'var(--bg-secondary)',
                                    border: '1px solid var(--border-color)',
                                    borderRadius: 'var(--radius-sm)',
                                    color: 'var(--text-primary)',
                                    fontSize: '13px',
                                    cursor: 'pointer',
                                    transition: 'var(--transition)',
                                    height: '32px',
                                    maxWidth: isMobile ? '120px' : 'none'
                                }}
                            >
                                <Settings size={13} />
                                <span className="header-user-name" style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                                    {user?.username}
                                </span>
                                <ChevronDown size={13} style={{
                                    transform: showSettings ? 'rotate(180deg)' : 'rotate(0deg)',
                                    transition: 'transform 0.15s',
                                    flexShrink: 0
                                }} />
                            </button>

                            {showSettings && (
                                <div className="settings-dropdown" style={{
                                    position: 'absolute',
                                    top: '100%',
                                    right: 0,
                                    marginTop: '6px',
                                    width: isMobile ? 'min(300px, calc(100vw - 32px))' : '300px',
                                    maxWidth: 'calc(100vw - 32px)',
                                    overflow: 'hidden'
                                }}>
                                    <div style={{
                                        padding: '12px 16px',
                                        borderBottom: '1px solid var(--border-color)',
                                        display: 'flex',
                                        justifyContent: 'space-between',
                                        alignItems: 'center'
                                    }}>
                                        <span style={{ fontWeight: '500', fontSize: '13px' }}>{t.layout.settings}</span>
                                        <button
                                            onClick={resetAndCloseSettings}
                                            style={{
                                                padding: '2px',
                                                borderRadius: '4px',
                                                color: 'var(--text-tertiary)',
                                                cursor: 'pointer'
                                            }}
                                        >
                                            <X size={15} />
                                        </button>
                                    </div>

                                    <div style={{ padding: '16px' }}>
                                        {error && (
                                            <div style={{
                                                padding: '8px 10px',
                                                marginBottom: '12px',
                                                background: 'rgba(255, 0, 0, 0.05)',
                                                border: '1px solid rgba(255, 0, 0, 0.15)',
                                                borderRadius: 'var(--radius-sm)',
                                                color: 'var(--danger)',
                                                fontSize: '13px'
                                            }}>
                                                {error}
                                            </div>
                                        )}

                                        {success && (
                                            <div style={{
                                                padding: '8px 10px',
                                                marginBottom: '12px',
                                                background: 'rgba(0, 224, 84, 0.05)',
                                                border: '1px solid rgba(0, 224, 84, 0.15)',
                                                borderRadius: 'var(--radius-sm)',
                                                color: 'var(--success)',
                                                fontSize: '13px'
                                            }}>
                                                {success}
                                            </div>
                                        )}

                                        <form onSubmit={handlePasswordChange} style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                                            <div>
                                                <label className="form-label" style={{ fontSize: '12px', marginBottom: '4px' }}>{t.common.password.current}</label>
                                                <input
                                                    type="password"
                                                    value={passwordForm.old_password}
                                                    onChange={(e) => setPasswordForm({ ...passwordForm, old_password: e.target.value })}
                                                    className="form-input"
                                                    style={{ height: '34px', fontSize: '13px' }}
                                                    required
                                                />
                                            </div>

                                            <div>
                                                <label className="form-label" style={{ fontSize: '12px', marginBottom: '4px' }}>{t.common.password.new}</label>
                                                <input
                                                    type="password"
                                                    value={passwordForm.new_password}
                                                    onChange={(e) => setPasswordForm({ ...passwordForm, new_password: e.target.value })}
                                                    className="form-input"
                                                    style={{ height: '34px', fontSize: '13px' }}
                                                    required
                                                    minLength={6}
                                                />
                                            </div>

                                            <div>
                                                <label className="form-label" style={{ fontSize: '12px', marginBottom: '4px' }}>{t.common.password.confirm}</label>
                                                <input
                                                    type="password"
                                                    value={passwordForm.confirm_password}
                                                    onChange={(e) => setPasswordForm({ ...passwordForm, confirm_password: e.target.value })}
                                                    className="form-input"
                                                    style={{ height: '34px', fontSize: '13px' }}
                                                    required
                                                    minLength={6}
                                                />
                                            </div>

                                            <button
                                                type="submit"
                                                disabled={updating}
                                                className="btn btn-primary"
                                                style={{ marginTop: '4px', height: '34px', fontSize: '13px' }}
                                            >
                                                {updating ? t.common.updating : t.common.password.change}
                                            </button>
                                        </form>
                                    </div>

                                    <div style={{
                                        padding: '10px 16px',
                                        borderTop: '1px solid var(--border-color)',
                                        background: 'var(--bg-secondary)'
                                    }}>
                                        <button
                                            onClick={handleLogout}
                                            style={{
                                                width: '100%',
                                                padding: '6px 12px',
                                                background: 'transparent',
                                                border: '1px solid var(--border-color)',
                                                borderRadius: 'var(--radius-sm)',
                                                color: 'var(--text-secondary)',
                                                fontSize: '12px',
                                                cursor: 'pointer',
                                                transition: 'var(--transition)'
                                            }}
                                            onMouseEnter={(e) => {
                                                e.target.style.color = 'var(--text-primary)';
                                                e.target.style.borderColor = 'var(--text-tertiary)';
                                            }}
                                            onMouseLeave={(e) => {
                                                e.target.style.color = 'var(--text-secondary)';
                                                e.target.style.borderColor = 'var(--border-color)';
                                            }}
                                        >
                                            {t.layout.logout}
                                        </button>
                                    </div>
                                </div>
                            )}
                        </div>
                    </div>
                </header>

                <div className="app-content">
                    <Outlet />
                </div>
            </main>
            <BackToTop />
        </div>
    );
};

export default Layout;
