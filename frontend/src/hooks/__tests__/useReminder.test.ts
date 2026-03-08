import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useReminder } from '../useReminder';
import { ReminderRepository } from '../../repositories/ReminderRepository';

vi.mock('../../repositories/ReminderRepository');
const mockedRepo = vi.mocked(ReminderRepository);

describe('useReminder', () => {
  const mockSetting = { enabled: true, reminderTime: '20:00', daysOfWeek: 'mon,tue,wed,thu,fri' };

  beforeEach(() => {
    vi.clearAllMocks();
    mockedRepo.fetchSetting.mockResolvedValue(mockSetting);
    mockedRepo.saveSetting.mockResolvedValue(mockSetting);
  });

  it('初期ロード時に設定を取得する', async () => {
    const { result } = renderHook(() => useReminder());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.setting).toEqual(mockSetting);
  });

  it('有効/無効を切り替える', async () => {
    const { result } = renderHook(() => useReminder());
    await waitFor(() => expect(result.current.loading).toBe(false));
    act(() => result.current.toggleEnabled());
    expect(result.current.setting.enabled).toBe(false);
  });

  it('時間を変更する', async () => {
    const { result } = renderHook(() => useReminder());
    await waitFor(() => expect(result.current.loading).toBe(false));
    act(() => result.current.setTime('19:30'));
    expect(result.current.setting.reminderTime).toBe('19:30');
  });

  it('曜日を切り替える', async () => {
    const { result } = renderHook(() => useReminder());
    await waitFor(() => expect(result.current.loading).toBe(false));
    act(() => result.current.toggleDay('sat'));
    expect(result.current.selectedDays).toContain('sat');
  });

  it('設定を保存する', async () => {
    const { result } = renderHook(() => useReminder());
    await waitFor(() => expect(result.current.loading).toBe(false));
    await act(async () => { await result.current.save(); });
    expect(mockedRepo.saveSetting).toHaveBeenCalled();
  });

  it('曜日をトグルで削除する', async () => {
    const { result } = renderHook(() => useReminder());
    await waitFor(() => expect(result.current.loading).toBe(false));

    // mockSettingのdaysOfWeekは 'mon,tue,wed,thu,fri' なので 'mon' は含まれている
    expect(result.current.selectedDays).toContain('mon');

    act(() => result.current.toggleDay('mon'));

    expect(result.current.selectedDays).not.toContain('mon');
    expect(result.current.setting.daysOfWeek).not.toContain('mon');
  });

  it('保存中フラグが正しく管理される', async () => {
    let resolveSave: (value: typeof mockSetting) => void;
    mockedRepo.saveSetting.mockImplementation(() => new Promise((resolve) => { resolveSave = resolve; }));

    const { result } = renderHook(() => useReminder());
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.saving).toBe(false);

    let savePromise: Promise<void>;
    act(() => { savePromise = result.current.save(); });

    // saving should be true while the promise is pending
    expect(result.current.saving).toBe(true);

    await act(async () => {
      resolveSave!(mockSetting);
      await savePromise!;
    });

    expect(result.current.saving).toBe(false);
  });
});
