import Toast from '@/shared/ui/Toast';
import { useToast } from '@/hooks/useToast';

/**
 * 画面上部中央に Toast を積む。 pointer-events-none で本体クリックを邪魔しないように
 * しつつ、 Toast 1 つ 1 つは pointer-events-auto で閉じるボタンを押せる。
 */
export default function ToastContainer() {
  const { toasts, removeToast } = useToast();

  if (toasts.length === 0) return null;

  return (
    <div
      aria-live="polite"
      className="pointer-events-none fixed top-4 left-1/2 -translate-x-1/2 z-50 flex flex-col items-center gap-2 w-full max-w-md px-4"
    >
      {toasts.map((toast) => (
        <Toast
          key={toast.id}
          type={toast.type}
          message={toast.message}
          onClose={() => removeToast(toast.id)}
        />
      ))}
    </div>
  );
}
