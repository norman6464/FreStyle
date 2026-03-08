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
});
