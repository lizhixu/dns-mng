import { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate, useLocation } from 'react-router-dom';
import { api } from '../api';
import { ArrowLeft, Plus, Edit2, Trash2, Search, RefreshCw, AlertCircle, Server } from 'lucide-react';
import Modal from '../components/Modal';
import ConfirmDialog from '../components/ConfirmDialog';
import { useLanguage } from '../LanguageContext';

const RECORD_TYPES = ['A', 'AAAA', 'CNAME', 'MX', 'TXT', 'SPF', 'SRV'];

const Records = () => {
    const { t } = useLanguage();
    const { accountId, domainId } = useParams();
    const navigate = useNavigate();
    const location = useLocation();
    const [records, setRecords] = useState([]);
    const [filteredRecords, setFilteredRecords] = useState([]);
    const [domain, setDomain] = useState(null);
    const [account, setAccount] = useState(null);
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
                const [recordsData, domainData, accountsData] = await Promise.all([
                    api.getRecords(accountId, domainId),
                    api.getDomain(accountId, domainId),
                    api.getAccounts()
                ]);
                
                if (isMounted) {
                    setRecords(recordsData || []);
                    setFilteredRecords(recordsData || []);
                    setDomain(domainData);
                    setAccount(accountsData?.find(a => a.id === parseInt(accountId)));
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
            ttl: 300,
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

        // Replace underscores and spaces with hyphens (DNS doesn't allow these in hostnames)
        sanitized = sanitized.replace(/[_\s]+/g, '-');

        // Remove any characters that are not alphanumeric, hyphen, or dot
        sanitized = sanitized.replace(/[^a-z0-9\-\.]/g, '');

        // Remove consecutive hyphens
        sanitized = sanitized.replace(/-+/g, '-');

        // Remove leading/trailing hyphens
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
            // Must start and end with alphanumeric
            if (!/^[a-z0-9][a-z0-9\-]*[a-z0-9]$/.test(label) && !/^[a-z0-9]$/.test(label)) {
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
                setFormError(t.records.invalidNodeName || 'Invalid node name. Node name must contain only letters, numbers, and hyphens, and cannot start or end with a hyphen.');
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
            return t.records.backToAllDomains || t.records.backToDomains;
        }
        return t.records.backToDomains;
    };

    return (
        <div>
            <div style={{ marginBottom: '2rem' }}>
                <button
                    onClick={handleBack}
                    style={{
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: '0.5rem',
                        color: 'var(--text-secondary)',
                        fontSize: '0.875rem',
                        marginBottom: '1rem',
                        background: 'none',
                        border: 'none',
                        cursor: 'pointer',
                        padding: 0,
                        fontFamily: 'inherit'
                    }}
                    onMouseEnter={(e) => e.target.style.color = 'var(--text-primary)'}
                    onMouseLeave={(e) => e.target.style.color = 'var(--text-secondary)'}
                >
                    <ArrowLeft size={16} />
                    {getBackText()}
                </button>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
                    <h2 style={{ fontSize: '1.5rem', fontWeight: 'bold' }}>{t.records.title}</h2>
                    <button onClick={openCreateModal} className="btn btn-primary">
                        <Plus size={18} />
                        {t.records.addRecord}
                    </button>
                </div>
            </div>

            {/* Domain Info Card */}
            {domain && (
                <div className="glass-panel" style={{ padding: '1rem 1.25rem', marginBottom: '1.5rem' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '1.5rem', flexWrap: 'wrap' }}>
                        <div style={{ flex: 1, minWidth: '200px' }}>
                            <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)', marginBottom: '0.25rem' }}>
                                {t.records.domainName || 'Domain'}
                            </div>
                            <div style={{ fontSize: '1rem', fontWeight: '600', color: 'var(--text-primary)' }}>
                                {domain.unicode_name || domain.name}
                            </div>
                        </div>
                        {account && (
                            <div style={{ minWidth: '150px' }}>
                                <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)', marginBottom: '0.25rem' }}>
                                    {t.records.account}
                                </div>
                                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.875rem', color: 'var(--text-primary)' }}>
                                    <Server size={14} style={{ color: 'var(--text-secondary)' }} />
                                    {account.name}
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            )}

            <div className="glass-panel" style={{ padding: '1rem', marginBottom: '1.5rem', display: 'flex', gap: '1rem', flexWrap: 'wrap', alignItems: 'center' }}>
                <div style={{ position: 'relative', flex: 1 }}>
                    <Search size={16} style={{ position: 'absolute', left: '0.75rem', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-tertiary)' }} />
                    <input
                        type="text"
                        className="form-input"
                        placeholder={t.records.searchPlaceholder}
                        style={{ paddingLeft: '2.5rem' }}
                        value={searchTerm}
                        onChange={e => setSearchTerm(e.target.value)}
                    />
                </div>
                <select
                    className="form-input"
                    style={{ width: 'auto' }}
                    value={typeFilter}
                    onChange={e => setTypeFilter(e.target.value)}
                >
                    <option value="ALL">{t.records.allTypes}</option>
                    {RECORD_TYPES.map(type => <option key={type} value={type}>{type}</option>)}
                </select>
                <button onClick={loadRecords} className="btn btn-secondary" title={t.common.refresh}>
                    <RefreshCw size={18} className={loading ? "spin" : ""} style={{ animation: loading ? 'spin 1s linear infinite' : 'none' }} />
                </button>
            </div>

            {error && <div style={{ color: 'var(--danger)', marginBottom: '1rem', padding: '1rem', backgroundColor: 'rgba(239, 68, 68, 0.1)', borderRadius: 'var(--radius-md)' }}>{t.common.error}: {error}</div>}

            {loading && !records.length ? (
                <div className="spinner" style={{ margin: '4rem auto' }}></div>
            ) : (
                <div className="glass-panel" style={{ overflowX: 'auto' }}>
                    <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left' }}>
                        <thead>
                            <tr style={{ borderBottom: '1px solid var(--border-color)' }}>
                                <th style={{ padding: '1rem', color: 'var(--text-secondary)', fontWeight: 500 }}>{t.records.type}</th>
                                <th style={{ padding: '1rem', color: 'var(--text-secondary)', fontWeight: 500 }}>{t.records.nodeName}</th>
                                <th style={{ padding: '1rem', color: 'var(--text-secondary)', fontWeight: 500 }}>{t.records.content}</th>
                                <th style={{ padding: '1rem', color: 'var(--text-secondary)', fontWeight: 500 }}>{t.records.updatedAt}</th>
                                <th style={{ padding: '1rem', color: 'var(--text-secondary)', fontWeight: 500, textAlign: 'right' }}>{t.common.actions}</th>
                            </tr>
                        </thead>
                        <tbody>
                            {filteredRecords.map(record => (
                                <tr key={record.id} style={{ borderBottom: '1px solid var(--border-color)', transition: 'background 0.2s', ':hover': { backgroundColor: 'rgba(255,255,255,0.02)' } }}>
                                    <td style={{ padding: '1rem' }}>
                                        <span className="badge badge-neutral" style={{ fontWeight: 'bold', width: '60px', justifyContent: 'center' }}>
                                            {record.record_type}
                                        </span>
                                    </td>
                                    <td style={{ padding: '1rem', fontWeight: 500 }}>{record.node_name || '@'}</td>
                                    <td style={{ padding: '1rem', maxWidth: '400px' }}>
                                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', flexWrap: 'wrap' }}>
                                            <span style={{
                                                display: 'inline-block',
                                                fontSize: '0.7rem',
                                                fontWeight: 600,
                                                padding: '0.15rem 0.4rem',
                                                borderRadius: '4px',
                                                backgroundColor: 'rgba(99, 102, 241, 0.15)',
                                                color: 'var(--accent-primary)',
                                                whiteSpace: 'nowrap',
                                                textTransform: 'uppercase',
                                                letterSpacing: '0.03em'
                                            }}>
                                                {getContentLabel(record.record_type)}
                                            </span>
                                            <span style={{ fontFamily: 'monospace', fontSize: '0.875rem', wordBreak: 'break-all' }} title={record.content}>
                                                {record.content}
                                            </span>
                                        </div>
                                        {record.record_type === 'MX' && record.priority != null && (
                                            <span style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)', marginLeft: '0.25rem' }}>
                                                ({t.records.priority}: {record.priority})
                                            </span>
                                        )}
                                    </td>
                                    <td style={{ padding: '1rem', color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
                                        {record.updated_on ? new Date(record.updated_on).toLocaleString('zh-CN', { 
                                            year: 'numeric', 
                                            month: '2-digit', 
                                            day: '2-digit',
                                            hour: '2-digit',
                                            minute: '2-digit'
                                        }) : '-'}
                                    </td>
                                    <td style={{ padding: '1rem', textAlign: 'right' }}>
                                        <div style={{ display: 'flex', gap: '0.5rem', justifyContent: 'flex-end' }}>
                                            <button onClick={() => openEditModal(record)} className="btn btn-ghost" title={t.common.edit}>
                                                <Edit2 size={16} />
                                            </button>
                                            <button onClick={() => handleDelete(record.id)} className="btn btn-ghost" style={{ color: 'var(--danger)' }} title={t.common.delete}>
                                                <Trash2 size={16} />
                                            </button>
                                        </div>
                                    </td>
                                </tr>
                            ))}
                            {filteredRecords.length === 0 && (
                                <tr>
                                    <td colSpan="5" style={{ padding: '3rem', textAlign: 'center', color: 'var(--text-tertiary)' }}>
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
                                min="60"
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
                            {submitting ? <div className="spinner" style={{ width: '1rem', height: '1rem', borderWidth: '2px' }}></div> : (modalMode === 'create' ? t.records.addRecord : t.accounts.saveChanges)}
                        </button>
                    </div>
                </form>
            </Modal>

            {/* Delete Confirmation Dialog */}
            <ConfirmDialog
                isOpen={showDeleteConfirm}
                onClose={cancelDelete}
                onConfirm={confirmDelete}
                title={t.records.deleteRecordTitle}
                message={
                    deletingRecord 
                        ? t.records.deleteRecordMessage
                            .replace('{name}', deletingRecord.node_name || '@')
                            .replace('{type}', deletingRecord.record_type)
                        : t.records.deleteRecordMessageDefault
                }
                confirmText={t.common.delete}
                cancelText={t.common.cancel}
                loading={deleting}
                danger={true}
            />
        </div>
    );
};

export default Records;
