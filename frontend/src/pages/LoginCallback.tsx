import { useLoginCallback } from '../hooks/useLoginCallback';

export default function LoginCallback() {
  useLoginCallback();

  return (
    <div className="fixed inset-0 bg-white flex flex-col items-center justify-center z-50">
      <div className="w-10 h-10 border-4 border-gray-200 border-t-blue-500 rounded-full animate-spin" role="status" aria-label="読み込み中" />
      <p className="mt-4 text-sm text-gray-500">ログイン中...</p>
    </div>
  );
}
