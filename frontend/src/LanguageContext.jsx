import { createContext, useContext, useState, useEffect } from 'react';
import { languages, defaultLanguage } from './locales';

const LanguageContext = createContext();

export const LanguageProvider = ({ children }) => {
    const [language, setLanguage] = useState(() => {
        // 从 localStorage 读取保存的语言设置
        const saved = localStorage.getItem('language');
        return saved && languages[saved] ? saved : defaultLanguage;
    });

    const t = languages[language].translations;

    const changeLanguage = (lang) => {
        if (languages[lang]) {
            setLanguage(lang);
            localStorage.setItem('language', lang);
        }
    };

    useEffect(() => {
        // 设置 HTML lang 属性
        document.documentElement.lang = language;
    }, [language]);

    return (
        <LanguageContext.Provider value={{ language, t, changeLanguage, languages }}>
            {children}
        </LanguageContext.Provider>
    );
};

export const useLanguage = () => {
    const context = useContext(LanguageContext);
    if (!context) {
        throw new Error('useLanguage must be used within a LanguageProvider');
    }
    return context;
};

export default LanguageContext;
