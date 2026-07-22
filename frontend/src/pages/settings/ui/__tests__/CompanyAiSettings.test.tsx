import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import authReducer from '@/entities/user/model/authSlice';
import CompanyAiSettings from '../CompanyAiSettings';
import { CompanySettingsRepository } from '@/entities/company';

vi.mock('@/entities/company/api/companySettingsRepository', () => ({
  default: {
    get: vi.fn(),
    update: vi.fn(),
  },
}));

const mockRepo = vi.mocked(CompanySettingsRepository);

function renderComponent() {
  const store = configureStore({ reducer: { auth: authReducer } });
  return render(
    <Provider store={store}>
      <CompanyAiSettings />
    </Provider>
  );
}

describe('CompanyAiSettings', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('現在の設定を読み込んでチェックボックスに反映する', async () => {
    mockRepo.get.mockResolvedValue({ aiChatEnabledForTrainees: true });
    renderComponent();
    const checkbox = await screen.findByRole('checkbox');
    await waitFor(() => expect((checkbox as HTMLInputElement).checked).toBe(true));
  });

  it('チェックを外すと update(false) が呼ばれ反映される', async () => {
    mockRepo.get.mockResolvedValue({ aiChatEnabledForTrainees: true });
    mockRepo.update.mockResolvedValue({ aiChatEnabledForTrainees: false });
    renderComponent();

    const checkbox = (await screen.findByRole('checkbox')) as HTMLInputElement;
    await waitFor(() => expect(checkbox.checked).toBe(true));

    fireEvent.click(checkbox);

    await waitFor(() => expect(mockRepo.update).toHaveBeenCalledWith(false));
    await waitFor(() => expect(checkbox.checked).toBe(false));
  });

  it('取得失敗時はエラーメッセージを表示する', async () => {
    mockRepo.get.mockRejectedValue(new Error('boom'));
    renderComponent();
    expect(await screen.findByText(/設定の取得に失敗しました/)).toBeInTheDocument();
  });
});
