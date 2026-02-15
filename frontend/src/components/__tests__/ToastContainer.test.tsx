import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import ToastContainer from '../ToastContainer';

const mockToasts: { id: string; type: 'success' | 'error' | 'info'; message: string }[] = [];
const mockRemoveToast = vi.fn();

vi.mock('../../hooks/useToast', () => ({
  useToast: () => ({
    toasts: mockToasts,
    removeToast: mockRemoveToast,
  }),
}));

vi.mock('../Toast', () => ({
  default: ({ message, onClose }: { message: string; onClose: () => void }) => (
    <div data-testid="toast" onClick={onClose}>{message}</div>
  ),
}));

describe('ToastContainer', () => {
  it('トーストがない場合は何も表示しない', () => {
    mockToasts.length = 0;
    const { container } = render(<ToastContainer />);
    expect(container.firstChild).toBeNull();
  });

  it('トーストがある場合はメッセージが表示される', () => {
    mockToasts.length = 0;
    mockToasts.push({ id: '1', type: 'success', message: '保存しました' });
    render(<ToastContainer />);
    expect(screen.getByText('保存しました')).toBeInTheDocument();
  });

  it('複数のトーストが表示される', () => {
    mockToasts.length = 0;
    mockToasts.push(
      { id: '1', type: 'success', message: '成功' },
      { id: '2', type: 'error', message: 'エラー' },
    );
    render(<ToastContainer />);
    expect(screen.getByText('成功')).toBeInTheDocument();
    expect(screen.getByText('エラー')).toBeInTheDocument();
  });

  it('aria-live属性がある', () => {
    mockToasts.length = 0;
    mockToasts.push({ id: '1', type: 'info', message: 'テスト' });
    render(<ToastContainer />);
    expect(screen.getByText('テスト').closest('[aria-live]')).toHaveAttribute('aria-live', 'polite');
  });
});
