import { Outlet, Link, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../AuthContext';
import { useLanguage } from '../LanguageContext';
import { useTheme } from '../ThemeContext';
import { Sun, Moon, Monitor } from 'lucide-react';

const Layout = () => {
    const { user, logout } = useAuth();
    const navigate = useNavigate();
    const location = useLocation();
    const { themeMode, changeTheme } = useTheme();

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
                flexDirection: 'column'
            }}>
                <h1 style={{ 
                    fontSize: '16px', 
                    fontWeight: '600', 
                    marginBottom: '24px',
                    padding: '0 8px'
                }}>
                    {t.layout.title}
                </h1>
                <nav style={{ flex: 1 }}>
                    <Link 
                        to="/domains" 
                        style={{ 
                            display: 'block', 
                            padding: '8px', 
                            marginBottom: '4px', 
                            borderRadius: 'var(--radius-sm)',
                            background: isActive('/domains') ? 'var(--bg-tertiary)' : 'transparent',
                            color: isActive('/domains') ? 'var(--text-primary)' : 'var(--text-secondary)',
                            fontSize: '14px',
                            fontWeight: isActive('/domains') ? '500' : '400',
                            transition: 'var(--transition)',
                            textDecoration: 'none'
                        }}
                        onMouseEnter={(e) => {
                            if (!isActive('/domains')) {
                                e.target.style.background = 'var(--bg-hover)';
                                e.target.style.color = 'var(--text-primary)';
                            }
                        }}
                        onMouseLeave={(e) => {
                            if (!isActive('/domains')) {
                                e.target.style.background = 'transparent';
                                e.target.style.color = 'var(--text-secondary)';
                            }
                        }}
                    >
                        {t.layout.domains}
                    </Link>
                    <Link 
                        to="/accounts" 
                        style={{ 
                            display: 'block', 
                            padding: '8px', 
                            marginBottom: '4px', 
                            borderRadius: 'var(--radius-sm)',
                            background: isActive('/accounts') ? 'var(--bg-tertiary)' : 'transparent',
                            color: isActive('/accounts') ? 'var(--text-primary)' : 'var(--text-secondary)',
                            fontSize: '14px',
                            fontWeight: isActive('/accounts') ? '500' : '400',
                            transition: 'var(--transition)',
                            textDecoration: 'none'
                        }}
                        onMouseEnter={(e) => {
                            if (!isActive('/accounts')) {
                                e.target.style.background = 'var(--bg-hover)';
                                e.target.style.color = 'var(--text-primary)';
                            }
                        }}
                        onMouseLeave={(e) => {
                            if (!isActive('/accounts')) {
                                e.target.style.background = 'transparent';
                                e.target.style.color = 'var(--text-secondary)';
                            }
                        }}
                    >
                        {t.layout.accounts}
                    </Link>
                </nav>
            </aside>

            {/* 主内容区 */}
            <main style={{
                flex: 1,
                display: 'flex',
                flexDirection: 'column',
                overflow: 'hidden'
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

                        <span style={{ fontSize: '14px', color: 'var(--text-secondary)' }}>{user?.username}</span>
                        <button onClick={handleLogout} className="btn btn-secondary" style={{ height: '32px', fontSize: '13px' }}>
                            {t.layout.logout}
                        </button>
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
        </div>
    );
};

export default Layout;
