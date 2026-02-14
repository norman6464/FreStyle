import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
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
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              自己紹介
            </label>
            <textarea
              name="bio"
              value={form.bio}
              onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
                updateField('bio', e.target.value)
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
