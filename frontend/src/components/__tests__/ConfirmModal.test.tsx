import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import ConfirmModal from '../ConfirmModal';

describe('ConfirmModal', () => {
  const mockOnConfirm = vi.fn();
  const mockOnCancel = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('isOpen=trueでモーダルが表示される', () => {
    render(
      <ConfirmModal isOpen={true} message="削除しますか？" onConfirm={mockOnConfirm} onCancel={mockOnCancel} />
    );

    expect(screen.getByText('確認')).toBeInTheDocument();
    expect(screen.getByText('削除しますか？')).toBeInTheDocument();
  });

  it('isOpen=falseでモーダルが非表示になる', () => {
    render(
      <ConfirmModal isOpen={false} message="削除しますか？" onConfirm={mockOnConfirm} onCancel={mockOnCancel} />
    );

    expect(screen.queryByText('削除しますか？')).not.toBeInTheDocument();
  });

  it('確認ボタンクリックでonConfirmが呼ばれる', () => {
    render(
      <ConfirmModal isOpen={true} message="削除しますか？" onConfirm={mockOnConfirm} onCancel={mockOnCancel} />
    );

    fireEvent.click(screen.getByText('削除'));
    expect(mockOnConfirm).toHaveBeenCalled();
  });

  it('キャンセルボタンクリックでonCancelが呼ばれる', () => {
    render(
      <ConfirmModal isOpen={true} message="削除しますか？" onConfirm={mockOnConfirm} onCancel={mockOnCancel} />
    );

    fireEvent.click(screen.getByText('キャンセル'));
    expect(mockOnCancel).toHaveBeenCalled();
  });

  it('カスタムタイトルとボタンテキストが表示される', () => {
    render(
      <ConfirmModal
        isOpen={true}
        title="注意"
        message="本当に実行しますか？"
        confirmText="実行"
        cancelText="戻る"
        onConfirm={mockOnConfirm}
        onCancel={mockOnCancel}
      />
    );

    expect(screen.getByText('注意')).toBeInTheDocument();
    expect(screen.getByText('実行')).toBeInTheDocument();
    expect(screen.getByText('戻る')).toBeInTheDocument();
  });

  it('ESCキーでonCancelが呼ばれる', () => {
    render(
      <ConfirmModal isOpen={true} message="削除しますか？" onConfirm={mockOnConfirm} onCancel={mockOnCancel} />
    );

    fireEvent.keyDown(window, { key: 'Escape' });
    expect(mockOnCancel).toHaveBeenCalled();
  });

  it('isDanger=falseで確認ボタンがprimaryスタイルになる', () => {
    render(
      <ConfirmModal isOpen={true} message="実行しますか？" isDanger={false} onConfirm={mockOnConfirm} onCancel={mockOnCancel} />
    );
    const confirmBtn = screen.getByText('削除');
    expect(confirmBtn.className).toContain('bg-primary-500');
  });

  it('isDanger=trueで確認ボタンがredスタイルになる', () => {
    render(
      <ConfirmModal isOpen={true} message="削除しますか？" isDanger={true} onConfirm={mockOnConfirm} onCancel={mockOnCancel} />
    );
    const confirmBtn = screen.getByText('削除');
    expect(confirmBtn.className).toContain('bg-red-500');
  });

  it('オーバーレイクリックでonCancelが呼ばれる', () => {
    const { container } = render(
      <ConfirmModal isOpen={true} message="削除しますか？" onConfirm={mockOnConfirm} onCancel={mockOnCancel} />
    );
    const overlay = container.querySelector('.bg-black\\/50');
    if (overlay) fireEvent.click(overlay);
    expect(mockOnCancel).toHaveBeenCalled();
  });
});
