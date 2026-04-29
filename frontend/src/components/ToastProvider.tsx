import { useCallback, useState, ReactNode } from 'react';
import { ToastContext, type ToastItem } from '../hooks/useToastContext';
import type { ToastType } from './Toast';

let toastId = 0;

/**
 * ToastProvider は ToastContext の React Context Provider。
 *
 * showToast / removeToast を Context として配り、`useToast()` hook 経由で
 * 任意の component から呼べるようにする。HMR を壊さないため hook と
 * 同居させず単体ファイルに切り出している。
 */
export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<ToastItem[]>([]);

  const showToast = useCallback((type: ToastType, message: string) => {
    const id = String(++toastId);
    setToasts((prev) => [...prev, { id, type, message }]);
  }, []);

  const removeToast = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  return (
    <ToastContext.Provider value={{ toasts, showToast, removeToast }}>
      {children}
    </ToastContext.Provider>
  );
}
