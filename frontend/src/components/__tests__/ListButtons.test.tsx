import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import ListButtons from '../ListButtons';

describe('ListButtons', () => {
  it('箇条書きボタンが表示される', () => {
    render(<ListButtons onBulletList={vi.fn()} onOrderedList={vi.fn()} />);
    expect(screen.getByLabelText('箇条書き')).toBeInTheDocument();
  });

  it('番号付きリストボタンが表示される', () => {
    render(<ListButtons onBulletList={vi.fn()} onOrderedList={vi.fn()} />);
    expect(screen.getByLabelText('番号付きリスト')).toBeInTheDocument();
  });

  it('箇条書きボタンクリックでonBulletListが呼ばれる', () => {
    const onBulletList = vi.fn();
    render(<ListButtons onBulletList={onBulletList} onOrderedList={vi.fn()} />);
    fireEvent.click(screen.getByLabelText('箇条書き'));
    expect(onBulletList).toHaveBeenCalledTimes(1);
  });

  it('番号付きリストボタンクリックでonOrderedListが呼ばれる', () => {
    const onOrderedList = vi.fn();
    render(<ListButtons onBulletList={vi.fn()} onOrderedList={onOrderedList} />);
    fireEvent.click(screen.getByLabelText('番号付きリスト'));
    expect(onOrderedList).toHaveBeenCalledTimes(1);
  });

  it('ボタンがbutton要素である', () => {
    render(<ListButtons onBulletList={vi.fn()} onOrderedList={vi.fn()} />);
    const bulletBtn = screen.getByLabelText('箇条書き');
    const orderedBtn = screen.getByLabelText('番号付きリスト');
    expect(bulletBtn.tagName).toBe('BUTTON');
    expect(bulletBtn).toHaveAttribute('type', 'button');
    expect(orderedBtn.tagName).toBe('BUTTON');
    expect(orderedBtn).toHaveAttribute('type', 'button');
  });
});
