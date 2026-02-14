import InputField from '../components/InputField';
import TextareaField from '../components/TextareaField';
import PrimaryButton from '../components/PrimaryButton';
import FormMessage from '../components/FormMessage';
import Avatar from '../components/Avatar';
import Loading from '../components/Loading';
import { useProfileEdit } from '../hooks/useProfileEdit';

export default function ProfilePage() {
  const { form, message, loading, updateField, handleUpdate } = useProfileEdit();

  if (loading) {
    return (
      <Loading className="h-full" />
    );
  }

  return (
    <div className="p-6 max-w-2xl mx-auto">
      <FormMessage message={message} />

      <div className="bg-surface-1 rounded-lg border border-surface-3 p-6">
        <div className="flex items-center gap-4 mb-6">
          <Avatar name={form.name || 'U'} size="xl" />
          <div>
            <h2 className="text-lg font-bold text-[var(--color-text-primary)]">プロフィールを編集</h2>
            <p className="text-sm text-[var(--color-text-muted)]">あなたの情報を更新してください</p>
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
            onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
              updateField('name', e.target.value)
            }
          />
          <TextareaField
            label="自己紹介"
            name="bio"
            value={form.bio}
            onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
              updateField('bio', e.target.value)
            }
            placeholder="あなたについて教えてください..."
            rows={4}
            maxLength={200}
          />
          <PrimaryButton type="submit">プロフィールを更新</PrimaryButton>
        </form>
      </div>
    </div>
  );
}
