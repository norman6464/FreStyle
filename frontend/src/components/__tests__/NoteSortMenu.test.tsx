import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import NoteSortMenu from '../NoteSortMenu';

describe('NoteSortMenu', () => {
  it('現在の選択ラベルがトリガーに表示される', () => {
    render(<NoteSortMenu selected="title" onChange={() => {}} />);
    expect(screen.getByText('タイトル順')).toBeInTheDocument();
  });

  it('クリックでメニューが開き 4 件のオプションが表示される', () => {
    render(<NoteSortMenu selected="default" onChange={() => {}} />);
    fireEvent.click(screen.getByRole('button', { name: /並べ替え/ }));
    const menu = screen.getByRole('menu');
    const items = menu.querySelectorAll('[role="menuitemradio"]');
    expect(items).toHaveLength(4);
  });

  it('オプションを選ぶと onChange が呼ばれメニューが閉じる', () => {
    const onChange = vi.fn();
    render(<NoteSortMenu selected="default" onChange={onChange} />);
    fireEvent.click(screen.getByRole('button', { name: /並べ替え/ }));
    fireEvent.click(screen.getByRole('menuitemradio', { name: /タイトル順/ }));
    expect(onChange).toHaveBeenCalledWith('title');
    expect(screen.queryByRole('menu')).not.toBeInTheDocument();
  });

  it('外側クリックでメニューが閉じる', () => {
    render(
      <div>
        <NoteSortMenu selected="default" onChange={() => {}} />
        <button>外側</button>
      </div>,
    );
    fireEvent.click(screen.getByRole('button', { name: /並べ替え/ }));
    expect(screen.getByRole('menu')).toBeInTheDocument();
    fireEvent.mouseDown(screen.getByRole('button', { name: '外側' }));
    expect(screen.queryByRole('menu')).not.toBeInTheDocument();
  });
});
