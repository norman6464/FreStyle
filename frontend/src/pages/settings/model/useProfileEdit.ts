import { useState, useEffect, useCallback } from 'react';
import { ProfileRepository } from '@/entities/user';
import { useToast } from '@/shared/lib/hooks/useToast';
import type { FormMessage } from '@/shared/ui/FormMessage';
import type { Profile } from '@/entities/user';

/**
 * useProfileEdit — ProfilePage で「氏名 / 自己紹介 / アイコン / ステータス」
 * の編集を扱う。フォーム形は backend `domain.ProfileView` のサブセット。
 *
 * displayName は OIDC ログイン時に id_token の `name` claim を初期値として
 * セットするため（auth_handler.upsertUserFromIDToken）、 通常は氏名がそのまま入る。
 * ユーザは ProfilePage 上で自由に書き換え可能。
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
  // 失敗系（取得エラー / バリデーション）はインライン表示を維持。
  // 成功系（更新しました）は Toast で通知する（画面上部からバウンドで降りてくる）。
  const [message, setMessage] = useState<FormMessage | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const { showToast } = useToast();

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
      setMessage({ type: 'error', text: '氏名を入力してください。' });
      return;
    }
    setSubmitting(true);
    try {
      await ProfileRepository.updateProfile(form);
      // 成功時はインラインメッセージを消して Toast を出す。
      setMessage(null);
      showToast('success', 'プロフィールを更新しました。');
    } catch {
      setMessage({ type: 'error', text: '通信エラーが発生しました。' });
    } finally {
      setSubmitting(false);
    }
  }, [form, showToast]);

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
