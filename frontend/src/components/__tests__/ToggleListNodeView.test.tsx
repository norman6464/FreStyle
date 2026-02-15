import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import ToggleListNodeView from '../ToggleListNodeView';

vi.mock('@tiptap/react', () => ({
  NodeViewWrapper: ({ children, className, ...rest }: { children: React.ReactNode; className: string; [key: string]: unknown }) => (
    <div data-testid="node-view-wrapper" className={className} {...rest}>{children}</div>
  ),
  NodeViewContent: () => <div data-testid="node-view-content" />,
}));

const defaultProps = {
  node: { attrs: { open: false } },
  updateAttributes: vi.fn(),
};

describe('ToggleListNodeView', () => {
  it('コンテンツ領域が表示される', () => {
    render(<ToggleListNodeView {...defaultProps} />);
    expect(screen.getByTestId('node-view-content')).toBeInTheDocument();
  });

  it('閉じた状態でaria-label「トグルを開く」が表示される', () => {
    render(<ToggleListNodeView {...defaultProps} />);
    expect(screen.getByLabelText('トグルを開く')).toBeInTheDocument();
  });

  it('開いた状態でaria-label「トグルを閉じる」が表示される', () => {
    const props = { ...defaultProps, node: { attrs: { open: true } } };
    render(<ToggleListNodeView {...props} />);
    expect(screen.getByLabelText('トグルを閉じる')).toBeInTheDocument();
  });

  it('閉じた状態でaria-expanded=falseが設定される', () => {
    render(<ToggleListNodeView {...defaultProps} />);
    expect(screen.getByLabelText('トグルを開く')).toHaveAttribute('aria-expanded', 'false');
  });

  it('開いた状態でaria-expanded=trueが設定される', () => {
    const props = { ...defaultProps, node: { attrs: { open: true } } };
    render(<ToggleListNodeView {...props} />);
    expect(screen.getByLabelText('トグルを閉じる')).toHaveAttribute('aria-expanded', 'true');
  });

  it('ボタンクリックでupdateAttributesが呼ばれる（閉→開）', () => {
    const updateAttributes = vi.fn();
    render(<ToggleListNodeView {...defaultProps} updateAttributes={updateAttributes} />);
    fireEvent.click(screen.getByLabelText('トグルを開く'));
    expect(updateAttributes).toHaveBeenCalledWith({ open: true });
  });

  it('ボタンクリックでupdateAttributesが呼ばれる（開→閉）', () => {
    const updateAttributes = vi.fn();
    const props = { node: { attrs: { open: true } }, updateAttributes };
    render(<ToggleListNodeView {...props} />);
    fireEvent.click(screen.getByLabelText('トグルを閉じる'));
    expect(updateAttributes).toHaveBeenCalledWith({ open: false });
  });

  it('閉じた状態でtoggle-closedクラスが適用される', () => {
    render(<ToggleListNodeView {...defaultProps} />);
    const wrapper = screen.getByTestId('node-view-wrapper');
    expect(wrapper.className).toContain('toggle-closed');
  });

  it('開いた状態でtoggle-openクラスが適用される', () => {
    const props = { ...defaultProps, node: { attrs: { open: true } } };
    render(<ToggleListNodeView {...props} />);
    const wrapper = screen.getByTestId('node-view-wrapper');
    expect(wrapper.className).toContain('toggle-open');
  });
});
