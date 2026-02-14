import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import MessageSelectionPanel from '../MessageSelectionPanel';

describe('MessageSelectionPanel', () => {
  const defaultProps = {
    selectedCount: 0,
    onQuickSelect: vi.fn(),
    onSelectAll: vi.fn(),
    onDeselectAll: vi.fn(),
    onCancel: vi.fn(),
    onSend: vi.fn(),
  };

  it('選択件数0の場合にガイドメッセージが表示される', () => {
    render(<MessageSelectionPanel {...defaultProps} />);
    expect(screen.getByText('開始位置のメッセージをタップしてください')).toBeInTheDocument();
  });

  it('選択件数1以上の場合に件数メッセージが表示される', () => {
    render(<MessageSelectionPanel {...defaultProps} selectedCount={3} />);
    expect(screen.getByText('3件のメッセージを選択しました')).toBeInTheDocument();
  });

  it('クイック選択ボタンが表示される', () => {
    render(<MessageSelectionPanel {...defaultProps} />);
    expect(screen.getByText('直近5件')).toBeInTheDocument();
    expect(screen.getByText('直近10件')).toBeInTheDocument();
    expect(screen.getByText('直近20件')).toBeInTheDocument();
    expect(screen.getByText('すべて')).toBeInTheDocument();
  });

  it('クイック選択ボタンクリックでonQuickSelectが呼ばれる', () => {
    const onQuickSelect = vi.fn();
    render(<MessageSelectionPanel {...defaultProps} onQuickSelect={onQuickSelect} />);
    fireEvent.click(screen.getByText('直近10件'));
    expect(onQuickSelect).toHaveBeenCalledWith(10);
  });

  it('すべてボタンクリックでonSelectAllが呼ばれる', () => {
    const onSelectAll = vi.fn();
    render(<MessageSelectionPanel {...defaultProps} onSelectAll={onSelectAll} />);
    fireEvent.click(screen.getByText('すべて'));
    expect(onSelectAll).toHaveBeenCalled();
  });

  it('リセットボタンクリックでonDeselectAllが呼ばれる', () => {
    const onDeselectAll = vi.fn();
    render(<MessageSelectionPanel {...defaultProps} onDeselectAll={onDeselectAll} />);
    fireEvent.click(screen.getByText('リセット'));
    expect(onDeselectAll).toHaveBeenCalled();
  });

  it('キャンセルボタンクリックでonCancelが呼ばれる', () => {
    const onCancel = vi.fn();
    render(<MessageSelectionPanel {...defaultProps} onCancel={onCancel} />);
    fireEvent.click(screen.getByText('キャンセル'));
    expect(onCancel).toHaveBeenCalled();
  });

  it('選択件数0の場合に送信ボタンが無効', () => {
    render(<MessageSelectionPanel {...defaultProps} />);
    expect(screen.getByText('範囲を選択してください')).toBeDisabled();
  });

  it('選択件数1以上の場合に送信ボタンが有効', () => {
    const onSend = vi.fn();
    render(<MessageSelectionPanel {...defaultProps} selectedCount={5} onSend={onSend} />);
    const button = screen.getByText('5件をAIに送信');
    expect(button).not.toBeDisabled();
    fireEvent.click(button);
    expect(onSend).toHaveBeenCalled();
  });
});
