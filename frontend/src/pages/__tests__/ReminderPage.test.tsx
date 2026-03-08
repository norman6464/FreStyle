import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import ReminderPage from '../ReminderPage';
import { useReminder } from '../../hooks/useReminder';

vi.mock('../../hooks/useReminder');
vi.mock('../../hooks/useToast', () => ({
  useToast: () => ({ showToast: vi.fn() }),
}));
const mockedUseReminder = vi.mocked(useReminder);

function renderPage() {
  return render(<MemoryRouter><ReminderPage /></MemoryRouter>);
}

describe('ReminderPage', () => {
  const defaultReturn = {
    setting: { enabled: true, reminderTime: '20:00', daysOfWeek: 'mon,tue,wed,thu,fri' },
    loading: false,
    saving: false,
    toggleEnabled: vi.fn(),
    setTime: vi.fn(),
    toggleDay: vi.fn(),
    save: vi.fn(),
    dayOptions: [{ key: 'mon', label: '月' }, { key: 'tue', label: '火' }],
    selectedDays: ['mon', 'tue'],
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockedUseReminder.mockReturnValue(defaultReturn);
  });

  it('タイトルを表示する', () => {
    renderPage();
    expect(screen.getByText('練習リマインダー')).toBeInTheDocument();
  });

  it('トグルスイッチを表示する', () => {
    renderPage();
    expect(screen.getByTestId('toggle-enabled')).toBeInTheDocument();
  });

  it('保存ボタンを表示する', () => {
    renderPage();
    expect(screen.getByTestId('save-button')).toBeInTheDocument();
  });

  it('ローディング中はローディング表示する', () => {
    mockedUseReminder.mockReturnValue({ ...defaultReturn, loading: true });
    renderPage();
    expect(screen.getByText('設定を読み込み中...')).toBeInTheDocument();
  });

  it('トグルクリックでtoggleEnabledが呼ばれる', () => {
    renderPage();
    fireEvent.click(screen.getByTestId('toggle-enabled'));
    expect(defaultReturn.toggleEnabled).toHaveBeenCalled();
  });
});
