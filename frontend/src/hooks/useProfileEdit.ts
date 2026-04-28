import { useState, useEffect, useCallback } from 'react';
import ProfileRepository from '../repositories/ProfileRepository';
import type { FormMessage, Profile } from '../types';

/**
 * useProfileEdit — ProfilePage で「ニックネーム / 自己紹介 / アイコン / ステータス」
 * の編集を扱う。フォーム形は backend `domain.ProfileView` のサブセット。
 */
type ProfileForm = Pick<Profile, 'displayName' | 'bio' | 'avatarUrl' | 'status'>;

const EMPTY_FORM: ProfileForm = {
  displayName: '',
  bio: '',
  avatarUrl: '',
  status: '',
};

export function useProfileEdit() {
  const [form, setForm] = useState<ProfileForm>(EMPTY_FORM);
  const [message, setMessage] = useState<FormMessage | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    const loadProfile = async () => {
      try {
        const data = await ProfileRepository.fetchProfile();
        setForm({
          displayName: data.displayName ?? '',
          bio: data.bio ?? '',
          avatarUrl: data.avatarUrl ?? '',
          status: data.status ?? '',
        });
      } catch {
        setMessage({ type: 'error', text: 'プロフィール取得に失敗しました。' });
      } finally {
        setLoading(false);
      }
    };
    loadProfile();
  }, []);

  const updateField = useCallback(<K extends keyof ProfileForm>(field: K, value: ProfileForm[K]) => {
    setForm((prev) => ({ ...prev, [field]: value }));
  }, []);

  const handleUpdate = useCallback(async () => {
    if (!form.displayName.trim()) {
      setMessage({ type: 'error', text: 'ニックネームを入力してください。' });
      return;
    }
    setSubmitting(true);
    try {
      await ProfileRepository.updateProfile(form);
      setMessage({ type: 'success', text: 'プロフィールを更新しました。' });
    } catch {
      setMessage({ type: 'error', text: '通信エラーが発生しました。' });
    } finally {
      setSubmitting(false);
    }
  }, [form]);

  return {
    form,
    message,
    setMessage,
    loading,
    submitting,
    updateField,
    handleUpdate,
  };
}
