import { useState, useEffect, useCallback, useRef } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api';
import { Search, RefreshCw, Globe, ExternalLink, Server } from 'lucide-react';
import { useLanguage } from '../LanguageContext';

const AllDomains = () => {
    const { t } = useLanguage();
    const [domains, setDomains] = useState([]);
    const [filteredDomains, setFilteredDomains] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [searchTerm, setSearchTerm] = useState('');
    const fetchedRef = useRef(false);

    const loadDomains = useCallback(async () => {
        setLoading(true);
        try {
            const data = await api.getAllDomains();
            setDomains(data || []);
            setFilteredDomains(data || []);
            setError('');
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        // Prevent duplicate requests in React StrictMode
        if (fetchedRef.current) return;
        fetchedRef.current = true;
        loadDomains();
    }, [loadDomains]);

    useEffect(() => {
        if (!searchTerm) {
            setFilteredDomains(domains);
        } else {
            const term = searchTerm.toLowerCase();
            setFilteredDomains(
                domains.filter(d =>
                    d.name.toLowerCase().includes(term) ||
                    (d.account_name && d.account_name.toLowerCase().includes(term)) ||
                    (d.ipv4_address && d.ipv4_address.toLowerCase().includes(term))
                )
            );
        }
    }, [searchTerm, domains]);

    return (
        <div>
            <div style={{ marginBottom: '2rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
                    <div>
                        <h2 style={{ fontSize: '1.5rem', fontWeight: 'bold' }}>{t.allDomains.title}</h2>
                        <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', marginTop: '0.25rem' }}>{t.allDomains.subtitle}</p>
                    </div>
                    <div style={{ display: 'flex', gap: '0.75rem' }}>
                        <div style={{ position: 'relative' }}>
                            <Search size={16} style={{ position: 'absolute', left: '0.75rem', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-tertiary)' }} />
                            <input
                                type="text"
                                className="form-input"
                                placeholder={t.allDomains.searchPlaceholder}
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
                        <div key={`${domain.account_id}-${domain.id}`} className="glass-panel" style={{ padding: '1.25rem', transition: 'transform 0.2s' }}>
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                <div style={{ display: 'flex', gap: '1rem', alignItems: 'flex-start' }}>
                                    <div style={{ marginTop: '0.25rem', color: 'var(--accent-primary)' }}>
                                        <Globe size={24} />
                                    </div>
                                    <div>
                                        <h3 style={{ fontSize: '1.1rem', fontWeight: '600', marginBottom: '0.25rem' }}>{domain.name}</h3>
                                        <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center', fontSize: '0.875rem', flexWrap: 'wrap' }}>
                                            <span className={`badge ${domain.state === 'Active' ? 'badge-success' : 'badge-warning'}`}>
                                                {domain.state}
                                            </span>
                                            {domain.ipv4_address && (
                                                <span style={{ color: 'var(--text-secondary)' }}>{domain.ipv4_address}</span>
                                            )}
                                            <span style={{ display: 'inline-flex', alignItems: 'center', gap: '0.25rem', color: 'var(--text-tertiary)', fontSize: '0.8rem' }}>
                                                <Server size={12} />
                                                {domain.account_name}
                                            </span>
                                        </div>
                                    </div>
                                </div>
                                <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: '0.5rem' }}>
                                    <span style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)' }}>{t.domains.ttl}: {domain.ttl}s</span>
                                    <Link
                                        to={`/accounts/${domain.account_id}/domains/${domain.id}/records`}
                                        state={{ from: '/domains' }}
                                        className="btn btn-secondary"
                                        style={{ fontSize: '0.875rem', padding: '0.4rem 0.8rem' }}
                                    >
                                        {t.domains.manageRecords}
                                        <ExternalLink size={14} />
                                    </Link>
                                </div>
                            </div>
                        </div>
                    ))}

                    {filteredDomains.length === 0 && (
                        <div style={{ textAlign: 'center', padding: '4rem', color: 'var(--text-secondary)' }}>
                            {searchTerm ? t.allDomains.noSearchResults : t.allDomains.noDomains}
                        </div>
                    )}
                </div>
            )}
        </div>
    );
};

export default AllDomains;
