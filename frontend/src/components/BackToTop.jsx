import { useState, useEffect } from 'react';
import { ArrowUp } from 'lucide-react';
import { useLanguage } from '../LanguageContext';

const BackToTop = () => {
    const [isVisible, setIsVisible] = useState(false);
    const { t } = useLanguage();

    useEffect(() => {
        // Wait for DOM to be ready
        const timer = setTimeout(() => {
            // Find the main content scrollable div
            const mainContent = document.querySelector('main > div[style*="overflow: auto"]');
            
            if (!mainContent) {
                console.log('Scroll container not found');
                return;
            }

            const toggleVisibility = () => {
                if (mainContent.scrollTop > 300) {
                    setIsVisible(true);
                } else {
                    setIsVisible(false);
                }
            };

            // Initial check
            toggleVisibility();

            mainContent.addEventListener('scroll', toggleVisibility);

            return () => {
                mainContent.removeEventListener('scroll', toggleVisibility);
            };
        }, 100);

        return () => clearTimeout(timer);
    }, []);

    const scrollToTop = () => {
        const mainContent = document.querySelector('main > div[style*="overflow: auto"]');
        if (mainContent) {
            mainContent.scrollTo({
                top: 0,
                behavior: 'smooth'
            });
        }
    };

    if (!isVisible) return null;

    return (
        <button
            onClick={scrollToTop}
            style={{
                position: 'fixed',
                bottom: '0',
                right: '20px',
                padding: '8px 14px',
                borderRadius: '6px 6px 0 0',
                background: 'var(--surface-secondary)',
                color: 'var(--text-primary)',
                border: '1px solid var(--border-primary)',
                borderBottom: 'none',
                boxShadow: '0 -2px 6px rgba(0, 0, 0, 0.08)',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                gap: '6px',
                fontSize: '13px',
                fontWeight: '500',
                transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                zIndex: 40,
                opacity: 0.9,
                backdropFilter: 'blur(12px)',
                WebkitBackdropFilter: 'blur(12px)'
            }}
            onMouseEnter={(e) => {
                e.currentTarget.style.transform = 'translateY(-3px)';
                e.currentTarget.style.boxShadow = '0 -3px 10px rgba(0, 0, 0, 0.12)';
                e.currentTarget.style.opacity = '1';
                e.currentTarget.style.background = 'var(--accent-primary)';
                e.currentTarget.style.color = '#fff';
            }}
            onMouseLeave={(e) => {
                e.currentTarget.style.transform = 'translateY(0)';
                e.currentTarget.style.boxShadow = '0 -2px 6px rgba(0, 0, 0, 0.08)';
                e.currentTarget.style.opacity = '0.9';
                e.currentTarget.style.background = 'var(--surface-secondary)';
                e.currentTarget.style.color = 'var(--text-primary)';
            }}
            aria-label="Back to top"
            title={t.common.backToTop}
        >
            <ArrowUp size={16} strokeWidth={2.5} />
            <span>{t.common.top}</span>
        </button>
    );
};

export default BackToTop;