import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import CalloutNodeView from '../CalloutNodeView';

vi.mock('@tiptap/react', () => ({
  NodeViewWrapper: ({ children, className }: { children: React.ReactNode; className: string }) => (
    <div data-testid="node-view-wrapper" className={className}>{children}</div>
  ),
  NodeViewContent: () => <div data-testid="node-view-content" />,
}));

const defaultProps = {
  node: { attrs: { type: 'info' as const, emoji: '💡' } },
  updateAttributes: vi.fn(),
};

describe('CalloutNodeView', () => {
  it('絵文字が表示される', () => {
    render(<CalloutNodeView {...defaultProps} />);
    expect(screen.getByText('💡')).toBeInTheDocument();
  });

  it('コンテンツ領域が表示される', () => {
    render(<CalloutNodeView {...defaultProps} />);
    expect(screen.getByTestId('node-view-content')).toBeInTheDocument();
  });

  it('タイプ変更ボタンのaria-labelがある', () => {
    render(<CalloutNodeView {...defaultProps} />);
    expect(screen.getByLabelText('コールアウトタイプを変更')).toBeInTheDocument();
  });

  it('絵文字クリックでメニューが表示される', () => {
    render(<CalloutNodeView {...defaultProps} />);
    fireEvent.click(screen.getByLabelText('コールアウトタイプを変更'));
    expect(screen.getByText('情報')).toBeInTheDocument();
    expect(screen.getByText('警告')).toBeInTheDocument();
    expect(screen.getByText('エラー')).toBeInTheDocument();
    expect(screen.getByText('成功')).toBeInTheDocument();
  });

  it('メニューからタイプを選択するとupdateAttributesが呼ばれる', () => {
    const updateAttributes = vi.fn();
    render(<CalloutNodeView {...defaultProps} updateAttributes={updateAttributes} />);
    fireEvent.click(screen.getByLabelText('コールアウトタイプを変更'));
    fireEvent.click(screen.getByText('警告'));
    expect(updateAttributes).toHaveBeenCalledWith({ type: 'warning', emoji: '⚠️' });
  });

  it('メニュー選択後にメニューが閉じる', () => {
    render(<CalloutNodeView {...defaultProps} />);
    fireEvent.click(screen.getByLabelText('コールアウトタイプを変更'));
    expect(screen.getByText('情報')).toBeInTheDocument();
    fireEvent.click(screen.getByText('成功'));
    expect(screen.queryByText('情報')).not.toBeInTheDocument();
  });

  it('warningタイプで正しいスタイルが適用される', () => {
    const props = {
      ...defaultProps,
      node: { attrs: { type: 'warning' as const, emoji: '⚠️' } },
    };
    render(<CalloutNodeView {...props} />);
    const wrapper = screen.getByTestId('node-view-wrapper');
    expect(wrapper.className).toContain('border-amber');
  });

  it('errorタイプで正しいスタイルが適用される', () => {
    const props = {
      ...defaultProps,
      node: { attrs: { type: 'error' as const, emoji: '🚫' } },
    };
    render(<CalloutNodeView {...props} />);
    const wrapper = screen.getByTestId('node-view-wrapper');
    expect(wrapper.className).toContain('border-red');
  });

  it('successタイプで正しいスタイルが適用される', () => {
    const props = {
      ...defaultProps,
      node: { attrs: { type: 'success' as const, emoji: '✅' } },
    };
    render(<CalloutNodeView {...props} />);
    const wrapper = screen.getByTestId('node-view-wrapper');
    expect(wrapper.className).toContain('border-green');
  });
});
