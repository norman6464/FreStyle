import InputField from '../components/InputField';
import TextareaField from '../components/TextareaField';
import PrimaryButton from '../components/PrimaryButton';
import FormMessage from '../components/FormMessage';
import { useProfileEdit } from '../hooks/useProfileEdit';

export default function ProfilePage() {
  const { form, message, loading, updateField, handleUpdate } = useProfileEdit();

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary-500" />
      </div>
    );
  }

  return (
    <div className="p-6 max-w-2xl mx-auto">
      <FormMessage message={message} />

      <div className="bg-surface-1 rounded-lg border border-surface-3 p-6">
        <div className="flex items-center gap-4 mb-6">
          <div className="w-16 h-16 bg-primary-500 rounded-full flex items-center justify-center">
            <span className="text-white text-2xl font-bold">
              {form.name?.charAt(0)?.toUpperCase() || 'U'}
            </span>
          </div>
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
