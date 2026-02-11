import { useState, useEffect } from 'react';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import { useDispatch } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import { clearAuth } from '../store/authSlice';

interface FormMessage {
  type: 'success' | 'error';
  text: string;
}

export default function ProfilePage() {
  const [form, setForm] = useState({ name: '', bio: '' });
  const [message, setMessage] = useState<FormMessage | null>(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();
  const dispatch = useDispatch();

  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  const fetchProfile = async () => {
    try {
      const res = await fetch(`${API_BASE_URL}/api/profile/me`, {
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
      });

      if (res.status === 401) {
        const refreshRes = await fetch(
          `${API_BASE_URL}/api/auth/cognito/refresh-token`,
          { method: 'POST', credentials: 'include' }
        );
        if (!refreshRes.ok) {
          dispatch(clearAuth());
          return;
        }

        const retryRes = await fetch(`${API_BASE_URL}/api/profile/me`, {
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
        });
        const retryData = await retryRes.json();
        if (!retryRes.ok) throw new Error(retryData.error || 'プロフィール再取得に失敗しました。');

        setForm({ name: retryData.name || '', bio: retryData.bio || '' });
        setLoading(false);
        return;
      }

      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'プロフィール取得に失敗しました。');

      setForm({ name: data.name || '', bio: data.bio || '' });
    } catch (err) {
      setMessage({ type: 'error', text: (err as Error).message });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProfile();
  }, []);

  const handleUpdate = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    try {
      const res = await fetch(`${API_BASE_URL}/api/profile/me/update`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(form),
      });

      if (res.status === 401) {
        const refreshRes = await fetch(
          `${API_BASE_URL}/api/auth/cognito/refresh-token`,
          { method: 'POST', credentials: 'include' }
        );
        if (!refreshRes.ok) {
          navigate('/login');
          return;
        }
        await refreshRes.json();
        const retryRes = await fetch(`${API_BASE_URL}/api/profile/me/update`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify(form),
        });
        const retryData = await retryRes.json();
        if (!retryRes.ok) throw new Error(retryData.error || 'プロフィール更新に失敗しました。');

        setMessage({ type: 'success', text: retryData.success || 'プロフィールを更新しました。' });
        return;
      }

      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'プロフィール更新に失敗しました。');

      setMessage({ type: 'success', text: data.success || 'プロフィールを更新しました。' });
    } catch (error) {
      setMessage({ type: 'error', text: '通信エラーが発生しました。' });
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary-500" />
      </div>
    );
  }

  return (
    <div className="p-6 max-w-2xl mx-auto">
      {/* メッセージ */}
      {message && (
        <div
          className={`mb-4 p-3 rounded-lg text-sm font-medium ${
            message.type === 'error'
              ? 'bg-rose-50 text-rose-700 border border-rose-200'
              : 'bg-emerald-50 text-emerald-700 border border-emerald-200'
          }`}
        >
          {message.text}
        </div>
      )}

      {/* プロフィールカード */}
      <div className="bg-white rounded-lg border border-slate-200 p-6">
        <div className="flex items-center gap-4 mb-6">
          <div className="w-16 h-16 bg-primary-500 rounded-full flex items-center justify-center">
            <span className="text-white text-2xl font-bold">
              {form.name?.charAt(0)?.toUpperCase() || 'U'}
            </span>
          </div>
          <div>
            <h2 className="text-lg font-bold text-slate-800">プロフィールを編集</h2>
            <p className="text-sm text-slate-500">あなたの情報を更新してください</p>
          </div>
        </div>

        <form onSubmit={handleUpdate} className="space-y-4">
          <InputField
            label="ニックネーム"
            name="name"
            value={form.name}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
              setForm((prev) => ({ ...prev, name: e.target.value }))
            }
          />
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              自己紹介
            </label>
            <textarea
              name="bio"
              value={form.bio}
              onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
                setForm((prev) => ({ ...prev, bio: e.target.value }))
              }
              placeholder="あなたについて教えてください..."
              rows={4}
              className="w-full border border-slate-300 rounded-lg px-3 py-2 text-sm focus:border-primary-500 focus:ring-1 focus:ring-primary-500 transition-colors resize-none"
            />
          </div>
          <PrimaryButton type="submit">プロフィールを更新</PrimaryButton>
        </form>
      </div>
    </div>
  );
}
