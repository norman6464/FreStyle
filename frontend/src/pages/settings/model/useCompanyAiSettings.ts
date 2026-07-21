import { useCallback, useEffect, useState } from 'react';
import { useAppDispatch } from '@/shared/lib/store';

import { CompanySettingsRepository } from '@/entities/company';
import { setAiChatEnabledForTrainees } from '@/entities/user';
import { classifyApiError } from '@/shared/lib/classifyApiError';

/**
 * 会社の AI 設定（trainee への AI 有効化）を読み書きするフック。
 * 取得・更新の状態を持ち、更新後は Redux にも反映して即座にサイドバー表示へ波及させる。
 */
export function useCompanyAiSettings() {
  const dispatch = useAppDispatch();
  const [enabled, setEnabled] = useState<boolean | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    CompanySettingsRepository.get()
      .then((s) => {
        if (cancelled) return;
        setEnabled(s.aiChatEnabledForTrainees);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(classifyApiError(err, '設定の取得に失敗しました。'));
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const update = useCallback(
    async (next: boolean): Promise<boolean> => {
      setSaving(true);
      setError(null);
      try {
        const s = await CompanySettingsRepository.update(next);
        setEnabled(s.aiChatEnabledForTrainees);
        dispatch(setAiChatEnabledForTrainees(s.aiChatEnabledForTrainees));
        return true;
      } catch (err) {
        setError(classifyApiError(err, '設定の更新に失敗しました。'));
        return false;
      } finally {
        setSaving(false);
      }
    },
    [dispatch]
  );

  return { enabled, loading, saving, error, update };
}
