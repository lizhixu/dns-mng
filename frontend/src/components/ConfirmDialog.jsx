import { useEffect, useRef } from 'react';
import { AlertTriangle, X } from 'lucide-react';

const ConfirmDialog = ({ 
    isOpen, 
    onClose, 
    onConfirm, 
    title = '确认删除', 
    message = '确定要删除此项吗？此操作无法撤销。',
    confirmText = '删除',
    cancelText = '取消',
    loading = false,
    danger = true
}) => {
    const dialogRef = useRef(null);

    useEffect(() => {
        const handleEscape = (e) => {
            if (e.key === 'Escape') onClose();
        };

        if (isOpen) {
            document.addEventListener('keydown', handleEscape);
            document.body.style.overflow = 'hidden';
        }

        return () => {
            document.removeEventListener('keydown', handleEscape);
            document.body.style.overflow = 'unset';
        };
    }, [isOpen, onClose]);

    const handleBackdropClick = (e) => {
        if (dialogRef.current && !dialogRef.current.contains(e.target)) {
            onClose();
        }
    };

    const handleConfirm = () => {
        if (!loading) {
            onConfirm();
        }
    };

    if (!isOpen) return null;

    return (
        <div className="modal-overlay" onClick={handleBackdropClick}>
            <div 
                className="confirm-dialog" 
                ref={dialogRef}
                style={{
                    background: 'var(--bg-primary)',
                    border: '1px solid var(--border-color)',
                    borderRadius: 'var(--radius-lg)',
                    padding: '24px',
                    width: '100%',
                    maxWidth: '400px',
                    boxShadow: 'var(--shadow-lg)',
                    animation: 'scaleIn 0.15s ease-out'
                }}
            >
                <div style={{ 
                    display: 'flex', 
                    alignItems: 'flex-start', 
                    gap: '16px',
                    marginBottom: '20px'
                }}>
                    <div style={{
                        width: '40px',
                        height: '40px',
                        borderRadius: '50%',
                        background: danger ? 'rgba(239, 68, 68, 0.1)' : 'rgba(0, 112, 243, 0.1)',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        flexShrink: 0
                    }}>
                        <AlertTriangle 
                            size={20} 
                            style={{ 
                                color: danger ? 'var(--danger)' : 'var(--accent-primary)' 
                            }} 
                        />
                    </div>
                    <div style={{ flex: 1 }}>
                        <h3 style={{ 
                            fontSize: '16px', 
                            fontWeight: '600', 
                            margin: '0 0 8px 0',
                            color: 'var(--text-primary)'
                        }}>
                            {title}
                        </h3>
                        <p style={{ 
                            fontSize: '14px', 
                            margin: 0,
                            color: 'var(--text-secondary)',
                            lineHeight: 1.5
                        }}>
                            {message}
                        </p>
                    </div>
                </div>

                <div style={{ 
                    display: 'flex', 
                    justifyContent: 'flex-end', 
                    gap: '12px',
                    marginTop: '24px'
                }}>
                    <button
                        onClick={onClose}
                        disabled={loading}
                        style={{
                            padding: '8px 16px',
                            borderRadius: 'var(--radius-sm)',
                            border: '1px solid var(--border-color)',
                            background: 'transparent',
                            color: 'var(--text-secondary)',
                            fontSize: '14px',
                            fontWeight: '500',
                            cursor: loading ? 'not-allowed' : 'pointer',
                            opacity: loading ? 0.5 : 1,
                            transition: 'var(--transition)'
                        }}
                    >
                        {cancelText}
                    </button>
                    <button
                        onClick={handleConfirm}
                        disabled={loading}
                        style={{
                            padding: '8px 16px',
                            borderRadius: 'var(--radius-sm)',
                            border: 'none',
                            background: danger ? 'var(--danger)' : 'var(--accent-primary)',
                            color: '#fff',
                            fontSize: '14px',
                            fontWeight: '500',
                            cursor: loading ? 'not-allowed' : 'pointer',
                            opacity: loading ? 0.7 : 1,
                            transition: 'var(--transition)',
                            display: 'flex',
                            alignItems: 'center',
                            gap: '8px'
                        }}
                    >
                        {loading && (
                            <div style={{
                                width: '14px',
                                height: '14px',
                                border: '2px solid transparent',
                                borderTopColor: '#fff',
                                borderRadius: '50%',
                                animation: 'spin 0.6s linear infinite'
                            }} />
                        )}
                        {confirmText}
                    </button>
                </div>
            </div>
        </div>
    );
};

export default ConfirmDialog;