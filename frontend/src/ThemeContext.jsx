import { createContext, useContext, useState, useEffect } from 'react';
import { flushSync } from 'react-dom';

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
        let timer;
        
        const applyThemeWithTransition = (newTheme) => {
            const currentTheme = root.getAttribute('data-theme');
            if (currentTheme && currentTheme !== newTheme) {
                root.classList.add('theme-transitioning');
                root.setAttribute('data-theme', newTheme);
                if (timer) clearTimeout(timer);
                timer = setTimeout(() => {
                    root.classList.remove('theme-transitioning');
                }, 300);
            } else {
                root.setAttribute('data-theme', newTheme);
            }
        };

        if (themeMode === 'system') {
            // Listen to system theme changes
            const mediaQuery = window.matchMedia('(prefers-color-scheme: light)');
            const handleChange = (e) => {
                const newTheme = e.matches ? 'light' : 'dark';
                setTheme(newTheme);
                applyThemeWithTransition(newTheme);
            };

            // Set initial theme
            const systemTheme = mediaQuery.matches ? 'light' : 'dark';
            setTheme(systemTheme);
            applyThemeWithTransition(systemTheme);

            // Add listener
            mediaQuery.addEventListener('change', handleChange);
            return () => {
                mediaQuery.removeEventListener('change', handleChange);
                if (timer) clearTimeout(timer);
            };
        } else {
            // Use manual theme
            setTheme(themeMode);
            applyThemeWithTransition(themeMode);
            return () => {
                if (timer) clearTimeout(timer);
            };
        }
    }, [themeMode]);

    const changeTheme = (newMode) => {
        // Fallback for browsers that don't support View Transitions
        if (!document.startViewTransition) {
            setThemeMode(newMode);
            localStorage.setItem('theme', newMode);
            return;
        }

        // Target theme determines coordinates: dark theme starts from bottom-left, light theme from top-right
        const targetTheme = newMode === 'system'
            ? (window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark')
            : newMode;

        const x = targetTheme === 'dark' ? 0 : window.innerWidth;
        const y = targetTheme === 'dark' ? window.innerHeight : 0;
        
        // Calculate radius to the furthest corner (top-right: x = window.innerWidth, y = 0)
        const endRadius = Math.hypot(
            Math.max(x, window.innerWidth - x),
            Math.max(y, window.innerHeight - y)
        );

        const transition = document.startViewTransition(() => {
            flushSync(() => {
                setThemeMode(newMode);
            });
        });

        transition.ready.then(() => {
            const clipPath = [
                `circle(0px at ${x}px ${y}px)`,
                `circle(${endRadius}px at ${x}px ${y}px)`
            ];
            
            document.documentElement.animate(
                {
                    clipPath: clipPath,
                },
                {
                    duration: 700,
                    easing: 'cubic-bezier(0.85, 0, 0.15, 1)',
                    pseudoElement: '::view-transition-new(root)',
                }
            );
        });

        localStorage.setItem('theme', newMode);
    };

    return (
        <ThemeContext.Provider value={{ theme, themeMode, changeTheme }}>
            {children}
        </ThemeContext.Provider>
    );
};
