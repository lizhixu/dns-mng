import { useState, useEffect, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api } from '../api';
import { ArrowLeft, Search, RefreshCw, Globe, ExternalLink } from 'lucide-react';
import { useLanguage } from '../LanguageContext';

const Domains = () => {
    const { t } = useLanguage();
    const { accountId } = useParams();
    const [domains, setDomains] = useState([]);
    const [filteredDomains, setFilteredDomains] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [searchTerm, setSearchTerm] = useState('');

    const loadDomains = useCallback(async () => {
        setLoading(true);
        try {
            const data = await api.getDomains(accountId);
            setDomains(data || []);
            setFilteredDomains(data || []);
            setError('');
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    }, [accountId]);

    useEffect(() => {
        loadDomains();
    }, [loadDomains]);

    return (
        <div>
            <div style={{ marginBottom: '2rem' }}>
                <Link to="/accounts" style={{ display: 'inline-flex', alignItems: 'center', gap: '0.5rem', color: 'var(--text-secondary)', fontSize: '0.875rem', marginBottom: '1rem', transition: 'color 0.2s', ':hover': { color: 'var(--text-primary)' } }}>
                    <ArrowLeft size={16} />
                    {t.domains.backToAccounts}
                </Link>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
                    <h2 style={{ fontSize: '1.5rem', fontWeight: 'bold' }}>{t.domains.title}</h2>
                    <div style={{ display: 'flex', gap: '0.75rem' }}>
                        <div style={{ position: 'relative' }}>
                            <Search size={16} style={{ position: 'absolute', left: '0.75rem', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-tertiary)' }} />
                            <input
                                type="text"
                                className="form-input"
                                placeholder={t.domains.searchPlaceholder}
                                style={{ paddingLeft: '2.5rem', width: '250px' }}
                                value={searchTerm}
                                onChange={e => setSearchTerm(e.target.value)}
                            />
                        </div>
                        <button onClick={loadDomains} className="btn btn-secondary" title={t.common.refresh}>
                            <RefreshCw size={18} className={loading ? "spin" : ""} style={{ animation: loading ? 'spin 1s linear infinite' : 'none' }} />
                        </button>
                    </div>
                </div>
            </div>

            {error && <div style={{ color: 'var(--danger)', marginBottom: '1rem', padding: '1rem', backgroundColor: 'rgba(239, 68, 68, 0.1)', borderRadius: 'var(--radius-md)' }}>{t.common.error}: {error}</div>}

            {loading && !domains.length ? (
                <div className="spinner" style={{ margin: '4rem auto' }}></div>
            ) : (
                <div style={{ display: 'grid', gap: '1rem' }}>
                    {filteredDomains.map(domain => (
                        <div key={domain.id} className="glass-panel" style={{ 
                            padding: '1rem 1.25rem',
                            transition: 'all 0.2s',
                            cursor: 'pointer'
                        }}
                        onMouseEnter={(e) => {
                            e.currentTarget.style.transform = 'translateY(-2px)';
                            e.currentTarget.style.boxShadow = '0 4px 12px rgba(0, 0, 0, 0.1)';
                        }}
                        onMouseLeave={(e) => {
                            e.currentTarget.style.transform = 'translateY(0)';
                            e.currentTarget.style.boxShadow = 'none';
                        }}>
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center', flex: 1 }}>
                                    <Globe size={20} style={{ color: 'var(--accent-primary)', flexShrink: 0 }} />
                                    <div>
                                        <h3 style={{ fontSize: '1rem', fontWeight: '500', margin: 0, marginBottom: '0.25rem' }}>{domain.name}</h3>
                                        {domain.updated_on && (
                                            <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)' }}>
                                                更新于 {new Date(domain.updated_on).toLocaleString('zh-CN', { 
                                                    year: 'numeric', 
                                                    month: '2-digit', 
                                                    day: '2-digit',
                                                    hour: '2-digit',
                                                    minute: '2-digit'
                                                })}
                                            </div>
                                        )}
                                    </div>
                                </div>
                                <Link to={`/accounts/${accountId}/domains/${domain.id}/records`} className="btn btn-secondary" style={{ fontSize: '0.875rem', padding: '0.5rem 1rem' }}>
                                    {t.domains.manageRecords}
                                    <ExternalLink size={14} />
                                </Link>
                            </div>
                        </div>
                    ))}

                    {filteredDomains.length === 0 && (
                        <div style={{ textAlign: 'center', padding: '4rem', color: 'var(--text-secondary)' }}>
                            {searchTerm ? t.domains.noSearchResults : t.domains.noDomains}
                        </div>
                    )}
                </div>
            )}
        </div>
    );
};

export default Domains;
