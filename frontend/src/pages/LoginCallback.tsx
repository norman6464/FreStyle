import { useLoginCallback } from '../hooks/useLoginCallback';

export default function LoginCallback() {
  useLoginCallback();

  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-surface">
      <div className="flex flex-col items-center space-y-6">
        <div className="relative w-20 h-20">
          <div className="absolute inset-0 rounded-full border-4 border-primary-200"></div>
          <div className="absolute inset-0 rounded-full border-4 border-transparent border-t-primary-600 border-r-primary-600 animate-spin"></div>
        </div>
        <div className="text-center">
          <h3 className="text-2xl font-bold text-gray-800 mb-2">
            ログイン中...
          </h3>
          <p className="text-gray-600">お待たせしています</p>
        </div>
      </div>
    </div>
  );
}
