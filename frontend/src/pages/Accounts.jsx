import { useState, useEffect, useRef } from 'react';
import { api } from '../api';
import { Link } from 'react-router-dom';
import { Plus, Settings, Trash2, ExternalLink, Eye, EyeOff, Copy, Key, RefreshCw } from 'lucide-react';
import Modal from '../components/Modal';
import ConfirmDialog from '../components/ConfirmDialog';
import { useLanguage } from '../LanguageContext';

const Accounts = () => {
    const { t } = useLanguage();
    const [accounts, setAccounts] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    
    // Prevent duplicate requests in React StrictMode
    const fetchedRef = useRef(false);

    // Modal state
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [modalMode, setModalMode] = useState('create'); // 'create' | 'edit'
    const [currentAccount, setCurrentAccount] = useState(null);
    const [providers, setProviders] = useState([]);

    // Form state
    const [formData, setFormData] = useState({
        name: '',
        provider_type: '',
        api_key: ''
    });
    const [formError, setFormError] = useState('');
    const [submitting, setSubmitting] = useState(false);

    // Visibility state
    const [visibleKeys, setVisibleKeys] = useState({});
    const [showModalApiKey, setShowModalApiKey] = useState(false);

    // Delete confirmation dialog state
    const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
    const [deletingAccount, setDeletingAccount] = useState(null);
    const [deleting, setDeleting] = useState(false);

    // DDNS Token modal state
    const [ddnsTokenModalOpen, setDdnsTokenModalOpen] = useState(false);
    const [ddnsTokensLoading, setDdnsTokensLoading] = useState(false);
    const [ddnsTokenEditModal, setDdnsTokenEditModal] = useState(false);
    const [ddnsTokenForm, setDdnsTokenForm] = useState({ token: '', enabled: true });
    const [ddnsTokenFormError, setDdnsTokenFormError] = useState('');
    const [ddnsTokenSubmitting, setDdnsTokenSubmitting] = useState(false);
    const [ddnsTokenDeleteConfirm, setDdnsTokenDeleteConfirm] = useState(false);
    const [ddnsTokenDeleteLoading, setDdnsTokenDeleteLoading] = useState(false);
    const [showDdnsTokenValue, setShowDdnsTokenValue] = useState(false);

    const toggleKeyVisibility = (id) => {
        setVisibleKeys(prev => ({
            ...prev,
            [id]: !prev[id]
        }));
    };

    useEffect(() => {
        // Prevent duplicate requests in React StrictMode
        if (fetchedRef.current) return;
        fetchedRef.current = true;
        
        loadData();
    }, []);

    const loadData = async () => {
        try {
            const [accountsData, providersData] = await Promise.all([
                api.getAccounts(),
                api.getProviders()
            ]);
            setAccounts(accountsData || []);
            setProviders(providersData || []);
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    const openCreateModal = () => {
        setModalMode('create');
        setFormData({ name: '', provider_type: providers[0]?.name || '', api_key: '' });
        setFormError('');
        setShowModalApiKey(false);
        setIsModalOpen(true);
    };

    const openEditModal = (account) => {
        setModalMode('edit');
        setCurrentAccount(account);
        setFormData({ name: account.name, provider_type: account.provider_type, api_key: account.api_key });
        setFormError('');
        setShowModalApiKey(false);
        setIsModalOpen(true);
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setSubmitting(true);
        setFormError('');

        try {
            if (modalMode === 'create') {
                const newAccount = await api.createAccount(formData);
                setAccounts([newAccount, ...accounts]);
            } else {
                const updatedAccount = await api.updateAccount(currentAccount.id, {
                    name: formData.name,
                    api_key: formData.api_key
                });
                setAccounts(accounts.map(acc => acc.id === updatedAccount.id ? updatedAccount : acc));
            }
            setIsModalOpen(false);
        } catch (err) {
            setFormError(err.message);
        } finally {
            setSubmitting(false);
        }
    };

    const handleDelete = async (id) => {
        const account = accounts.find(acc => acc.id === id);
        setDeletingAccount(account);
        setShowDeleteConfirm(true);
    };

    const confirmDelete = async () => {
        if (!deletingAccount) return;
        
        setDeleting(true);
        try {
            await api.deleteAccount(deletingAccount.id);
            setAccounts(accounts.filter(acc => acc.id !== deletingAccount.id));
            setShowDeleteConfirm(false);
        } catch (err) {
            alert(err.message);
        } finally {
            setDeleting(false);
            setDeletingAccount(null);
        }
    };

    const cancelDelete = () => {
        setShowDeleteConfirm(false);
        setDeletingAccount(null);
    };

    // DDNS Token functions (user-level, single token)
    const [ddnsToken, setDdnsToken] = useState(null);
    const ddnsTokenLoadingRef = useRef(false);

    const openDdnsTokenModal = async () => {
        setDdnsTokenModalOpen(true);
        if (ddnsTokenLoadingRef.current) return;
        ddnsTokenLoadingRef.current = true;
        setDdnsTokensLoading(true);
        try {
            const result = await api.getDDNSToken();
            setDdnsToken(result);
        } catch (err) {
            console.error('Failed to load DDNS token:', err);
        } finally {
            setDdnsTokensLoading(false);
            ddnsTokenLoadingRef.current = false;
        }
    };

    const refreshDdnsToken = async () => {
        if (ddnsTokenLoadingRef.current) return;
        ddnsTokenLoadingRef.current = true;
        setDdnsTokensLoading(true);
        try {
            const result = await api.getDDNSToken();
            setDdnsToken(result);
        } catch (err) {
            console.error('Failed to load DDNS token:', err);
        } finally {
            setDdnsTokensLoading(false);
            ddnsTokenLoadingRef.current = false;
        }
    };

    const openDdnsTokenEdit = () => {
        setDdnsTokenForm({
            token: ddnsToken?.token?.token || '',
            enabled: ddnsToken?.token?.enabled ?? true
        });
        setDdnsTokenFormError('');
        setShowDdnsTokenValue(false);
        setDdnsTokenEditModal(true);
    };

    const openDdnsTokenCreate = () => {
        // Auto-generate a random token
        const randomToken = Array.from(crypto.getRandomValues(new Uint8Array(32)))
            .map(b => b.toString(16).padStart(2, '0'))
            .join('');
        setDdnsTokenForm({ token: randomToken, enabled: true });
        setDdnsTokenFormError('');
        setShowDdnsTokenValue(true);
        setDdnsTokenEditModal(true);
    };

    const handleDdnsTokenSubmit = async (e) => {
        e.preventDefault();
        setDdnsTokenSubmitting(true);
        setDdnsTokenFormError('');
        try {
            await api.updateDDNSToken({
                token: ddnsTokenForm.token,
                enabled: ddnsTokenForm.enabled,
            });
            setDdnsTokenEditModal(false);
            await refreshDdnsToken();
        } catch (err) {
            setDdnsTokenFormError(err.message);
        } finally {
            setDdnsTokenSubmitting(false);
        }
    };

    const confirmDdnsTokenDelete = () => {
        setDdnsTokenDeleteConfirm(true);
    };

    const handleDdnsTokenDelete = async () => {
        setDdnsTokenDeleteLoading(true);
        try {
            await api.deleteDDNSToken();
            setDdnsTokenDeleteConfirm(false);
            setDdnsToken(null);
            await refreshDdnsToken();
        } catch (err) {
            alert(err.message);
        } finally {
            setDdnsTokenDeleteLoading(false);
        }
    };

    const copyDdnsToken = (token) => {
        navigator.clipboard.writeText(token);
    };

    if (loading) return <div className="spinner" style={{ margin: '2rem auto' }}></div>;
    if (error) return <div style={{ color: 'var(--danger)', padding: '1rem' }}>{t.common.error}: {error}</div>;

    return (
        <div>
            <div style={{ marginBottom: '32px' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '8px' }}>
                    <h1 style={{ fontSize: '24px', fontWeight: '600', margin: 0 }}>{t.accounts.title}</h1>
                    <div style={{ display: 'flex', gap: '8px' }}>
                        <button onClick={() => openDdnsTokenModal()} className="btn btn-secondary">
                            <Key size={16} />
                            {t.accounts.manageDDNSToken}
                        </button>
                        <button onClick={openCreateModal} className="btn btn-primary">
                            <Plus size={16} />
                            {t.accounts.addAccount}
                        </button>
                    </div>
                </div>
                <p style={{ color: 'var(--text-secondary)', fontSize: '14px', margin: 0 }}>{t.accounts.subtitle}</p>
            </div>

            <div style={{ 
                display: 'grid', 
                gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))',
                gap: '16px'
            }}>
                {accounts.map(account => (
                    <div key={account.id} className="glass-panel" style={{ 
                        padding: '20px',
                        display: 'flex',
                        flexDirection: 'column',
                        height: '100%'
                    }}>
                        <div style={{ flex: 1 }}>
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '12px' }}>
                                <div style={{ flex: 1, minWidth: 0 }}>
                                    <h3 style={{ 
                                        fontSize: '16px', 
                                        fontWeight: '500', 
                                        margin: 0,
                                        marginBottom: '6px',
                                        overflow: 'hidden',
                                        textOverflow: 'ellipsis',
                                        whiteSpace: 'nowrap'
                                    }}>
                                        {account.name}
                                    </h3>
                                    <span className="badge badge-neutral" style={{ fontSize: '11px', height: '20px' }}>
                                        {providers.find(p => p.name === account.provider_type)?.display_name || account.provider_type}
                                    </span>
                                </div>
                                <div style={{ display: 'flex', gap: '2px', marginLeft: '8px' }}>
                                    <button 
                                        onClick={() => openEditModal(account)} 
                                        className="btn btn-ghost" 
                                        title={t.accounts.editAccount} 
                                        style={{ padding: '4px', minWidth: 'auto', height: 'auto' }}
                                    >
                                        <Settings size={14} />
                                    </button>
                                    <button 
                                        onClick={() => handleDelete(account.id)} 
                                        className="btn btn-ghost" 
                                        style={{ color: 'var(--danger)', padding: '4px', minWidth: 'auto', height: 'auto' }} 
                                        title={t.accounts.deleteAccount}
                                    >
                                        <Trash2 size={14} />
                                    </button>
                                </div>
                            </div>

                            <div style={{ fontSize: '12px', color: 'var(--text-tertiary)', marginBottom: '12px' }}>
                                {t.accounts.addedOn} {new Date(account.created_at).toLocaleDateString()}
                            </div>

                            <div style={{ 
                                display: 'flex', 
                                alignItems: 'center', 
                                gap: '6px',
                                padding: '8px 10px',
                                background: 'var(--bg-tertiary)',
                                border: '1px solid var(--border-color)',
                                borderRadius: 'var(--radius-sm)',
                                marginBottom: '16px'
                            }}>
                                <div style={{
                                    flex: 1,
                                    fontSize: '12px',
                                    color: 'var(--text-secondary)',
                                    fontFamily: 'monospace',
                                    overflow: 'hidden',
                                    textOverflow: 'ellipsis',
                                    whiteSpace: 'nowrap'
                                }}>
                                    {visibleKeys[account.id] ? account.api_key : '••••••••••••••••'}
                                </div>
                                <button
                                    onClick={() => toggleKeyVisibility(account.id)}
                                    className="btn btn-ghost"
                                    style={{ padding: '3px', minWidth: 'auto', height: 'auto' }}
                                    title={visibleKeys[account.id] ? t.common.hide : t.common.show}
                                >
                                    {visibleKeys[account.id] ? <EyeOff size={13} /> : <Eye size={13} />}
                                </button>
                                <button
                                    onClick={() => navigator.clipboard.writeText(account.api_key)}
                                    className="btn btn-ghost"
                                    style={{ padding: '3px', minWidth: 'auto', height: 'auto' }}
                                    title={t.common.copy}
                                >
                                    <Copy size={13} />
                                </button>
                            </div>
                        </div>

                        <Link 
                            to={`/accounts/${account.id}/domains`} 
                            className="btn btn-secondary" 
                            style={{ 
                                fontSize: '13px',
                                width: '100%',
                                justifyContent: 'center',
                                marginBottom: providers.find(p => p.name === account.provider_type)?.website_url ? '8px' : '0'
                            }}
                        >
                            {t.accounts.viewDomains}
                            <ExternalLink size={13} />
                        </Link>

                        {providers.find(p => p.name === account.provider_type)?.website_url && (
                            <a
                                href={providers.find(p => p.name === account.provider_type).website_url}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="btn btn-ghost"
                                style={{
                                    fontSize: '13px',
                                    width: '100%',
                                    justifyContent: 'center'
                                }}
                            >
                                {t.accounts.goToProvider}
                                <ExternalLink size={13} />
                            </a>
                        )}
                    </div>
                ))}

                {accounts.length === 0 && (
                    <div className="glass-panel" style={{ 
                        textAlign: 'center', 
                        padding: '64px 32px',
                        borderStyle: 'dashed',
                        gridColumn: '1 / -1'
                    }}>
                        <div style={{ color: 'var(--text-tertiary)', marginBottom: '16px', fontSize: '14px' }}>
                            {t.accounts.noAccounts}
                        </div>
                        <button onClick={openCreateModal} className="btn btn-primary">
                            <Plus size={16} />
                            {t.accounts.linkFirst}
                        </button>
                    </div>
                )}
            </div>

            <Modal
                isOpen={isModalOpen}
                onClose={() => setIsModalOpen(false)}
                title={modalMode === 'create' ? t.accounts.linkNewAccount : t.accounts.editAccount}
            >
                <form onSubmit={handleSubmit}>
                    {formError && (
                        <div style={{ 
                            backgroundColor: 'rgba(238, 0, 0, 0.1)', 
                            color: 'var(--danger)', 
                            padding: '12px', 
                            borderRadius: 'var(--radius-sm)', 
                            marginBottom: '20px', 
                            fontSize: '13px',
                            border: '1px solid rgba(238, 0, 0, 0.2)'
                        }}>
                            {formError}
                        </div>
                    )}

                    <div className="form-group">
                        <label className="form-label">{t.accounts.accountName}</label>
                        <input
                            type="text"
                            className="form-input"
                            placeholder={t.accounts.accountNamePlaceholder}
                            value={formData.name}
                            onChange={e => setFormData({ ...formData, name: e.target.value })}
                            required
                        />
                    </div>

                    <div className="form-group">
                        <label className="form-label">{t.accounts.provider}</label>
                        <select
                            className="form-input"
                            value={formData.provider_type}
                            onChange={e => setFormData({ ...formData, provider_type: e.target.value })}
                            disabled={modalMode === 'edit'}
                            required
                        >
                            <option value="" disabled>{t.accounts.selectProvider}</option>
                            {providers.map(p => (
                                <option key={p.name} value={p.name}>{p.display_name}</option>
                            ))}
                        </select>
                        {modalMode === 'edit' && (
                            <p style={{ fontSize: '12px', color: 'var(--text-tertiary)', marginTop: '6px', margin: 0 }}>
                                {t.accounts.providerCannotChange}
                            </p>
                        )}
                    </div>

                    <div className="form-group">
                        <label className="form-label">{t.accounts.apiKey}</label>
                        <div style={{ position: 'relative' }}>
                            <input
                                type={showModalApiKey ? "text" : "password"}
                                className="form-input"
                                placeholder={
                                    formData.provider_type === 'tencentcloud' 
                                        ? 'SecretId,SecretKey' 
                                        : formData.provider_type === 'dnshe'
                                            ? 'API Key,API Secret'
                                            : formData.provider_type === 'cloudflare' || formData.provider_type === 'ndjp' || formData.provider_type === 'desec'
                                                ? 'API Token'
                                                : t.accounts.apiKeyPlaceholder
                                }
                                value={formData.api_key}
                                onChange={e => setFormData({ ...formData, api_key: e.target.value })}
                                required={modalMode === 'create'}
                                style={{ paddingRight: '40px' }}
                            />
                            <button
                                type="button"
                                onClick={() => setShowModalApiKey(!showModalApiKey)}
                                style={{
                                    position: 'absolute',
                                    right: '8px',
                                    top: '50%',
                                    transform: 'translateY(-50%)',
                                    background: 'none',
                                    border: 'none',
                                    color: 'var(--text-tertiary)',
                                    cursor: 'pointer',
                                    padding: '4px',
                                    display: 'flex',
                                    alignItems: 'center'
                                }}
                            >
                                {showModalApiKey ? <EyeOff size={16} /> : <Eye size={16} />}
                            </button>
                        </div>
                        {modalMode === 'edit' && (
                            <p style={{ fontSize: '12px', color: 'var(--text-tertiary)', marginTop: '6px', margin: 0 }}>
                                {t.accounts.apiKeyKeepBlank}
                            </p>
                        )}
                        {formData.provider_type === 'tencentcloud' && (
                            <p style={{ fontSize: '12px', color: 'var(--text-secondary)', marginTop: '6px', margin: 0 }}>
                                {t.accounts.tencentcloudFormat}
                            </p>
                        )}
                        {formData.provider_type === 'cloudflare' && (
                            <p style={{ fontSize: '12px', color: 'var(--text-secondary)', marginTop: '6px', margin: 0 }}>
                                {t.accounts.cloudflareFormat}
                            </p>
                        )}
                        {formData.provider_type === 'ndjp' && (
                            <p style={{ fontSize: '12px', color: 'var(--text-secondary)', marginTop: '6px', margin: 0 }}>
                                {t.accounts.ndjpFormat || 'Bearer Token from NDJP NET dashboard'}
                            </p>
                        )}
                        {formData.provider_type === 'desec' && (
                            <p style={{ fontSize: '12px', color: 'var(--text-secondary)', marginTop: '6px', margin: 0 }}>
                                {t.accounts.desecFormat || 'Token from deSEC dashboard'}
                            </p>
                        )}
                        {formData.provider_type === 'dnshe' && (
                            <p style={{ fontSize: '12px', color: 'var(--text-secondary)', marginTop: '6px', margin: 0 }}>
                                {t.accounts.dnsheFormat || 'Format: API Key,API Secret (comma separated)'}
                            </p>
                        )}
                    </div>

                    <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px', marginTop: '24px', paddingTop: '20px', borderTop: '1px solid var(--border-color)' }}>
                        <button type="button" onClick={() => setIsModalOpen(false)} className="btn btn-secondary">
                            {t.common.cancel}
                        </button>
                        <button type="submit" className="btn btn-primary" disabled={submitting}>
                            {submitting ? (
                                <div className="spinner"></div>
                            ) : (
                                modalMode === 'create' ? t.accounts.linkAccount : t.accounts.saveChanges
                            )}
                        </button>
                    </div>
                </form>
            </Modal>

            {/* Delete Confirmation Dialog */}
            <ConfirmDialog
                isOpen={showDeleteConfirm}
                onClose={cancelDelete}
                onConfirm={confirmDelete}
                title={t.accounts.deleteAccountTitle}
                message={
                    deletingAccount
                        ? t.accounts.deleteAccountMessage.replace('{name}', deletingAccount.name)
                        : t.accounts.deleteAccountMessageDefault
                }
                confirmText={t.common.delete}
                cancelText={t.common.cancel}
                loading={deleting}
                danger={true}
            />

            {/* DDNS Token Management Modal */}
            <Modal
                isOpen={ddnsTokenModalOpen}
                onClose={() => setDdnsTokenModalOpen(false)}
                title={t.accounts.ddnsTokenTitle}
            >
                <div style={{ marginBottom: '16px' }}>
                    <p style={{ color: 'var(--text-secondary)', fontSize: '14px', margin: 0 }}>
                        {t.accounts.ddnsTokenSubtitle}
                    </p>
                </div>

                {ddnsTokensLoading ? (
                    <div className="spinner" style={{ margin: '2rem auto' }}></div>
                ) : (
                    <div style={{ maxHeight: '400px', overflowY: 'auto' }}>
                            <div className="glass-panel" style={{ padding: '16px' }}>
                            {ddnsToken?.has_token ? (
                                <>
                                    <div style={{ marginBottom: '12px', padding: '10px 12px', background: 'rgba(59, 130, 246, 0.1)', borderRadius: 'var(--radius-sm)', border: '1px solid rgba(59, 130, 246, 0.2)' }}>
                                        <div style={{ fontSize: '12px', color: 'var(--text-secondary)', marginBottom: '4px' }}>
                                            {t.accounts.ddnsTokenUsageExample}
                                        </div>
                                        <code style={{ fontSize: '11px', color: '#3b82f6', wordBreak: 'break-all' }}>
                                            {`${window.location.origin}/api/ddns/update?domains=example.com&token=${ddnsToken.token.token.substring(0, 8)}...`}
                                        </code>
                                    </div>
                                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '12px' }}>
                                        <div style={{ fontWeight: '500', fontSize: '14px' }}>
                                            {t.accounts.ddnsTokenName}
                                        </div>
                                        <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                                            <span className={`badge ${ddnsToken.token.enabled ? 'badge-success' : 'badge-neutral'}`} style={{ fontSize: '10px' }}>
                                                {ddnsToken.token.enabled ? t.accounts.ddnsTokenEnabled : t.accounts.ddnsTokenDisabled}
                                            </span>
                                        </div>
                                    </div>

                                    <div style={{
                                        padding: '12px',
                                        background: 'var(--bg-tertiary)',
                                        border: '1px solid var(--border-color)',
                                        borderRadius: 'var(--radius-sm)',
                                        marginBottom: '12px'
                                    }}>
                                        <div style={{
                                            fontSize: '11px',
                                            color: 'var(--text-tertiary)',
                                            marginBottom: '4px'
                                        }}>
                                            Token
                                        </div>
                                        <div style={{
                                            display: 'flex',
                                            alignItems: 'center',
                                            gap: '6px'
                                        }}>
                                            <div style={{
                                                flex: 1,
                                                fontSize: '13px',
                                                color: 'var(--text-primary)',
                                                fontFamily: 'monospace',
                                                wordBreak: 'break-all',
                                                lineHeight: '1.5'
                                            }}>
                                                {showDdnsTokenValue ? ddnsToken.token.token : '••••••••••••••••••••••••••••••••'}
                                            </div>
                                            <button
                                                onClick={() => setShowDdnsTokenValue(!showDdnsTokenValue)}
                                                className="btn btn-ghost"
                                                style={{ padding: '4px', minWidth: 'auto', height: 'auto', flexShrink: 0 }}
                                                title={showDdnsTokenValue ? t.common.hide : t.common.show}
                                            >
                                                {showDdnsTokenValue ? <EyeOff size={14} /> : <Eye size={14} />}
                                            </button>
                                        </div>
                                    </div>

                                    <div style={{ display: 'flex', gap: '6px', flexWrap: 'wrap', marginBottom: '8px' }}>
                                        <button
                                            onClick={() => copyDdnsToken(ddnsToken.token.token)}
                                            className="btn btn-ghost"
                                            style={{ fontSize: '11px', padding: '4px 8px' }}
                                            title={t.accounts.ddnsTokenCopy}
                                        >
                                            <Copy size={12} /> {t.accounts.ddnsTokenCopy}
                                        </button>
                                        <button
                                            onClick={() => copyDdnsToken(`${window.location.origin}/api/ddns/update?domains=<domain>&token=${ddnsToken.token.token}`)}
                                            className="btn btn-ghost"
                                            style={{ fontSize: '11px', padding: '4px 8px' }}
                                            title={t.accounts.ddnsTokenCopyUrl}
                                        >
                                            <Copy size={12} /> {t.accounts.ddnsTokenCopyUrl}
                                        </button>
                                    </div>

                                    {ddnsToken.token.last_used_at && (
                                        <div style={{ fontSize: '11px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                                            {t.accounts.ddnsTokenLastUsed}: {new Date(ddnsToken.token.last_used_at).toLocaleString()} | {t.accounts.ddnsTokenLastIP}: {ddnsToken.token.last_ip || '-'}
                                        </div>
                                    )}

                                    <div style={{ display: 'flex', gap: '6px', justifyContent: 'flex-end' }}>
                                        <button
                                            onClick={openDdnsTokenEdit}
                                            className="btn btn-ghost"
                                            style={{ fontSize: '11px', padding: '4px 8px' }}
                                        >
                                            <RefreshCw size={12} /> {t.common.edit}
                                        </button>
                                        <button
                                            onClick={confirmDdnsTokenDelete}
                                            className="btn btn-ghost"
                                            style={{ fontSize: '11px', padding: '4px 8px', color: 'var(--danger)' }}
                                        >
                                            <Trash2 size={12} /> {t.common.delete}
                                        </button>
                                    </div>
                                </>
                            ) : (
                                <div style={{ textAlign: 'center', padding: '32px', color: 'var(--text-tertiary)' }}>
                                    <Key size={48} style={{ marginBottom: '16px', opacity: 0.5 }} />
                                    <p style={{ margin: '0 0 8px 0' }}>{t.accounts.ddnsTokenNoTokens}</p>
                                    <p style={{ fontSize: '13px', margin: '0 0 16px 0' }}>{t.accounts.ddnsTokenCreateHint}</p>
                                    <button
                                        onClick={openDdnsTokenCreate}
                                        className="btn btn-secondary"
                                        style={{ fontSize: '13px', padding: '6px 12px' }}
                                    >
                                        <Key size={14} /> {t.accounts.ddnsTokenGenerate}
                                    </button>
                                </div>
                            )}
                        </div>
                    </div>
                )}
            </Modal>

            {/* DDNS Token Edit Modal */}
            <Modal
                isOpen={ddnsTokenEditModal}
                onClose={() => setDdnsTokenEditModal(false)}
                title={ddnsTokenForm.token ? t.common.edit : t.accounts.ddnsTokenGenerate}
            >
                <form onSubmit={handleDdnsTokenSubmit}>
                    {ddnsTokenFormError && (
                        <div style={{
                            backgroundColor: 'rgba(238, 0, 0, 0.1)',
                            color: 'var(--danger)',
                            padding: '12px',
                            borderRadius: 'var(--radius-sm)',
                            marginBottom: '20px',
                            fontSize: '13px',
                            border: '1px solid rgba(238, 0, 0, 0.2)'
                        }}>
                            {ddnsTokenFormError}
                        </div>
                    )}

                    <div className="form-group">
                        <label className="form-label">{t.accounts.ddnsTokenValue}</label>
                        <div style={{ position: 'relative' }}>
                            <input
                                type={showDdnsTokenValue ? "text" : "password"}
                                className="form-input"
                                placeholder={t.accounts.ddnsTokenCustom}
                                value={ddnsTokenForm.token}
                                onChange={e => setDdnsTokenForm({ ...ddnsTokenForm, token: e.target.value })}
                                style={{ paddingRight: '40px' }}
                                required
                            />
                            <button
                                type="button"
                                onClick={() => setShowDdnsTokenValue(!showDdnsTokenValue)}
                                style={{
                                    position: 'absolute',
                                    right: '8px',
                                    top: '50%',
                                    transform: 'translateY(-50%)',
                                    background: 'none',
                                    border: 'none',
                                    color: 'var(--text-tertiary)',
                                    cursor: 'pointer',
                                    padding: '4px',
                                    display: 'flex',
                                    alignItems: 'center'
                                }}
                            >
                                {showDdnsTokenValue ? <EyeOff size={16} /> : <Eye size={16} />}
                            </button>
                        </div>
                    </div>

                    <div className="form-group">
                        <label style={{ display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer' }}>
                            <input
                                type="checkbox"
                                checked={ddnsTokenForm.enabled}
                                onChange={e => setDdnsTokenForm({ ...ddnsTokenForm, enabled: e.target.checked })}
                            />
                            <span style={{ fontSize: '14px' }}>{t.accounts.ddnsTokenStatus}</span>
                        </label>
                    </div>

                    <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px', marginTop: '24px', paddingTop: '20px', borderTop: '1px solid var(--border-color)' }}>
                        <button type="button" onClick={() => setDdnsTokenEditModal(false)} className="btn btn-secondary">
                            {t.common.cancel}
                        </button>
                        <button type="submit" className="btn btn-primary" disabled={ddnsTokenSubmitting}>
                            {ddnsTokenSubmitting ? (
                                <div className="spinner"></div>
                            ) : (
                                t.accounts.saveChanges
                            )}
                        </button>
                    </div>
                </form>
            </Modal>

            {/* DDNS Token Delete Confirmation */}
            <ConfirmDialog
                isOpen={ddnsTokenDeleteConfirm}
                onClose={() => setDdnsTokenDeleteConfirm(false)}
                onConfirm={handleDdnsTokenDelete}
                title={t.accounts.ddnsTokenConfirmDelete}
                message={t.accounts.ddnsTokenConfirmDeleteMsg}
                confirmText={t.common.delete}
                cancelText={t.common.cancel}
                loading={ddnsTokenDeleteLoading}
                danger={true}
            />
        </div>
    );
};

export default Accounts;
