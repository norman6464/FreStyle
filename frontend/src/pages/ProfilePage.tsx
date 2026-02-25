import { useRef, useState, useEffect, useCallback, useMemo } from 'react';
import InputField from '../components/InputField';
import TextareaField from '../components/TextareaField';
import PrimaryButton from '../components/PrimaryButton';
import FormMessage from '../components/FormMessage';
import Avatar from '../components/Avatar';
import Loading from '../components/Loading';
import ProfileStatsSection from '../components/profile/ProfileStatsSection';
import ActivityHeatmap from '../components/ActivityHeatmap';
import { useProfileEdit } from '../hooks/useProfileEdit';
import { useProfileImageUpload } from '../hooks/useProfileImageUpload';
import { useScoreHistory } from '../hooks/useScoreHistory';
import ProfileStatsRepository, { type ProfileStats } from '../repositories/ProfileStatsRepository';
import { CameraIcon } from '@heroicons/react/24/outline';

export default function ProfilePage() {
  const { form, message, setMessage, loading, submitting, updateField, handleUpdate } = useProfileEdit();
  const { upload, uploading } = useProfileImageUpload();
  const { history } = useScoreHistory();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const practiceDates = useMemo(() => {
    return history.map((h) => h.createdAt.split('T')[0]);
  }, [history]);

  // 学習統計
  const [stats, setStats] = useState<ProfileStats | null>(null);
  const [statsLoading, setStatsLoading] = useState(true);

  const loadStats = useCallback(async () => {
    try {
      const data = await ProfileStatsRepository.fetchStats();
      setStats(data);
    } catch {
      // 統計取得エラーは無視
    } finally {
      setStatsLoading(false);
    }
  }, []);

  useEffect(() => {
    loadStats();
  }, [loadStats]);

  const handleImageSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const imageUrl = await upload(file);
    if (imageUrl) {
      updateField('iconUrl', imageUrl);
    } else {
      setMessage({ type: 'error', text: '画像のアップロードに失敗しました。' });
    }

    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  if (loading) {
    return <Loading className="h-full" />;
  }

  return (
    <div className="p-6 max-w-2xl mx-auto space-y-6">
      <FormMessage message={message} />

      {/* セクション1: 基本情報 */}
      <div className="bg-surface-1 rounded-lg border border-surface-3 p-6">
        <div className="flex items-center gap-4 mb-6">
          <div className="relative">
            <Avatar name={form.name || 'U'} src={form.iconUrl || undefined} size="xl" />
            <button
              type="button"
              onClick={() => fileInputRef.current?.click()}
              disabled={uploading}
              className="absolute bottom-0 right-0 bg-primary-500 text-white rounded-full p-1 hover:bg-primary-600 transition-colors disabled:opacity-50"
              aria-label="プロフィール画像を変更"
            >
              <CameraIcon className="w-4 h-4" />
            </button>
            <input
              ref={fileInputRef}
              type="file"
              accept="image/png,image/jpeg,image/gif,image/webp"
              onChange={handleImageSelect}
              className="hidden"
              data-testid="profile-image-input"
            />
          </div>
          <div>
            <h2 className="text-lg font-bold text-[var(--color-text-primary)]">プロフィールを編集</h2>
            <p className="text-sm text-[var(--color-text-muted)]">
              {uploading ? '画像をアップロード中...' : 'あなたの情報を更新してください'}
            </p>
          </div>
        </div>

        <form
          onSubmit={(e) => {
            e.preventDefault();
            handleUpdate();
          }}
          className="space-y-4"
        >
          <InputField
            label="ニックネーム"
            name="name"
            value={form.name}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => updateField('name', e.target.value)}
          />
          <TextareaField
            label="自己紹介"
            name="bio"
            value={form.bio}
            onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => updateField('bio', e.target.value)}
            placeholder="あなたについて教えてください..."
            rows={4}
            maxLength={200}
          />
          <InputField
            label="ステータス"
            name="status"
            value={form.status}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => updateField('status', e.target.value)}
            placeholder="例: 学習中、チャット可能、取り込み中..."
            maxLength={100}
          />
          <PrimaryButton type="submit" disabled={submitting}>
            {submitting ? '更新中...' : '基本情報を保存'}
          </PrimaryButton>
        </form>
      </div>

      {/* セクション2: 年間活動ヒートマップ */}
      <ActivityHeatmap practiceDates={practiceDates} />

      {/* セクション3: 学習統計 */}
      <ProfileStatsSection stats={stats} loading={statsLoading} />
    </div>
  );
}
