import { useState } from 'react';
import { Outlet, Link, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../AuthContext';
import { useLanguage } from '../LanguageContext';
import { useTheme } from '../ThemeContext';
import { api } from '../api';
import { Sun, Moon, Monitor, FileText, Globe, Server, User, Settings, ChevronDown, X } from 'lucide-react';
import BackToTop from './BackToTop';


const Layout = () => {
    const { user, logout } = useAuth();
    const navigate = useNavigate();
    const location = useLocation();
    const { themeMode, changeTheme } = useTheme();
    const [showSettings, setShowSettings] = useState(false);
    const [updating, setUpdating] = useState(false);
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');
    const [passwordForm, setPasswordForm] = useState({
        old_password: '',
        new_password: '',
        confirm_password: ''
    });

    const handleLogout = () => {
        logout();
        navigate('/login');
    };
    const { language, changeLanguage, languages, t } = useLanguage();

    // Check if a path is active - improved logic
    const isActive = (path) => {
        // Exact match
        if (location.pathname === path) return true;
        // Check if current path starts with the menu path
        // Add trailing slash to avoid false matches (e.g., /accounts matching /account)
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
            setError('新密码与确认密码不匹配');
            return;
        }

        if (passwordForm.new_password.length < 6) {
            setError('新密码至少需要6个字符');
            return;
        }

        setUpdating(true);
        try {
            await api.updatePassword({
                old_password: passwordForm.old_password,
                new_password: passwordForm.new_password
            });
            setSuccess('密码已更新');
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

    return (
        <div style={{ display: 'flex', height: '100vh', overflow: 'hidden' }}>
            {/* 侧边栏 */}
            <aside style={{
                width: '200px',
                background: 'var(--bg-primary)',
                borderRight: '1px solid var(--border-color)',
                padding: '24px 16px',
                flexShrink: 0,
                display: 'flex',
                flexDirection: 'column',
                overflow: 'auto'
            }}>
                <h1 style={{ 
                    fontSize: '16px', 
                    fontWeight: '600', 
                    marginBottom: '24px',
                    padding: '0 8px',
                    flexShrink: 0
                }}>
                    {t.layout.title}
                </h1>
                <nav style={{ flex: 1, minHeight: 0 }}>
                    <Link to="/domains" style={menuItemStyle('/domains')}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                            <Globe size={14} />
                            {t.layout.domains}
                        </div>
                    </Link>
                    <Link to="/accounts" style={menuItemStyle('/accounts')}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                            <Server size={14} />
                            {t.layout.accounts}
                        </div>
                    </Link>
                    <Link to="/logs" style={menuItemStyle('/logs')}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                            <FileText size={14} />
                            操作日志
                        </div>
                    </Link>
                </nav>
            </aside>

            {/* 主内容区 */}
            <main style={{
                flex: 1,
                display: 'flex',
                flexDirection: 'column',
                overflow: 'hidden',
                minWidth: 0
            }}>
                {/* 顶部 Header */}
                <header style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    padding: '0 24px',
                    height: '64px',
                    background: 'var(--bg-primary)',
                    borderBottom: '1px solid var(--border-color)',
                    flexShrink: 0
                }}>
                    <div></div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                        {/* Theme Switcher */}
                        <div style={{ display: 'flex', alignItems: 'center', gap: '2px', background: 'var(--bg-secondary)', border: '1px solid var(--border-color)', borderRadius: 'var(--radius-sm)', padding: '2px' }}>
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

                        <div style={{ 
                            height: '20px', 
                            width: '1px', 
                            background: 'var(--border-color)' 
                        }}></div>

                        {/* User Menu */}
                        <div style={{ position: 'relative' }}>
                            <button
                                onClick={() => setShowSettings(!showSettings)}
                                style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: '0.5rem',
                                    padding: '6px 10px',
                                    background: 'var(--bg-secondary)',
                                    border: '1px solid var(--border-color)',
                                    borderRadius: 'var(--radius-sm)',
                                    color: 'var(--text-primary)',
                                    fontSize: '14px',
                                    cursor: 'pointer',
                                    transition: 'var(--transition)'
                                }}
                            >
                                <Settings size={14} />
                                {user?.username}
                                <ChevronDown size={14} style={{ 
                                    transform: showSettings ? 'rotate(180deg)' : 'rotate(0deg)',
                                    transition: 'transform 0.2s'
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
                                        width: '320px',
                                        background: 'var(--bg-primary)',
                                        border: '1px solid var(--border-color)',
                                        borderRadius: 'var(--radius-md)',
                                        boxShadow: 'var(--shadow-lg)',
                                        zIndex: 50,
                                        overflow: 'hidden'
                                    }}>
                                        <div style={{
                                            padding: '16px',
                                            borderBottom: '1px solid var(--border-color)',
                                            display: 'flex',
                                            justifyContent: 'space-between',
                                            alignItems: 'center'
                                        }}>
                                            <span style={{ fontWeight: '500', fontSize: '14px' }}>个人设置</span>
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
                                                    <label className="form-label" style={{ fontSize: '13px' }}>当前密码</label>
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
                                                    <label className="form-label" style={{ fontSize: '13px' }}>新密码</label>
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
                                                    <label className="form-label" style={{ fontSize: '13px' }}>确认密码</label>
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
                                                    {updating ? '更新中...' : '修改密码'}
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

                {/* 内容区 */}
                <div style={{
                    flex: 1,
                    overflow: 'auto',
                    padding: '24px',
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
