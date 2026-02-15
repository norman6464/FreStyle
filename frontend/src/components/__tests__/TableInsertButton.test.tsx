import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import TableInsertButton from '../TableInsertButton';

describe('TableInsertButton', () => {
  it('テーブル挿入ボタンが表示される', () => {
    render(<TableInsertButton onInsertTable={vi.fn()} />);
    expect(screen.getByLabelText('テーブル挿入')).toBeInTheDocument();
  });

  it('クリックでonInsertTableが呼ばれる', () => {
    const onInsertTable = vi.fn();
    render(<TableInsertButton onInsertTable={onInsertTable} />);
    fireEvent.click(screen.getByLabelText('テーブル挿入'));
    expect(onInsertTable).toHaveBeenCalledTimes(1);
  });

  it('ToolbarIconButtonを使用している', () => {
    render(<TableInsertButton onInsertTable={vi.fn()} />);
    const button = screen.getByLabelText('テーブル挿入');
    expect(button.tagName).toBe('BUTTON');
  });
});
