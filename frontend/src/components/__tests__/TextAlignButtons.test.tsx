import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import TextAlignButtons from '../TextAlignButtons';

describe('TextAlignButtons', () => {
  const onAlign = vi.fn();

  it('3つの配置ボタンが表示される', () => {
    render(<TextAlignButtons onAlign={onAlign} />);
    const buttons = screen.getAllByRole('button');
    expect(buttons).toHaveLength(3);
  });

  it('左寄せボタンにaria-labelがある', () => {
    render(<TextAlignButtons onAlign={onAlign} />);
    expect(screen.getByLabelText('左寄せ')).toBeInTheDocument();
  });

  it('中央寄せボタンにaria-labelがある', () => {
    render(<TextAlignButtons onAlign={onAlign} />);
    expect(screen.getByLabelText('中央寄せ')).toBeInTheDocument();
  });

  it('右寄せボタンにaria-labelがある', () => {
    render(<TextAlignButtons onAlign={onAlign} />);
    expect(screen.getByLabelText('右寄せ')).toBeInTheDocument();
  });

  it('左寄せクリックでonAlignが"left"で呼ばれる', () => {
    const handler = vi.fn();
    render(<TextAlignButtons onAlign={handler} />);
    fireEvent.click(screen.getByLabelText('左寄せ'));
    expect(handler).toHaveBeenCalledWith('left');
  });

  it('中央寄せクリックでonAlignが"center"で呼ばれる', () => {
    const handler = vi.fn();
    render(<TextAlignButtons onAlign={handler} />);
    fireEvent.click(screen.getByLabelText('中央寄せ'));
    expect(handler).toHaveBeenCalledWith('center');
  });

  it('右寄せクリックでonAlignが"right"で呼ばれる', () => {
    const handler = vi.fn();
    render(<TextAlignButtons onAlign={handler} />);
    fireEvent.click(screen.getByLabelText('右寄せ'));
    expect(handler).toHaveBeenCalledWith('right');
  });

  it('activeAlignに応じてボタンがハイライトされる', () => {
    render(<TextAlignButtons onAlign={onAlign} activeAlign="center" />);
    const centerButton = screen.getByLabelText('中央寄せ');
    expect(centerButton.className).toContain('text-primary');
  });
});
