import { useState, useEffect } from 'react';
import { useUserProfile } from './useUserProfile';

interface UserProfileForm {
  displayName: string;
  selfIntroduction: string;
  communicationStyle: string;
  personalityTraits: string[];
  goals: string;
  concerns: string;
  preferredFeedbackStyle: string;
}

interface FormMessage {
  type: 'success' | 'error';
  text: string;
}

export const COMMUNICATION_STYLES = [
  { value: '', label: '選択してください' },
  { value: 'casual', label: 'カジュアル' },
  { value: 'formal', label: 'フォーマル' },
  { value: 'friendly', label: 'フレンドリー' },
  { value: 'professional', label: 'プロフェッショナル' },
] as const;

export const PERSONALITY_OPTIONS = [
  '内向的', '外向的', '論理的', '感情的', '共感力が高い',
  '分析的', 'クリエイティブ', '計画的', '柔軟性がある', 'リーダーシップがある',
] as const;

export const FEEDBACK_STYLES = [
  { value: '', label: '選択してください' },
  { value: 'direct', label: 'ストレート（はっきり伝えてほしい）' },
  { value: 'gentle', label: 'やさしく（配慮を持って伝えてほしい）' },
  { value: 'detailed', label: '詳細に（具体的に説明してほしい）' },
] as const;

/**
 * UserProfilePageフック
 *
 * <p>役割:</p>
 * <ul>
 *   <li>UserProfilePageのフォーム管理・バリデーション</li>
 *   <li>プロフィール取得・更新のロジック</li>
 * </ul>
 */
export function useUserProfilePage() {
  const [form, setForm] = useState<UserProfileForm>({
    displayName: '',
    selfIntroduction: '',
    communicationStyle: '',
    personalityTraits: [],
    goals: '',
    concerns: '',
    preferredFeedbackStyle: '',
  });
  const [message, setMessage] = useState<FormMessage | null>(null);
  const [isNewProfile, setIsNewProfile] = useState(false);

  const { profile, loading, fetchMyProfile, updateProfile } = useUserProfile();

  useEffect(() => {
    fetchMyProfile();
  }, [fetchMyProfile]);

  useEffect(() => {
    if (profile) {
      setForm({
        displayName: profile.displayName || '',
        selfIntroduction: profile.selfIntroduction || '',
        communicationStyle: profile.communicationStyle || '',
        personalityTraits: profile.personalityTraits || [],
        goals: profile.goals || '',
        concerns: profile.concerns || '',
        preferredFeedbackStyle: profile.preferredFeedbackStyle || '',
      });
      setIsNewProfile(false);
    }
  }, [profile]);

  const togglePersonalityTrait = (trait: string) => {
    setForm((prev) => {
      const traits = prev.personalityTraits.includes(trait)
        ? prev.personalityTraits.filter((t) => t !== trait)
        : [...prev.personalityTraits, trait];
      return { ...prev, personalityTraits: traits };
    });
  };

  const handleSave = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    const success = await updateProfile(form);

    if (success) {
      setMessage({ type: 'success', text: 'パーソナリティ設定を保存しました。' });
      setIsNewProfile(false);
    } else {
      setMessage({ type: 'error', text: '保存に失敗しました。' });
    }
  };

  return {
    form,
    setForm,
    message,
    isNewProfile,
    loading,
    togglePersonalityTrait,
    handleSave,
  };
}
