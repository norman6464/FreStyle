import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import HeadingSelect from '../HeadingSelect';

describe('HeadingSelect', () => {
  it('見出しセレクトが表示される', () => {
    render(<HeadingSelect onHeading={vi.fn()} />);
    expect(screen.getByLabelText('見出し')).toBeInTheDocument();
  });

  it('セレクトにオプションが含まれる', () => {
    render(<HeadingSelect onHeading={vi.fn()} />);
    const select = screen.getByLabelText('見出し') as HTMLSelectElement;
    expect(select.options.length).toBeGreaterThanOrEqual(4);
  });

  it('見出しレベル選択でonHeadingが呼ばれる', () => {
    const onHeading = vi.fn();
    render(<HeadingSelect onHeading={onHeading} />);
    fireEvent.change(screen.getByLabelText('見出し'), { target: { value: '2' } });
    expect(onHeading).toHaveBeenCalledWith(2);
  });

  it('標準テキスト選択でonHeadingが0で呼ばれる', () => {
    const onHeading = vi.fn();
    render(<HeadingSelect onHeading={onHeading} />);
    fireEvent.change(screen.getByLabelText('見出し'), { target: { value: '0' } });
    expect(onHeading).toHaveBeenCalledWith(0);
  });
});
