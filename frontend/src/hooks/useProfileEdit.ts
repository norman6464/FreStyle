import { useState, useEffect, useCallback } from 'react';
import ProfileRepository from '../repositories/ProfileRepository';
import type { FormMessage } from '../types';

export function useProfileEdit() {
  const [form, setForm] = useState({ name: '', bio: '', iconUrl: '' });
  const [message, setMessage] = useState<FormMessage | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    const loadProfile = async () => {
      try {
        const data = await ProfileRepository.fetchProfile();
        setForm({ name: data.name || '', bio: data.bio || '', iconUrl: data.iconUrl || '' });
      } catch {
        setMessage({ type: 'error', text: 'プロフィール取得に失敗しました。' });
      } finally {
        setLoading(false);
      }
    };
    loadProfile();
  }, []);

  const updateField = useCallback((field: 'name' | 'bio' | 'iconUrl', value: string) => {
    setForm((prev) => ({ ...prev, [field]: value }));
  }, []);

  const handleUpdate = useCallback(async () => {
    if (!form.name.trim()) {
      setMessage({ type: 'error', text: 'ニックネームを入力してください。' });
      return;
    }

    setSubmitting(true);
    try {
      const data = await ProfileRepository.updateProfile(form);
      setMessage({ type: 'success', text: data.success || 'プロフィールを更新しました。' });
    } catch {
      setMessage({ type: 'error', text: '通信エラーが発生しました。' });
    } finally {
      setSubmitting(false);
    }
  }, [form]);

  return {
    form,
    message,
    loading,
    submitting,
    updateField,
    handleUpdate,
  };
}
