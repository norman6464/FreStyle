import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import DocumentListItem from '../DocumentListItem';

const props = {
  id: 'a',
  title: 'メモA',
  updatedAt: '2026-08-02T00:00:00Z',
  isActive: false,
  onSelect: vi.fn(),
  onDelete: vi.fn(),
};

// li は ul の子であることが前提なので ul で包んで描画する。
const renderItem = (p = props) =>
  render(
    <ul>
      <DocumentListItem {...p} />
    </ul>,
  );

describe('DocumentListItem', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('タイトルを表示する（空なら「無題」）', () => {
    const { rerender } = renderItem();
    expect(screen.getByText('メモA')).toBeInTheDocument();
    rerender(
      <ul>
        <DocumentListItem {...props} title="" />
      </ul>,
    );
    expect(screen.getByText('無題')).toBeInTheDocument();
  });

  it('選択用と削除用の 2 つの独立したボタンを公開する', () => {
    renderItem();
    expect(screen.getByRole('button', { name: 'ノート「メモA」を選択' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ノート「メモA」を削除' })).toBeInTheDocument();
  });

  it('更新日を年込みで <time> に表示する', () => {
    renderItem();
    const time = screen.getByText(/2026/);
    expect(time.tagName.toLowerCase()).toBe('time');
    expect(time).toHaveAttribute('datetime', '2026-08-02T00:00:00Z');
  });

  it('選択ボタンで onSelect(id) を呼ぶ', () => {
    const onSelect = vi.fn();
    renderItem({ ...props, onSelect });
    fireEvent.click(screen.getByRole('button', { name: 'ノート「メモA」を選択' }));
    expect(onSelect).toHaveBeenCalledWith('a');
  });

  it('削除ボタンで onDelete(id) を呼び、選択はしない', () => {
    const onSelect = vi.fn();
    const onDelete = vi.fn();
    renderItem({ ...props, onSelect, onDelete });
    fireEvent.click(screen.getByRole('button', { name: 'ノート「メモA」を削除' }));
    expect(onDelete).toHaveBeenCalledWith('a');
    expect(onSelect).not.toHaveBeenCalled();
  });

  it('isActive のとき選択ボタンに aria-current=true', () => {
    renderItem({ ...props, isActive: true });
    expect(screen.getByRole('button', { name: 'ノート「メモA」を選択' })).toHaveAttribute('aria-current', 'true');
  });
});
