import { useContext } from 'react';
import { ToastContext } from './useToastContext';

/**
 * useToast は ToastContext から showToast / removeToast / toasts を取り出す hook。
 *
 * ToastProvider component は src/components/ToastProvider.tsx に分離した。
 * 同一ファイルで component + hook を export していると Vite React HMR の
 * react-refresh/only-export-components ルールに引っかかるため。
 *
 * 旧コードとの互換性を保つため `import { ToastProvider } from '../hooks/useToast'`
 * の参照は ../components/ToastProvider に書き換える必要がある。
 */
export function useToast() {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error('useToast must be used within a ToastProvider');
  }
  return context;
}
