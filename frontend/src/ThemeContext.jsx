import { createContext, useContext, useState, useEffect } from 'react';

const ThemeContext = createContext();

export const useTheme = () => {
    const context = useContext(ThemeContext);
    if (!context) {
        throw new Error('useTheme must be used within ThemeProvider');
    }
    return context;
};

export const ThemeProvider = ({ children }) => {
    const [theme, setTheme] = useState(() => {
        // Check localStorage first
        const savedTheme = localStorage.getItem('theme');
        if (savedTheme && savedTheme !== 'system') {
            return savedTheme;
        }
        // Check system preference
        if (window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches) {
            return 'light';
        }
        return 'dark';
    });

    const [themeMode, setThemeMode] = useState(() => {
        return localStorage.getItem('theme') || 'system';
    });

    useEffect(() => {
        const root = document.documentElement;
        
        if (themeMode === 'system') {
            // Listen to system theme changes
            const mediaQuery = window.matchMedia('(prefers-color-scheme: light)');
            const handleChange = (e) => {
                const newTheme = e.matches ? 'light' : 'dark';
                setTheme(newTheme);
                root.setAttribute('data-theme', newTheme);
            };

            // Set initial theme
            const systemTheme = mediaQuery.matches ? 'light' : 'dark';
            setTheme(systemTheme);
            root.setAttribute('data-theme', systemTheme);

            // Add listener
            mediaQuery.addEventListener('change', handleChange);
            return () => mediaQuery.removeEventListener('change', handleChange);
        } else {
            // Use manual theme
            root.setAttribute('data-theme', themeMode);
            setTheme(themeMode);
        }
    }, [themeMode]);

    const changeTheme = (newMode) => {
        setThemeMode(newMode);
        localStorage.setItem('theme', newMode);
    };

    return (
        <ThemeContext.Provider value={{ theme, themeMode, changeTheme }}>
            {children}
        </ThemeContext.Provider>
    );
};
