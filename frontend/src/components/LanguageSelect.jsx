import { useLanguage } from '../LanguageContext';

const LanguageSelect = ({ className = 'form-input header-language-select', style }) => {
    const { language, changeLanguage, languages } = useLanguage();

    return (
        <select
            value={language}
            onChange={(e) => changeLanguage(e.target.value)}
            className={className}
            style={{
                width: 'auto',
                height: '32px',
                paddingLeft: '8px',
                paddingRight: '28px',
                fontSize: '13px',
                ...style
            }}
        >
            {Object.entries(languages).map(([code, lang]) => (
                <option key={code} value={code}>{lang.name}</option>
            ))}
        </select>
    );
};

export default LanguageSelect;
