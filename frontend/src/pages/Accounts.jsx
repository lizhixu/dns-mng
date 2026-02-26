import { useState, useEffect } from 'react';
import { api } from '../api';
import { Link } from 'react-router-dom';
import { Plus, Settings, Trash2, ExternalLink, Eye, EyeOff, Copy } from 'lucide-react';
import Modal from '../components/Modal';
import ConfirmDialog from '../components/ConfirmDialog';
import { useLanguage } from '../LanguageContext';

const Accounts = () => {
    const { t } = useLanguage();
    const [accounts, setAccounts] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

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

    const toggleKeyVisibility = (id) => {
        setVisibleKeys(prev => ({
            ...prev,
            [id]: !prev[id]
        }));
    };

    useEffect(() => {
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

    if (loading) return <div className="spinner" style={{ margin: '2rem auto' }}></div>;
    if (error) return <div style={{ color: 'var(--danger)', padding: '1rem' }}>{t.common.error}: {error}</div>;

    return (
        <div>
            <div style={{ marginBottom: '32px' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '8px' }}>
                    <h1 style={{ fontSize: '24px', fontWeight: '600', margin: 0 }}>{t.accounts.title}</h1>
                    <button onClick={openCreateModal} className="btn btn-primary">
                        <Plus size={16} />
                        {t.accounts.addAccount}
                    </button>
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
                                justifyContent: 'center'
                            }}
                        >
                            {t.accounts.viewDomains}
                            <ExternalLink size={13} />
                        </Link>
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
                                        : formData.provider_type === 'cloudflare'
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
        </div>
    );
};

export default Accounts;
