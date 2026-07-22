import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import InkwellLoadingButton from '../InkwellLoadingButton';

describe('InkwellLoadingButton', () => {
  it('押下で onAction が走り、完了後に成功文言を読み上げる', async () => {
    const action = vi.fn().mockResolvedValue(undefined);
    render(
      <InkwellLoadingButton onAction={action} successLabel="保存しました">
        保存
      </InkwellLoadingButton>,
    );
    fireEvent.click(screen.getByRole('button', { name: /保存/ }));
    expect(action).toHaveBeenCalledTimes(1);
    // loading 中は aria-disabled
    await waitFor(() => expect(screen.getByRole('button')).toHaveAttribute('aria-disabled', 'true'));
    // 解決後に success 文言
    await waitFor(() => expect(screen.getByText('保存しました')).toBeInTheDocument());
  });

  it('loading 中の再クリックは二重送信されない', async () => {
    let resolveFn: () => void = () => {};
    const action = vi.fn(() => new Promise<void>((r) => (resolveFn = r)));
    render(<InkwellLoadingButton onAction={action}>送信</InkwellLoadingButton>);
    const btn = screen.getByRole('button');
    fireEvent.click(btn);
    fireEvent.click(btn);
    fireEvent.click(btn);
    expect(action).toHaveBeenCalledTimes(1);
    resolveFn();
    // success 表示(約1.2s)を経て idle に戻るまで待つ。
    await waitFor(() => expect(btn).not.toHaveAttribute('aria-disabled'), { timeout: 3000 });
  });

  it('reject で失敗文言を読み上げる', async () => {
    const action = vi.fn().mockRejectedValue(new Error('ng'));
    render(
      <InkwellLoadingButton onAction={action} errorLabel="失敗しました">
        削除
      </InkwellLoadingButton>,
    );
    fireEvent.click(screen.getByRole('button'));
    await waitFor(() => expect(screen.getByText('失敗しました')).toBeInTheDocument());
  });
});
