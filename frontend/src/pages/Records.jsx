import { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate, useLocation } from 'react-router-dom';
import { api } from '../api';
import { ArrowLeft, Plus, Edit2, Trash2, Search, RefreshCw, AlertCircle, Server, CheckCircle } from 'lucide-react';
import Modal from '../components/Modal';
import ConfirmDialog from '../components/ConfirmDialog';
import { useLanguage } from '../LanguageContext';
import useMediaQuery from '../hooks/useMediaQuery';

const RECORD_TYPES = ['A', 'AAAA', 'CNAME', 'MX', 'TXT', 'SPF', 'SRV'];

const Records = () => {
    const { t, language } = useLanguage();
    const isMobile = useMediaQuery('(max-width: 768px)');
    const { accountId, domainId } = useParams();
    const navigate = useNavigate();
    const location = useLocation();
    const [records, setRecords] = useState([]);
    const [filteredRecords, setFilteredRecords] = useState([]);
    const [domain, setDomain] = useState(null);
    const [account, setAccount] = useState(null);
    const [providerDefaultTTL, setProviderDefaultTTL] = useState(300);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [searchTerm, setSearchTerm] = useState('');
    const [typeFilter, setTypeFilter] = useState('ALL');

    // Modal state
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [modalMode, setModalMode] = useState('create');
    const [currentRecord, setCurrentRecord] = useState(null);

    // Delete confirmation dialog state
    const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
    const [deletingRecord, setDeletingRecord] = useState(null);
    const [deleting, setDeleting] = useState(false);

    // DNS Check state
    const [checkingRecord, setCheckingRecord] = useState(null);
    const [checkResult, setCheckResult] = useState(null);
    const [showCheckResult, setShowCheckResult] = useState(false);

    // Form state
    const [formData, setFormData] = useState({
        node_name: '',
        record_type: 'A',
        content: '',
        ttl: 300,
        priority: 10,
        state: true
    });
    const [formError, setFormError] = useState('');
    const [submitting, setSubmitting] = useState(false);

    // Load records and domain data
    useEffect(() => {
        let isMounted = true;

        const loadData = async () => {
            setLoading(true);
            try {
                const [recordsData, domainData, accountsData, providersData] = await Promise.all([
                    api.getRecords(accountId, domainId),
                    api.getDomain(accountId, domainId),
                    api.getAccounts(),
                    api.getProviders()
                ]);

                if (isMounted) {
                    setRecords(recordsData || []);
                    setFilteredRecords(recordsData || []);
                    setDomain(domainData);

                    const acc = accountsData?.find(a => a.id === parseInt(accountId));
                    setAccount(acc);

                    if (acc && providersData) {
                        const prov = providersData.find(p => p.name === acc.provider_type);
                        if (prov && typeof prov.default_ttl === 'number') {
                            setProviderDefaultTTL(prov.default_ttl);
                        }
                    }

                    setError('');
                }
            } catch (err) {
                if (isMounted) {
                    setError(err.message);
                }
            } finally {
                if (isMounted) {
                    setLoading(false);
                }
            }
        };

        loadData();

        return () => {
            isMounted = false;
        };
    }, [accountId, domainId]);

    // Refresh function for manual reload
    const loadRecords = useCallback(async () => {
        setLoading(true);
        try {
            const [recordsData, domainData] = await Promise.all([
                api.getRecords(accountId, domainId),
                api.getDomain(accountId, domainId)
            ]);
            setRecords(recordsData || []);
            setFilteredRecords(recordsData || []);
            setDomain(domainData);
            setError('');
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    }, [accountId, domainId]);

    useEffect(() => {
        let result = records;
        if (typeFilter !== 'ALL') {
            result = result.filter(r => r.record_type === typeFilter);
        }
        if (searchTerm) {
            const term = searchTerm.toLowerCase();
            result = result.filter(r =>
                (r.node_name && r.node_name.toLowerCase().includes(term)) ||
                (r.content && r.content.toLowerCase().includes(term)) ||
                (r.record_type && r.record_type.toLowerCase().includes(term))
            );
        }
        setFilteredRecords(result);
    }, [records, searchTerm, typeFilter]);

    const openCreateModal = () => {
        setModalMode('create');
        setFormData({
            node_name: '',
            record_type: 'A',
            content: '',
            ttl: providerDefaultTTL,
            priority: 10,
            state: true
        });
        setFormError('');
        setIsModalOpen(true);
    };

    const openEditModal = (record) => {
        setModalMode('edit');
        setCurrentRecord(record);
        setFormData({
            node_name: record.node_name,
            record_type: record.record_type,
            content: record.content,
            ttl: record.ttl,
            priority: record.priority || 10,
            state: record.state
        });
        setFormError('');
        setIsModalOpen(true);
    };

    // Validate and sanitize node name
    const sanitizeNodeName = (name) => {
        // Remove leading/trailing whitespace
        let sanitized = name.trim().toLowerCase();

        // Empty or @ means root domain
        if (!sanitized || sanitized === '@') {
            return '';
        }

        // Remove trailing dots
        sanitized = sanitized.replace(/\.$/, '');

        // Replace spaces with hyphens, but keep underscores (needed for _acme-challenge, _dmarc, etc.)
        sanitized = sanitized.replace(/\s+/g, '-');

        // Remove any characters that are not alphanumeric, hyphen, dot, or underscore
        sanitized = sanitized.replace(/[^a-z0-9\-\._]/g, '');

        // Remove consecutive hyphens
        sanitized = sanitized.replace(/-+/g, '-');

        // Remove leading/trailing hyphens (but keep underscores)
        sanitized = sanitized.replace(/^-+|-+$/g, '');

        return sanitized;
    };

    const validateNodeName = (name) => {
        if (!name || name === '@') {
            return true; // Root domain is valid
        }

        // Check length (max 63 characters per label)
        const labels = name.split('.');
        for (const label of labels) {
            if (label.length > 63) {
                return false;
            }
            // Allow underscores (for _acme-challenge, _dmarc, etc.)
            // Must start with alphanumeric or underscore, end with alphanumeric
            if (!/^[a-z0-9_][a-z0-9\-_]*[a-z0-9]$/.test(label) && !/^[a-z0-9_]$/.test(label)) {
                return false;
            }
        }

        return true;
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setSubmitting(true);
        setFormError('');

        try {
            // Sanitize node name
            const sanitizedNodeName = sanitizeNodeName(formData.node_name);

            // Validate node name
            if (!validateNodeName(sanitizedNodeName)) {
                setFormError(t.records.invalidNodeName);
                setSubmitting(false);
                return;
            }

            const payload = {
                ...formData,
                node_name: sanitizedNodeName
            };

            // integer conversion
            payload.ttl = parseInt(payload.ttl);
            if (['MX', 'SRV'].includes(payload.record_type)) {
                payload.priority = parseInt(payload.priority);
            } else {
                delete payload.priority;
            }

            if (modalMode === 'create') {
                const newRecord = await api.createRecord(accountId, domainId, payload);
                setRecords([...records, newRecord]);
            } else {
                const updatedRecord = await api.updateRecord(accountId, domainId, currentRecord.id, payload);
                setRecords(records.map(r => r.id === updatedRecord.id ? updatedRecord : r));
            }
            setIsModalOpen(false);
        } catch (err) {
            setFormError(err.message);
        } finally {
            setSubmitting(false);
        }
    };

    const handleDelete = async (id) => {
        const record = records.find(r => r.id === id);
        setDeletingRecord(record);
        setShowDeleteConfirm(true);
    };

    const confirmDelete = async () => {
        if (!deletingRecord) return;

        setDeleting(true);
        try {
            await api.deleteRecord(accountId, domainId, deletingRecord.id);
            setRecords(records.filter(r => r.id !== deletingRecord.id));
            setShowDeleteConfirm(false);
        } catch (err) {
            alert(err.message);
        } finally {
            setDeleting(false);
            setDeletingRecord(null);
        }
    };

    const cancelDelete = () => {
        setShowDeleteConfirm(false);
        setDeletingRecord(null);
    };

    // DNS Check functions
    const handleCheckDNS = async (record) => {
        setCheckingRecord(record.id);
        setCheckResult(null);

        try {
            // Build full domain name
            // Root record (@) should use the domain name directly
            const fullDomain = record.node_name && record.node_name !== '@'
                ? `${record.node_name}.${domain.name}`
                : domain.name;

            const result = await api.checkDNS({
                domain: fullDomain,
                record_type: record.record_type,
                expected: record.content
            });

            setCheckResult(result);
            setShowCheckResult(true);
        } catch (err) {
            setCheckResult({
                domain: record.node_name ? `${record.node_name}.${domain.name}` : domain.name,
                record_type: record.record_type,
                values: [],
                matched: false,
                message: err.message
            });
            setShowCheckResult(true);
        } finally {
            setCheckingRecord(null);
        }
    };

    const getContentLabel = (type) => {
        return t.records.contentLabels[type] || t.records.contentLabels.default;
    };

    // Smart back navigation - go back to previous page if it's domains related, otherwise go to account domains
    const handleBack = () => {
        const from = location.state?.from;
        if (from === '/domains' || from?.startsWith('/accounts/')) {
            navigate(-1);
        } else {
            navigate(`/accounts/${accountId}/domains`);
        }
    };

    // Determine back button text based on navigation source
    const getBackText = () => {
        const from = location.state?.from;
        if (from === '/domains') {
            return t.records.backToAllDomains;
        }
        return t.records.backToDomains;
    };

    return (
        <div>
            <div style={{ marginBottom: '1.5rem' }}>
                <button
                    onClick={handleBack}
                    style={{
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: '0.5rem',
                        color: 'var(--text-secondary)',
                        fontSize: '0.875rem',
                        marginBottom: '0.75rem',
                        background: 'none',
                        border: 'none',
                        cursor: 'pointer',
                        padding: 0,
                        fontFamily: 'inherit',
                        transition: 'color 0.15s'
                    }}
                    className="hover-text-primary"
                >
                    <ArrowLeft size={14} />
                    {getBackText()}
                </button>
                <div className="page-title-row" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
                    <h2 style={{ fontSize: '1.5rem', fontWeight: 'bold', letterSpacing: '-0.02em', margin: 0 }}>{t.records.title}</h2>
                    <button onClick={openCreateModal} className="btn btn-primary" style={{ height: '34px', fontSize: '13px' }}>
                        <Plus size={14} />
                        {t.records.addRecord}
                    </button>
                </div>
            </div>

            {/* Domain Info Card */}
            {domain && (
                <div className="domain-list-card" style={{ padding: '0.75rem 1rem', marginBottom: '0.75rem', cursor: 'default' }}>
                    <div className="record-domain-info-layout" style={{ display: 'flex', alignItems: 'center', gap: '2rem', flexWrap: 'wrap' }}>
                        <div className="record-domain-main" style={{ minWidth: '200px' }}>
                            <div style={{ fontSize: '11px', color: 'var(--text-secondary)', textTransform: 'uppercase', tracking: '0.05em', marginBottom: '2px' }}>
                                {t.records.domainName}
                            </div>
                            <div className="font-mono" style={{ fontSize: '15px', fontWeight: '600', color: 'var(--text-primary)' }}>
                                {domain.unicode_name || domain.name}
                            </div>
                        </div>
                        {account && (
                            <div className="record-domain-account" style={{ minWidth: '150px' }}>
                                <div style={{ fontSize: '11px', color: 'var(--text-secondary)', textTransform: 'uppercase', tracking: '0.05em', marginBottom: '2px' }}>
                                    {t.records.account}
                                </div>
                                <div style={{ display: 'flex', alignItems: 'center', gap: '0.375rem', fontSize: '13px', color: 'var(--text-primary)' }}>
                                    <Server size={13} style={{ color: 'var(--text-secondary)' }} />
                                    <span>{account.name}</span>
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            )}

            <div className="records-filter-bar" style={{ marginBottom: '0.75rem', display: 'flex', gap: '0.5rem', flexWrap: 'wrap', alignItems: 'center' }}>
                <div className="records-search-box" style={{ position: 'relative', flex: '1 1 160px', minWidth: 0 }}>
                    <Search size={14} style={{ position: 'absolute', left: '0.75rem', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-tertiary)' }} />
                    <input
                        type="text"
                        className="form-input"
                        placeholder={t.records.searchPlaceholder}
                        style={{ paddingLeft: '2.25rem', width: '100%', height: '34px' }}
                        value={searchTerm}
                        onChange={e => setSearchTerm(e.target.value)}
                    />
                </div>
                <div className="records-filter-controls" style={{ display: 'flex', gap: '0.5rem', flexShrink: 0 }}>
                    <select
                        className="form-input records-filter-select"
                        style={{ width: 'auto', height: '34px', paddingRight: '1.5rem', fontSize: '13px' }}
                        value={typeFilter}
                        onChange={e => setTypeFilter(e.target.value)}
                    >
                        <option value="ALL">{t.records.allTypes}</option>
                        {RECORD_TYPES.map(type => <option key={type} value={type}>{type}</option>)}
                    </select>
                    <button onClick={loadRecords} className="btn btn-secondary records-refresh-btn" style={{ height: '34px', padding: '0 10px' }} title={t.common.refresh}>
                        <RefreshCw size={15} className={loading ? "spin" : ""} style={{ animation: loading ? 'spin 1s linear infinite' : 'none' }} />
                    </button>
                </div>
            </div>

            {error && <div style={{ color: 'var(--danger)', marginBottom: '1rem', padding: '0.75rem 1rem', backgroundColor: 'rgba(255, 0, 0, 0.05)', border: '1px solid rgba(255, 0, 0, 0.15)', borderRadius: 'var(--radius-sm)', fontSize: '14px' }}>{t.common.error}: {error}</div>}

            {loading && !records.length ? (
                <div style={{ display: 'flex', justifyContent: 'center', padding: '4rem 0' }}>
                    <div className="spinner"></div>
                </div>
            ) : isMobile ? (
                /* 移动端卡片视图 */
                <div className="mobile-only-cards fade-in" style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                        {filteredRecords.length === 0 ? (
                            <div className="domain-list-card" style={{ padding: '2rem', textAlign: 'center', color: 'var(--text-tertiary)', cursor: 'default' }}>
                                {t.records.noRecords}
                            </div>
                        ) : (
                            filteredRecords.map(record => (
                                <div key={record.id} className="domain-list-card" style={{ padding: '1rem', cursor: 'default' }}>
                                    <div className="record-card-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem', gap: '0.5rem' }}>
                                        <span className="badge badge-neutral" style={{ fontWeight: 'bold', fontSize: '11px', height: '20px' }}>
                                            {record.record_type}
                                        </span>
                                        <span className="record-card-node font-mono" style={{ fontWeight: 600, fontSize: '14px', flex: 1, color: 'var(--text-primary)' }}>
                                            {record.node_name || '@'}
                                        </span>
                                        <span className={`badge ${record.state ? 'badge-success' : 'badge-neutral'}`} style={{ fontSize: '11px', height: '20px' }}>
                                            {record.state ? t.common.active : t.common.inactive}
                                        </span>
                                    </div>
                                    <div className="record-card-content" style={{ marginBottom: '0.5rem' }}>
                                        <div className="record-card-content-row" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', flexWrap: 'wrap' }}>
                                            <span style={{
                                                fontSize: '10px',
                                                fontWeight: 600,
                                                padding: '1px 4px',
                                                borderRadius: '3px',
                                                backgroundColor: 'var(--bg-secondary)',
                                                border: '1px solid var(--border-color)',
                                                color: 'var(--text-secondary)',
                                                textTransform: 'uppercase'
                                            }}>
                                                {getContentLabel(record.record_type)}
                                            </span>
                                            <span className="record-card-value font-mono" style={{ fontSize: '13px', wordBreak: 'break-all', color: 'var(--text-primary)' }}>
                                                {record.content}
                                            </span>
                                        </div>
                                        {record.record_type === 'MX' && record.priority != null && (
                                            <div style={{ fontSize: '11px', color: 'var(--text-tertiary)', marginTop: '4px' }}>
                                                {t.records.priority}: {record.priority}
                                            </div>
                                        )}
                                    </div>
                                    <div className="record-card-footer" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', paddingTop: '0.5rem', marginTop: '0.5rem' }}>
                                        <span style={{ color: 'var(--text-tertiary)', fontSize: '11px' }}>
                                            {record.updated_on ? new Date(record.updated_on).toLocaleString(language === 'en' ? 'en-US' : 'zh-CN', {
                                                month: '2-digit',
                                                day: '2-digit',
                                                hour: '2-digit',
                                                minute: '2-digit'
                                            }) : '-'}
                                        </span>
                                        <div className="record-card-actions" style={{ display: 'flex', gap: '2px' }}>
                                            <button
                                                onClick={() => handleCheckDNS(record)}
                                                className="btn btn-ghost"
                                                title={t.records.checkDns}
                                                disabled={checkingRecord === record.id}
                                                style={{ padding: '4px', minWidth: 'auto', height: 'auto', color: 'var(--text-secondary)' }}
                                            >
                                                {checkingRecord === record.id ? (
                                                    <div className="spinner" style={{ width: '13px', height: '13px', borderWidth: '2px' }}></div>
                                                ) : (
                                                    <CheckCircle size={13} />
                                                )}
                                            </button>
                                            <button onClick={() => openEditModal(record)} className="btn btn-ghost" title={t.common.edit} style={{ padding: '4px', minWidth: 'auto', height: 'auto', color: 'var(--text-secondary)' }}>
                                                <Edit2 size={13} />
                                            </button>
                                            <button onClick={() => handleDelete(record.id)} className="btn btn-ghost" style={{ color: 'var(--danger)', padding: '4px', minWidth: 'auto', height: 'auto' }} title={t.common.delete}>
                                                <Trash2 size={13} />
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            ))
                        )}
                    </div>
            ) : (
                    /* 桌面端表格视图 */
                    <div className="desktop-only-table domain-list-card table-scroll-container fade-in" style={{ overflowX: 'auto', padding: 0, cursor: 'default' }}>
                        <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', marginBottom: 0 }}>
                            <thead>
                                <tr style={{ borderBottom: '1px solid var(--border-color)', height: '40px' }}>
                                    <th style={{ padding: '10px 16px', color: 'var(--text-secondary)', fontWeight: 500, fontSize: '13px' }}>{t.records.type}</th>
                                    <th style={{ padding: '10px 16px', color: 'var(--text-secondary)', fontWeight: 500, fontSize: '13px' }}>{t.records.nodeName}</th>
                                    <th style={{ padding: '10px 16px', color: 'var(--text-secondary)', fontWeight: 500, fontSize: '13px' }}>{t.records.contentLabels.default}</th>
                                    <th style={{ padding: '10px 16px', color: 'var(--text-secondary)', fontWeight: 500, fontSize: '13px' }}>{t.common.status}</th>
                                    <th style={{ padding: '10px 16px', color: 'var(--text-secondary)', fontWeight: 500, fontSize: '13px' }}>{t.records.updatedAt}</th>
                                    <th style={{ padding: '10px 16px', color: 'var(--text-secondary)', fontWeight: 500, fontSize: '13px', textAlign: 'right' }}>{t.common.actions}</th>
                                </tr>
                            </thead>
                            <tbody>
                                {filteredRecords.map((record, index) => (
                                    <tr key={record.id} style={{ borderBottom: index === filteredRecords.length - 1 ? 'none' : '1px solid var(--border-color)', transition: 'background 0.2s' }}>
                                        <td style={{ padding: '10px 16px' }}>
                                            <span className="badge badge-neutral" style={{ fontWeight: 'bold', width: '54px', justifyContent: 'center', height: '20px', fontSize: '11px' }}>
                                                {record.record_type}
                                            </span>
                                        </td>
                                        <td style={{ padding: '10px 16px', fontWeight: 600, fontSize: '14px' }} className="font-mono">{record.node_name || '@'}</td>
                                        <td style={{ padding: '10px 16px', maxWidth: '400px' }}>
                                            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', flexWrap: 'wrap' }}>
                                                <span style={{
                                                    fontSize: '10px',
                                                    fontWeight: 600,
                                                    padding: '1px 4px',
                                                    borderRadius: '3px',
                                                    backgroundColor: 'var(--bg-secondary)',
                                                    border: '1px solid var(--border-color)',
                                                    color: 'var(--text-secondary)',
                                                    textTransform: 'uppercase'
                                                }}>
                                                    {getContentLabel(record.record_type)}
                                                </span>
                                                <span className="font-mono" style={{ fontSize: '13px', wordBreak: 'break-all', color: 'var(--text-primary)' }} title={record.content}>
                                                    {record.content}
                                                </span>
                                            </div>
                                            {record.record_type === 'MX' && record.priority != null && (
                                                <span style={{ fontSize: '11px', color: 'var(--text-tertiary)', marginLeft: '0.25rem' }}>
                                                    ({t.records.priority}: {record.priority})
                                                </span>
                                            )}
                                        </td>
                                        <td style={{ padding: '10px 16px' }}>
                                            <span className={`badge ${record.state ? 'badge-success' : 'badge-neutral'}`} style={{ fontSize: '11px', height: '20px' }}>
                                                {record.state ? t.common.active : t.common.inactive}
                                            </span>
                                        </td>
                                        <td style={{ padding: '10px 16px', color: 'var(--text-secondary)', fontSize: '12px' }}>
                                            {record.updated_on ? new Date(record.updated_on).toLocaleString(language === 'en' ? 'en-US' : 'zh-CN', {
                                                year: 'numeric',
                                                month: '2-digit',
                                                day: '2-digit',
                                                hour: '2-digit',
                                                minute: '2-digit'
                                            }) : '-'}
                                        </td>
                                        <td style={{ padding: '10px 16px', textAlign: 'right' }}>
                                            <div style={{ display: 'flex', gap: '4px', justifyContent: 'flex-end' }}>
                                                <button
                                                    onClick={() => handleCheckDNS(record)}
                                                    className="btn btn-ghost"
                                                    title={t.records.checkDns}
                                                    disabled={checkingRecord === record.id}
                                                    style={{ padding: '4px', minWidth: 'auto', height: 'auto', color: 'var(--text-secondary)' }}
                                                >
                                                    {checkingRecord === record.id ? (
                                                        <div className="spinner" style={{ width: '14px', height: '14px', borderWidth: '2px' }}></div>
                                                    ) : (
                                                        <CheckCircle size={14} />
                                                    )}
                                                </button>
                                                <button onClick={() => openEditModal(record)} className="btn btn-ghost" title={t.common.edit} style={{ padding: '4px', minWidth: 'auto', height: 'auto', color: 'var(--text-secondary)' }}>
                                                    <Edit2 size={14} />
                                                </button>
                                                <button onClick={() => handleDelete(record.id)} className="btn btn-ghost" style={{ color: 'var(--danger)', padding: '4px', minWidth: 'auto', height: 'auto' }} title={t.common.delete}>
                                                    <Trash2 size={14} />
                                                </button>
                                            </div>
                                        </td>
                                    </tr>
                                ))}
                                {filteredRecords.length === 0 && (
                                    <tr>
                                        <td colSpan="6" style={{ padding: '3rem', textAlign: 'center', color: 'var(--text-tertiary)' }}>
                                            {t.records.noRecords}
                                        </td>
                                    </tr>
                                )}
                            </tbody>
                        </table>
                    </div>
            )}

            <Modal
                isOpen={isModalOpen}
                onClose={() => setIsModalOpen(false)}
                title={modalMode === 'create' ? t.records.addDnsRecord : t.records.editDnsRecord}
            >
                <form onSubmit={handleSubmit}>
                    {formError && (
                        <div style={{ backgroundColor: 'rgba(239, 68, 68, 0.1)', color: 'var(--danger)', padding: '0.75rem', borderRadius: 'var(--radius-md)', marginBottom: '1rem', fontSize: '0.875rem', display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                            <AlertCircle size={16} />
                            {formError}
                        </div>
                    )}

                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                        <div className="form-group">
                            <label className="form-label">{t.records.type}</label>
                            <select
                                className="form-input"
                                value={formData.record_type}
                                onChange={e => setFormData({ ...formData, record_type: e.target.value })}
                                disabled={modalMode === 'edit'}
                            >
                                {RECORD_TYPES.map(type => <option key={type} value={type}>{type}</option>)}
                            </select>
                        </div>

                        <div className="form-group">
                            <label className="form-label">{t.records.nodeName}</label>
                            <input
                                type="text"
                                className="form-input"
                                placeholder={t.records.nodeNamePlaceholder}
                                value={formData.node_name}
                                onChange={e => setFormData({ ...formData, node_name: e.target.value })}
                            />
                            <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)', marginTop: '0.25rem' }}>
                                {t.records.rootRecordHint}
                            </div>
                        </div>
                    </div>

                    <div className="form-group">
                        <label className="form-label">{getContentLabel(formData.record_type)}</label>
                        <input
                            type="text"
                            className="form-input"
                            placeholder={formData.record_type === 'A' ? '1.2.3.4' : ''}
                            value={formData.content}
                            onChange={e => setFormData({ ...formData, content: e.target.value })}
                            required
                        />
                    </div>

                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                        <div className="form-group">
                            <label className="form-label">{t.records.ttlSeconds}</label>
                            <input
                                type="number"
                                className="form-input"
                                value={formData.ttl}
                                onChange={e => setFormData({ ...formData, ttl: e.target.value })}
                                required
                            />
                        </div>

                        {['MX', 'SRV'].includes(formData.record_type) && (
                            <div className="form-group">
                                <label className="form-label">{t.records.priority}</label>
                                <input
                                    type="number"
                                    className="form-input"
                                    value={formData.priority}
                                    onChange={e => setFormData({ ...formData, priority: e.target.value })}
                                    required
                                />
                            </div>
                        )}
                    </div>

                    <div className="form-group">
                        <label className="form-label" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', cursor: 'pointer' }}>
                            <input
                                type="checkbox"
                                checked={formData.state}
                                onChange={e => setFormData({ ...formData, state: e.target.checked })}
                                style={{ width: '1rem', height: '1rem', accentColor: 'var(--accent-primary)' }}
                            />
                            {t.common.active}
                        </label>
                    </div>

                    <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '2rem' }}>
                        <button type="button" onClick={() => setIsModalOpen(false)} className="btn btn-ghost">{t.common.cancel}</button>
                        <button type="submit" className="btn btn-primary" disabled={submitting}>
                            {submitting ? <div className="spinner" style={{ width: '1rem', height: '1rem', borderWidth: '2px' }}></div> : (modalMode === 'create' ? t.records.addRecord : t.common.save)}
                        </button>
                    </div>
                </form>
            </Modal>

            {/* Delete Confirmation Dialog */}
            <ConfirmDialog
                isOpen={showDeleteConfirm}
                onClose={cancelDelete}
                onConfirm={confirmDelete}
                title={t.common.confirmDelete}
                message={
                    deletingRecord
                        ? t.records.deleteRecordMessage
                            .replace('{name}', deletingRecord.node_name || '@')
                            .replace('{type}', deletingRecord.record_type)
                        : t.common.confirmDeleteMessage
                }
                confirmText={t.common.delete}
                cancelText={t.common.cancel}
                loading={deleting}
                danger={true}
            />

            {/* DNS Check Result Modal */}
            <Modal
                isOpen={showCheckResult}
                onClose={() => setShowCheckResult(false)}
                title={t.records.checkDnsTitle}
                size="large"
                footer={
                    <button
                        onClick={() => setShowCheckResult(false)}
                        className="btn btn-primary"
                    >
                        {t.common.close}
                    </button>
                }
            >
                {checkResult && (
                    <>
                        <div style={{ marginBottom: '1.5rem' }}>
                            <div style={{
                                padding: '1rem',
                                borderRadius: 'var(--radius-md)',
                                background: checkResult.matched ? 'rgba(16, 185, 129, 0.1)' : 'rgba(239, 68, 68, 0.1)',
                                border: `1px solid ${checkResult.matched ? 'rgba(16, 185, 129, 0.3)' : 'rgba(239, 68, 68, 0.3)'}`,
                                display: 'flex',
                                alignItems: 'center',
                                gap: '0.75rem'
                            }}>
                                <CheckCircle
                                    size={24}
                                    style={{
                                        color: checkResult.matched ? '#10b981' : '#ef4444',
                                        flexShrink: 0
                                    }}
                                />
                                <div style={{ flex: 1 }}>
                                    <div style={{
                                        fontWeight: '600',
                                        marginBottom: '0.25rem',
                                        color: checkResult.matched ? '#10b981' : '#ef4444'
                                    }}>
                                        {checkResult.matched ? `✓ ${t.records.checkDnsNormal}` : `✗ ${t.records.checkDnsAbnormal}`}
                                    </div>
                                    <div style={{ fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
                                        {checkResult.message === 'DNS record matches expected value' ? t.records.checkDnsMatched :
                                            checkResult.message === 'DNS record does not match expected value' ? t.records.checkDnsNotMatched :
                                                checkResult.message === 'DNS query failed' ? t.records.checkDnsFailed :
                                                    checkResult.message === 'No DNS records found' ? t.records.checkDnsNoRecords :
                                                        checkResult.message === 'DNS record found' ? t.records.checkDnsFound :
                                                            checkResult.message}
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div style={{
                            display: 'grid',
                            gap: '1rem',
                            gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))'
                        }}>
                            <div>
                                <label style={{ fontSize: '0.875rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '0.5rem' }}>
                                    {t.records.domainName}
                                </label>
                                <div style={{
                                    padding: '0.75rem',
                                    background: 'var(--bg-secondary)',
                                    borderRadius: 'var(--radius-sm)',
                                    fontFamily: 'monospace',
                                    fontSize: '0.875rem',
                                    wordBreak: 'break-all'
                                }}>
                                    {checkResult.domain}
                                </div>
                            </div>

                            <div>
                                <label style={{ fontSize: '0.875rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '0.5rem' }}>
                                    {t.records.recordType}
                                </label>
                                <div style={{
                                    padding: '0.75rem',
                                    background: 'var(--bg-secondary)',
                                    borderRadius: 'var(--radius-sm)',
                                    fontFamily: 'monospace',
                                    fontSize: '0.875rem'
                                }}>
                                    {checkResult.record_type}
                                </div>
                            </div>

                            {checkResult.expected && (
                                <div style={{ gridColumn: '1 / -1' }}>
                                    <label style={{ fontSize: '0.875rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '0.5rem' }}>
                                        {t.records.expectedValue}
                                    </label>
                                    <div style={{
                                        padding: '0.75rem',
                                        background: 'var(--bg-secondary)',
                                        borderRadius: 'var(--radius-sm)',
                                        fontFamily: 'monospace',
                                        fontSize: '0.875rem',
                                        wordBreak: 'break-all',
                                        whiteSpace: 'pre-wrap'
                                    }}>
                                        {checkResult.expected}
                                    </div>
                                </div>
                            )}

                            <div style={{ gridColumn: '1 / -1' }}>
                                <label style={{ fontSize: '0.875rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '0.5rem' }}>
                                    {t.records.actualValue}
                                </label>
                                <div style={{
                                    padding: '0.75rem',
                                    background: 'var(--bg-secondary)',
                                    borderRadius: 'var(--radius-sm)',
                                    fontFamily: 'monospace',
                                    fontSize: '0.875rem',
                                    minHeight: '2.5rem'
                                }}>
                                    {checkResult.values && checkResult.values.length > 0 ? (
                                        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                                            {checkResult.values.map((value, index) => (
                                                <div key={index} style={{
                                                    padding: '0.5rem',
                                                    background: 'var(--bg-tertiary)',
                                                    borderRadius: 'var(--radius-sm)',
                                                    wordBreak: 'break-all',
                                                    whiteSpace: 'pre-wrap'
                                                }}>
                                                    {value}
                                                </div>
                                            ))}
                                        </div>
                                    ) : (
                                        <span style={{ color: 'var(--text-tertiary)' }}>{t.records.noRecordsFound}</span>
                                    )}
                                </div>
                            </div>

                            <div>
                                <label style={{ fontSize: '0.875rem', color: 'var(--text-tertiary)', display: 'block', marginBottom: '0.5rem' }}>
                                    {t.records.checkTime}
                                </label>
                                <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)' }}>
                                    {checkResult.timestamp ? new Date(checkResult.timestamp).toLocaleString() : '-'}
                                </div>
                            </div>

                            {checkResult.dns_server && (
                                <div>
                                    <label style={{ fontSize: '0.875rem', color: 'var(--text-tertiary)', display: 'block', marginBottom: '0.5rem' }}>
                                        {t.records.dnsServer}
                                    </label>
                                    <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)', fontFamily: 'monospace' }}>
                                        {checkResult.dns_server}
                                    </div>
                                </div>
                            )}
                        </div>
                    </>
                )}
            </Modal>
        </div>
    );
};

export default Records;
