import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import CodeBlockNodeView from '../CodeBlockNodeView';

vi.mock('@tiptap/react', () => ({
  NodeViewWrapper: ({ children, className }: { children: React.ReactNode; className: string }) => (
    <div data-testid="node-view-wrapper" className={className}>{children}</div>
  ),
  NodeViewContent: ({ as }: { as: string }) => {
    const Tag = as as keyof JSX.IntrinsicElements;
    return <Tag data-testid="node-view-content">console.log("test")</Tag>;
  },
}));

describe('CodeBlockNodeView', () => {
  const mockUpdateAttributes = vi.fn();
  const defaultProps = {
    node: { attrs: { language: 'javascript' }, textContent: 'console.log("hello")' },
    updateAttributes: mockUpdateAttributes,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    Object.assign(navigator, {
      clipboard: { writeText: vi.fn().mockResolvedValue(undefined) },
    });
  });

  it('コードブロックが表示される', () => {
    render(<CodeBlockNodeView {...defaultProps} />);
    expect(screen.getByTestId('node-view-wrapper')).toHaveClass('code-block-wrapper');
    expect(screen.getByTestId('node-view-content')).toBeInTheDocument();
  });

  it('言語セレクターが選択された言語を表示する', () => {
    render(<CodeBlockNodeView {...defaultProps} />);
    const select = screen.getByLabelText('プログラミング言語') as HTMLSelectElement;
    expect(select.value).toBe('javascript');
  });

  it('言語を変更するとupdateAttributesが呼ばれる', () => {
    render(<CodeBlockNodeView {...defaultProps} />);
    fireEvent.change(screen.getByLabelText('プログラミング言語'), { target: { value: 'python' } });
    expect(mockUpdateAttributes).toHaveBeenCalledWith({ language: 'python' });
  });

  it('コピーボタンをクリックするとクリップボードにコピーされる', async () => {
    render(<CodeBlockNodeView {...defaultProps} />);
    fireEvent.click(screen.getByLabelText('コードをコピー'));
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith('console.log("hello")');
    await waitFor(() => {
      expect(screen.getByText('コピーしました')).toBeInTheDocument();
    });
  });

  it('言語がnullの場合はプレーンテキストが選択される', () => {
    const props = {
      ...defaultProps,
      node: { ...defaultProps.node, attrs: { language: null } },
    };
    render(<CodeBlockNodeView {...props} />);
    const select = screen.getByLabelText('プログラミング言語') as HTMLSelectElement;
    expect(select.value).toBe('');
  });

  it('コピーボタンにテキストが表示される', () => {
    render(<CodeBlockNodeView {...defaultProps} />);
    expect(screen.getByText('コピー')).toBeInTheDocument();
  });

  it('行番号が表示される（1行のコード）', () => {
    render(<CodeBlockNodeView {...defaultProps} />);
    const lineNumbers = screen.getByText('1');
    expect(lineNumbers).toBeInTheDocument();
  });

  it('複数行のコードで正しい行番号が表示される', () => {
    const props = {
      ...defaultProps,
      node: { ...defaultProps.node, textContent: 'line1\nline2\nline3' },
    };
    render(<CodeBlockNodeView {...props} />);
    expect(screen.getByText('1')).toBeInTheDocument();
    expect(screen.getByText('2')).toBeInTheDocument();
    expect(screen.getByText('3')).toBeInTheDocument();
  });

  it('空のコードブロックでも行番号1が表示される', () => {
    const props = {
      ...defaultProps,
      node: { ...defaultProps.node, textContent: '' },
    };
    render(<CodeBlockNodeView {...props} />);
    expect(screen.getByText('1')).toBeInTheDocument();
  });
});
