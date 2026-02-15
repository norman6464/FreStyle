import { useState, useEffect, useCallback } from 'react';
import ProfileRepository from '../repositories/ProfileRepository';
import type { FormMessage } from '../types';

export function useProfileEdit() {
  const [form, setForm] = useState({ name: '', bio: '' });
  const [message, setMessage] = useState<FormMessage | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const loadProfile = async () => {
      try {
        const data = await ProfileRepository.fetchProfile();
        setForm({ name: data.name || '', bio: data.bio || '' });
      } catch {
        setMessage({ type: 'error', text: 'プロフィール取得に失敗しました。' });
      } finally {
        setLoading(false);
      }
    };
    loadProfile();
  }, []);

  const updateField = useCallback((field: 'name' | 'bio', value: string) => {
    setForm((prev) => ({ ...prev, [field]: value }));
  }, []);

  const handleUpdate = useCallback(async () => {
    try {
      const data = await ProfileRepository.updateProfile(form);
      setMessage({ type: 'success', text: data.success || 'プロフィールを更新しました。' });
    } catch {
      setMessage({ type: 'error', text: '通信エラーが発生しました。' });
    }
  }, [form]);

  return {
    form,
    message,
    loading,
    updateField,
    handleUpdate,
  };
}
