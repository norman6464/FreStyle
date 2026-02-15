import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import IndentButtons from '../IndentButtons';

describe('IndentButtons', () => {
  it('2つのボタンが表示される', () => {
    render(<IndentButtons onIndent={vi.fn()} onOutdent={vi.fn()} />);
    const buttons = screen.getAllByRole('button');
    expect(buttons).toHaveLength(2);
  });

  it('インデント増加ボタンにaria-labelがある', () => {
    render(<IndentButtons onIndent={vi.fn()} onOutdent={vi.fn()} />);
    expect(screen.getByLabelText('インデント増加')).toBeInTheDocument();
  });

  it('インデント減少ボタンにaria-labelがある', () => {
    render(<IndentButtons onIndent={vi.fn()} onOutdent={vi.fn()} />);
    expect(screen.getByLabelText('インデント減少')).toBeInTheDocument();
  });

  it('インデント増加クリックでonIndentが呼ばれる', () => {
    const onIndent = vi.fn();
    render(<IndentButtons onIndent={onIndent} onOutdent={vi.fn()} />);
    fireEvent.click(screen.getByLabelText('インデント増加'));
    expect(onIndent).toHaveBeenCalledOnce();
  });

  it('インデント減少クリックでonOutdentが呼ばれる', () => {
    const onOutdent = vi.fn();
    render(<IndentButtons onIndent={vi.fn()} onOutdent={onOutdent} />);
    fireEvent.click(screen.getByLabelText('インデント減少'));
    expect(onOutdent).toHaveBeenCalledOnce();
  });
});
