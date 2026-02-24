import { useEffect, useRef } from 'react';
import { X } from 'lucide-react';

const Modal = ({ isOpen, onClose, title, children }) => {
    const modalRef = useRef(null);

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
        if (modalRef.current && !modalRef.current.contains(e.target)) {
            onClose();
        }
    };

    if (!isOpen) return null;

    return (
        <div className="modal-overlay" onClick={handleBackdropClick}>
            <div className="modal-container" ref={modalRef}>
                <div style={{ 
                    display: 'flex', 
                    justifyContent: 'space-between', 
                    alignItems: 'center', 
                    marginBottom: '20px',
                    paddingBottom: '16px',
                    borderBottom: '1px solid var(--border-color)'
                }}>
                    <h3 style={{ fontSize: '18px', fontWeight: '600', margin: 0 }}>{title}</h3>
                    <button 
                        onClick={onClose} 
                        className="btn btn-ghost"
                        style={{ 
                            color: 'var(--text-secondary)',
                            padding: '4px',
                            minWidth: 'auto',
                            height: 'auto'
                        }}
                    >
                        <X size={18} />
                    </button>
                </div>
                {children}
            </div>
        </div>
    );
};

export default Modal;
