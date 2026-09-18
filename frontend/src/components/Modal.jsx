import { useEffect, useRef } from 'react';
import { X } from 'lucide-react';

const Modal = ({ isOpen, onClose, title, children, footer, size = 'default', closeOnBackdrop = false, style }) => {
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
        if (closeOnBackdrop && modalRef.current && !modalRef.current.contains(e.target)) {
            onClose();
        }
    };

    if (!isOpen) return null;

    const sizeClass = size === 'large'
        ? 'modal-container-large'
        : (size === 'xl' || size === 'extra-large' ? 'modal-container-xl' : '');

    return (
        <div className="modal-overlay" onClick={handleBackdropClick}>
            <div className={`modal-container ${sizeClass}`} style={style} ref={modalRef}>
                {/* Fixed Header */}
                <div className="modal-header">
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
                
                {/* Scrollable Content */}
                <div className="modal-content">
                    {children}
                </div>
                
                {/* Fixed Footer */}
                {footer && (
                    <div className="modal-footer">
                        {footer}
                    </div>
                )}
            </div>
        </div>
    );
};

export default Modal;
