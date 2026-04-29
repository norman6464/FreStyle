import { createContext } from 'react';
import type { ToastType } from '../components/Toast';

export interface ToastItem {
  id: string;
  type: ToastType;
  message: string;
}

export interface ToastContextValue {
  toasts: ToastItem[];
  showToast: (type: ToastType, message: string) => void;
  removeToast: (id: string) => void;
}

/**
 * ToastProvider と useToast hook で共有する React Context。
 *
 * ToastProvider (component) と useToast (hook) を同一ファイルから export すると
 * react-refresh/only-export-components のルールに抵触し HMR が壊れるため、
 * Context オブジェクトをこの専用ファイルに切り出した。
 */
export const ToastContext = createContext<ToastContextValue | null>(null);
