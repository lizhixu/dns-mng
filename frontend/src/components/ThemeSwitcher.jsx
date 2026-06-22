import { Sun, Moon, Monitor } from 'lucide-react';
import { useTheme } from '../ThemeContext';
import { useLanguage } from '../LanguageContext';

const ThemeSwitcher = ({ style }) => {
    const { themeMode, changeTheme } = useTheme();
    const { t } = useLanguage();

    return (
        <div className="login-theme-switcher" style={style}>
            <button
                onClick={() => changeTheme('light')}
                title={t.layout.themeLight}
                className={`login-theme-btn ${themeMode === 'light' ? 'active' : ''}`}
            >
                <Sun size={13} />
            </button>
            <button
                onClick={() => changeTheme('system')}
                title={t.layout.themeSystem}
                className={`login-theme-btn ${themeMode === 'system' ? 'active' : ''}`}
            >
                <Monitor size={13} />
            </button>
            <button
                onClick={() => changeTheme('dark')}
                title={t.layout.themeDark}
                className={`login-theme-btn ${themeMode === 'dark' ? 'active' : ''}`}
            >
                <Moon size={13} />
            </button>
        </div>
    );
};

export default ThemeSwitcher;
