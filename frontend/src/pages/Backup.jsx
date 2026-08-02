import { useState, useRef } from 'react';
import { useLanguage } from '../LanguageContext';
import { api } from '../api';
import { Download, Upload, AlertCircle, CheckCircle } from 'lucide-react';

const MAX_BACKUP_FILE_SIZE = 10 * 1024 * 1024;

export default function Backup() {
    const { t } = useLanguage();

    // ── 导出状态 ──────────────────────────────────────────────────
    const [exportPassword, setExportPassword] = useState('');
    const [exporting, setExporting] = useState(false);
    const [exportSuccess, setExportSuccess] = useState('');
    const [exportError, setExportError] = useState('');

    // ── 导入状态 ──────────────────────────────────────────────────
    const [importFile, setImportFile] = useState(null);
    const [importContent, setImportContent] = useState('');
    const [importPassword, setImportPassword] = useState('');
    const [importOverwrite, setImportOverwrite] = useState(false);
    const [importing, setImporting] = useState(false);
    const [importError, setImportError] = useState('');
    const [importResult, setImportResult] = useState(null);
    const [backupSummary, setBackupSummary] = useState(null);
    const [encryptedBackup, setEncryptedBackup] = useState(false);
    const fileInputRef = useRef(null);

    const clearImportState = () => {
        setImportFile(null);
        setImportContent('');
        setImportPassword('');
        setImportOverwrite(false);
        setBackupSummary(null);
        setEncryptedBackup(false);
        if (fileInputRef.current) {
            fileInputRef.current.value = '';
        }
    };

    const parseBackupSummary = (content) => {
        let parsed;
        try {
            parsed = JSON.parse(content);
        } catch {
            throw new Error(t.backup.invalidFormat);
        }

        if (parsed.encrypted) {
            return { encrypted: true, summary: null };
        }

        if (!parsed.data || typeof parsed.data !== 'object') {
            throw new Error(t.backup.invalidFormat);
        }

        const data = parsed.data;
        return {
            encrypted: false,
            summary: {
                accounts: data.accounts?.length || 0,
                domainCaches: data.domain_caches?.length || 0,
                hasDDNS: Boolean(data.ddns_token),
                hasEmail: Boolean(data.email_config),
                hasWHOIS: Boolean(data.whois_config),
                hasDNSHEAutoRenew: Boolean(data.dnshe_auto_renew),
                cfOptimize: data.cf_optimize_configs?.length || 0,
            },
        };
    };

    // ── 导出 ─────────────────────────────────────────────────────
    const handleExport = async () => {
        setExporting(true);
        setExportSuccess('');
        setExportError('');
        try {
            const { blob, filename } = await api.exportBackup(exportPassword);
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = filename;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);
            setExportPassword('');
            setExportSuccess(t.backup.exportSuccess);
            setTimeout(() => setExportSuccess(''), 3000);
        } catch (e) {
            setExportError(e.message || t.backup.exportError);
        } finally {
            setExporting(false);
        }
    };

    // ── 选择文件 ──────────────────────────────────────────────────
    const handleFileChange = (e) => {
        const file = e.target.files?.[0];
        setImportResult(null);
        setImportError('');
        setBackupSummary(null);
        setEncryptedBackup(false);
        setImportContent('');

        if (!file) return;
        if (file.size > MAX_BACKUP_FILE_SIZE) {
            setImportFile(null);
            setImportError(t.backup.fileTooLarge);
            if (fileInputRef.current) fileInputRef.current.value = '';
            return;
        }

        setImportFile(file);
        const reader = new FileReader();
        reader.onload = () => {
            const content = String(reader.result || '');
            try {
                const { encrypted, summary } = parseBackupSummary(content);
                setEncryptedBackup(encrypted);
                setBackupSummary(summary);
                setImportContent(content);
            } catch (err) {
                setImportFile(null);
                setImportError(err.message || t.backup.invalidFormat);
                if (fileInputRef.current) fileInputRef.current.value = '';
            }
        };
        reader.onerror = () => setImportError(t.backup.fileReadError);
        reader.onabort = () => setImportError(t.backup.fileReadError);
        reader.readAsText(file);
    };

    // ── 导入 ─────────────────────────────────────────────────────
    const handleImport = async () => {
        if (!importContent) return;
        if (importOverwrite && !window.confirm(t.backup.overwriteConfirm)) {
            return;
        }

        setImporting(true);
        setImportError('');
        setImportResult(null);
        try {
            const result = await api.importBackup({
                password: importPassword,
                overwrite: importOverwrite,
                content: importContent,
            });
            setImportResult(result);
            clearImportState();
        } catch (e) {
            setImportError(e.message || t.backup.restoreError);
        } finally {
            setImporting(false);
        }
    };

    const resultRows = importResult ? [
        [t.backup.resultAccounts, importResult.accounts_imported, importResult.accounts_skipped],
        [t.backup.resultDomainCaches, importResult.domain_caches_imported, importResult.domain_caches_skipped],
        [t.accounts.ddns.label, importResult.ddns_token_imported ? t.common.yes : t.common.no, importResult.ddns_token_skipped ? t.common.yes : t.common.no],
        [t.backup.resultEmailConfig, importResult.email_config_imported ? t.common.yes : t.common.no, importResult.email_config_skipped ? t.common.yes : t.common.no],
        [t.backup.resultWHOISConfig, importResult.whois_config_imported ? t.common.yes : t.common.no, importResult.whois_config_skipped ? t.common.yes : t.common.no],
        [t.backup.resultDNSHEAutoRenew, importResult.dnshe_auto_renew_imported ? t.common.yes : t.common.no, importResult.dnshe_auto_renew_skipped ? t.common.yes : t.common.no],
        [t.backup.resultCFOptimize, importResult.cf_optimize_imported || 0, importResult.cf_optimize_skipped || 0],
    ] : [];

    return (
        <div>
            {/* ── 页面标题 ──────────────────────────────────────── */}
            <div style={{ marginBottom: '1.5rem' }}>
                <h2 style={{ fontSize: '1.5rem', fontWeight: 'bold', letterSpacing: '-0.02em', margin: 0 }}>
                    {t.backup.title}
                </h2>
                <p style={{ color: 'var(--text-secondary)', fontSize: '13px', marginTop: '0.25rem', marginBottom: 0 }}>
                    {t.backup.subtitle}
                </p>
            </div>

            {/* ── 敏感信息提示 ──────────────────────────────────── */}
            <div style={{
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
                color: 'var(--danger)',
                marginBottom: '1.5rem',
                padding: '0.75rem 1rem',
                backgroundColor: 'rgba(255, 0, 0, 0.05)',
                border: '1px solid rgba(255, 0, 0, 0.15)',
                borderRadius: 'var(--radius-sm)',
                fontSize: '14px',
            }}>
                <AlertCircle size={16} />
                {t.backup.sensitiveWarning}
            </div>

            {/* ── 备份（导出）────────────────────────────────────── */}
            <div className="domain-list-card" style={{ padding: '1.25rem', marginBottom: '1.25rem', cursor: 'default' }}>
                <h3 style={{ fontSize: '15px', fontWeight: '600', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem', margin: '0 0 1rem 0' }}>
                    <Download size={16} />
                    {t.backup.exportCard}
                </h3>

                {exportSuccess && (
                    <div style={{
                        display: 'flex', alignItems: 'center', gap: '0.5rem',
                        color: 'var(--success)', marginBottom: '1rem', padding: '0.75rem 1rem',
                        backgroundColor: 'rgba(16, 185, 129, 0.05)', border: '1px solid rgba(16, 185, 129, 0.15)', borderRadius: 'var(--radius-sm)',
                        fontSize: '14px'
                    }}>
                        <CheckCircle size={16} />
                        {exportSuccess}
                    </div>
                )}
                {exportError && (
                    <div style={{
                        display: 'flex', alignItems: 'center', gap: '0.5rem',
                        color: 'var(--danger)', marginBottom: '1rem', padding: '0.75rem 1rem',
                        backgroundColor: 'rgba(255, 0, 0, 0.05)', border: '1px solid rgba(255, 0, 0, 0.15)', borderRadius: 'var(--radius-sm)',
                        fontSize: '14px'
                    }}>
                        <AlertCircle size={16} />
                        {exportError}
                    </div>
                )}

                <div className="form-group" style={{ marginBottom: 0 }}>
                    <label className="form-label">{t.backup.encryptionPassword}</label>
                    <input
                        type="password"
                        className="form-input"
                        placeholder={t.backup.encryptionPasswordPlaceholder}
                        value={exportPassword}
                        onChange={(e) => setExportPassword(e.target.value)}
                    />
                    <div style={{ fontSize: '12px', color: 'var(--text-tertiary)', marginTop: '0.375rem' }}>
                        {t.backup.encryptionPasswordHint}
                    </div>
                </div>

                <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: '1.25rem' }}>
                    <button className="btn btn-primary" style={{ height: '34px', fontSize: '13px' }} onClick={handleExport} disabled={exporting}>
                        {exporting ? (
                            <div className="spinner" style={{ width: '1rem', height: '1rem', borderWidth: '2px' }}></div>
                        ) : (
                            <><Download size={14} style={{ marginRight: '4px' }} />{t.backup.exportButton}</>
                        )}
                    </button>
                </div>
            </div>

            {/* ── 还原（导入）────────────────────────────────────── */}
            <div className="domain-list-card" style={{ padding: '1.25rem', cursor: 'default' }}>
                <h3 style={{ fontSize: '15px', fontWeight: '600', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem', margin: '0 0 1rem 0' }}>
                    <Upload size={16} />
                    {t.backup.importCard}
                </h3>

                {importError && (
                    <div style={{
                        display: 'flex', alignItems: 'center', gap: '0.5rem',
                        color: 'var(--danger)', marginBottom: '1rem', padding: '0.75rem 1rem',
                        backgroundColor: 'rgba(255, 0, 0, 0.05)', border: '1px solid rgba(255, 0, 0, 0.15)', borderRadius: 'var(--radius-sm)',
                        fontSize: '14px'
                    }}>
                        <AlertCircle size={16} />
                        {importError}
                    </div>
                )}

                {/* 文件选择 */}
                <div className="form-group">
                    <label className="form-label">{t.backup.selectFile}</label>
                    <input
                        ref={fileInputRef}
                        type="file"
                        accept=".json,application/json"
                        style={{ display: 'none' }}
                        onChange={handleFileChange}
                    />
                    <button
                        type="button"
                        className="btn btn-secondary"
                        onClick={() => fileInputRef.current?.click()}
                        style={{ width: '100%', justifyContent: 'flex-start', height: '34px', fontSize: '13px' }}
                    >
                        <Upload size={14} style={{ marginRight: '6px' }} />
                        {importFile
                            ? <span style={{ fontSize: '13px' }}>{t.common.selected}: <strong>{importFile.name}</strong></span>
                            : <span style={{ fontSize: '13px' }}>{t.backup.noFileSelected}</span>
                        }
                    </button>
                </div>

                {encryptedBackup && (
                    <div style={{ fontSize: '13px', color: 'var(--warning)', marginBottom: '1rem' }}>
                        {t.backup.encryptedFileHint}
                    </div>
                )}

                {backupSummary && (
                    <div style={{
                        backgroundColor: 'var(--bg-secondary)',
                        padding: '0.75rem 1rem',
                        borderRadius: 'var(--radius-sm)',
                        border: '1px solid var(--border-color)',
                        marginBottom: '1rem',
                        fontSize: '13px',
                    }}>
                        <div style={{ fontWeight: 600, marginBottom: '0.5rem' }}>{t.backup.summaryTitle}</div>
                        <div style={{ color: 'var(--text-secondary)', lineHeight: 1.8 }}>
                            {t.backup.summaryText
                                .replace('{accounts}', backupSummary.accounts)
                                .replace('{domainCaches}', backupSummary.domainCaches)
                                .replace('{ddns}', backupSummary.hasDDNS ? t.common.yes : t.common.no)
                                .replace('{email}', backupSummary.hasEmail ? t.common.yes : t.common.no)
                                .replace('{whois}', backupSummary.hasWHOIS ? t.common.yes : t.common.no)
                                .replace('{dnshe}', backupSummary.hasDNSHEAutoRenew ? t.common.yes : t.common.no)
                                .replace('{cfOptimize}', backupSummary.cfOptimize)}
                        </div>
                    </div>
                )}

                {/* 加密密码 */}
                <div className="form-group">
                    <label className="form-label">{t.backup.encryptionPassword}</label>
                    <input
                        type="password"
                        className="form-input"
                        placeholder={t.backup.restorePasswordPlaceholder}
                        value={importPassword}
                        onChange={(e) => setImportPassword(e.target.value)}
                    />
                </div>

                {/* 覆盖开关 */}
                <div className="form-group" style={{ marginBottom: 0 }}>
                    <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', cursor: 'pointer', fontSize: '13px', fontWeight: '500', color: 'var(--text-primary)' }}>
                        <input
                            type="checkbox"
                            checked={importOverwrite}
                            onChange={(e) => setImportOverwrite(e.target.checked)}
                            style={{ width: 'auto', cursor: 'pointer' }}
                        />
                        {t.backup.overwrite}
                    </label>
                    <div style={{ fontSize: '12px', color: 'var(--text-tertiary)', marginTop: '0.375rem' }}>
                        {t.backup.overwriteHint}
                    </div>
                </div>

                {/* 还原按钮 */}
                <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: '1.25rem' }}>
                    <button
                        className="btn btn-primary"
                        style={{ height: '34px', fontSize: '13px' }}
                        onClick={handleImport}
                        disabled={importing || !importContent}
                    >
                        {importing ? (
                            <div className="spinner" style={{ width: '1rem', height: '1rem', borderWidth: '2px' }}></div>
                        ) : (
                            <><Upload size={14} style={{ marginRight: '4px' }} />{t.backup.restoreButton}</>
                        )}
                    </button>
                </div>

                {/* 还原结果 */}
                {importResult && (
                    <div style={{
                        marginTop: '1.25rem',
                        backgroundColor: 'var(--bg-secondary)',
                        padding: '1rem',
                        borderRadius: 'var(--radius-sm)',
                        fontSize: '13px',
                        border: '1px solid var(--border-color)'
                    }}>
                        <h4 style={{
                            fontSize: '14px', fontWeight: '600', marginBottom: '0.75rem',
                            display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--success)',
                            margin: '0 0 0.75rem 0'
                        }}>
                            <CheckCircle size={14} />
                            {t.backup.resultTitle}
                        </h4>
                        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                            <thead>
                                <tr style={{ borderBottom: '1px solid var(--border-color)' }}>
                                    <th style={{ textAlign: 'left', padding: '0.375rem 0', color: 'var(--text-secondary)', fontWeight: 500 }}></th>
                                    <th style={{ textAlign: 'right', padding: '0.375rem 0', color: 'var(--text-secondary)', fontWeight: 500 }}>{t.backup.imported}</th>
                                    <th style={{ textAlign: 'right', padding: '0.375rem 0', color: 'var(--text-secondary)', fontWeight: 500 }}>{t.backup.skipped}</th>
                                </tr>
                            </thead>
                            <tbody>
                                {resultRows.map(([label, imported, skipped], index) => (
                                    <tr key={label} style={{ borderBottom: index === resultRows.length - 1 ? 'none' : '1px solid var(--border-color)' }}>
                                        <td style={{ padding: '0.375rem 0' }}>{label}</td>
                                        <td style={{ textAlign: 'right', padding: '0.375rem 0' }}>{imported}</td>
                                        <td style={{ textAlign: 'right', padding: '0.375rem 0' }}>{skipped}</td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                )}
            </div>
        </div>
    );
}
