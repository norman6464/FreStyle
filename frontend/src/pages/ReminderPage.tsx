import { useReminder } from '../hooks/useReminder';
import Loading from '../components/Loading';
import { useToast } from '../hooks/useToast';

export default function ReminderPage() {
  const { setting, loading, saving, toggleEnabled, setTime, toggleDay, save, dayOptions, selectedDays } = useReminder();
  const { showToast } = useToast();

  const handleSave = async () => {
    await save();
    showToast('success', 'リマインダー設定を保存しました');
  };

  if (loading) return <Loading message="設定を読み込み中..." />;

  return (
    <div className="max-w-lg mx-auto px-4 py-6 space-y-6">
      <h1 className="text-xl font-bold">練習リマインダー</h1>

      <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4 space-y-4">
        {/* 有効/無効 */}
        <div className="flex items-center justify-between">
          <span className="font-medium">リマインダー</span>
          <button
            onClick={toggleEnabled}
            className={`relative w-12 h-6 rounded-full transition-colors ${
              setting.enabled ? 'bg-blue-500' : 'bg-gray-300 dark:bg-gray-600'
            }`}
            data-testid="toggle-enabled"
            role="switch"
            aria-checked={setting.enabled}
          >
            <span
              className={`absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full transition-transform ${
                setting.enabled ? 'translate-x-6' : ''
              }`}
            />
          </button>
        </div>

        {/* 時間設定 */}
        <div className="space-y-1">
          <label className="text-sm text-gray-500">通知時間</label>
          <input
            type="time"
            value={setting.reminderTime}
            onChange={(e) => setTime(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700"
            data-testid="time-input"
          />
        </div>

        {/* 曜日設定 */}
        <div className="space-y-1">
          <label className="text-sm text-gray-500">通知する曜日</label>
          <div className="flex gap-2">
            {dayOptions.map((day) => (
              <button
                key={day.key}
                onClick={() => toggleDay(day.key)}
                className={`w-9 h-9 rounded-full text-sm font-medium transition-colors ${
                  selectedDays.includes(day.key)
                    ? 'bg-blue-500 text-white'
                    : 'bg-gray-100 dark:bg-gray-700 text-gray-500'
                }`}
                data-testid={`day-${day.key}`}
              >
                {day.label}
              </button>
            ))}
          </div>
        </div>

        {/* 保存ボタン */}
        <button
          onClick={handleSave}
          disabled={saving}
          className="w-full px-4 py-2 bg-blue-500 text-white rounded-lg font-medium hover:bg-blue-600 disabled:opacity-50 transition-colors"
          data-testid="save-button"
        >
          {saving ? '保存中...' : '設定を保存'}
        </button>
      </div>
    </div>
  );
}
