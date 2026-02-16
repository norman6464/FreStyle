import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import AiSessionListItem from '../AiSessionListItem';

const baseProps = {
  id: 1,
  title: 'テストセッション',
  createdAt: '2025-01-15T10:00:00',
  isActive: false,
  isEditing: false,
  editingTitle: '',
  onSelect: vi.fn(),
  onStartEdit: vi.fn(),
  onDelete: vi.fn(),
  onSaveTitle: vi.fn(),
  onCancelEdit: vi.fn(),
  onEditingTitleChange: vi.fn(),
};

describe('AiSessionListItem', () => {
  it('セッションタイトルを表示する', () => {
    render(<AiSessionListItem {...baseProps} />);
    expect(screen.getByText('テストセッション')).toBeInTheDocument();
  });

  it('タイトルがない場合「新しいチャット」を表示する', () => {
    render(<AiSessionListItem {...baseProps} title="" />);
    expect(screen.getByText('新しいチャット')).toBeInTheDocument();
  });

  it('作成日を日本語フォーマットで表示する', () => {
    render(<AiSessionListItem {...baseProps} />);
    expect(screen.getByText('2025/1/15')).toBeInTheDocument();
  });

  it('クリックでonSelectが呼ばれる', () => {
    const onSelect = vi.fn();
    render(<AiSessionListItem {...baseProps} onSelect={onSelect} />);
    fireEvent.click(screen.getByText('テストセッション'));
    expect(onSelect).toHaveBeenCalledWith(1);
  });

  it('アクティブ時にアクティブスタイルが適用される', () => {
    const { container } = render(<AiSessionListItem {...baseProps} isActive={true} />);
    const item = container.firstElementChild as HTMLElement;
    expect(item.className).toContain('text-primary-300');
  });

  it('編集モードで入力欄が表示される', () => {
    render(
      <AiSessionListItem
        {...baseProps}
        isEditing={true}
        editingTitle="編集中タイトル"
      />
    );
    expect(screen.getByDisplayValue('編集中タイトル')).toBeInTheDocument();
  });

  it('編集中にEnterで保存される', () => {
    const onSaveTitle = vi.fn();
    render(
      <AiSessionListItem
        {...baseProps}
        isEditing={true}
        editingTitle="新タイトル"
        onSaveTitle={onSaveTitle}
      />
    );
    fireEvent.keyDown(screen.getByDisplayValue('新タイトル'), { key: 'Enter' });
    expect(onSaveTitle).toHaveBeenCalledWith(1);
  });

  it('編集中にEscapeでキャンセルされる', () => {
    const onCancelEdit = vi.fn();
    render(
      <AiSessionListItem
        {...baseProps}
        isEditing={true}
        editingTitle="新タイトル"
        onCancelEdit={onCancelEdit}
      />
    );
    fireEvent.keyDown(screen.getByDisplayValue('新タイトル'), { key: 'Escape' });
    expect(onCancelEdit).toHaveBeenCalled();
  });

  it('削除ボタンでonDeleteが呼ばれる', () => {
    const onDelete = vi.fn();
    render(<AiSessionListItem {...baseProps} onDelete={onDelete} />);
    const deleteButton = screen.getByTitle('削除');
    fireEvent.click(deleteButton);
    expect(onDelete).toHaveBeenCalledWith(1);
  });

  it('role="button"とtabIndex={0}を持つ', () => {
    render(<AiSessionListItem {...baseProps} />);
    const item = screen.getByRole('button', { name: /テストセッション/ });
    expect(item).toHaveAttribute('tabindex', '0');
  });

  it('Enterキーでセッション選択できる', () => {
    const onSelect = vi.fn();
    render(<AiSessionListItem {...baseProps} onSelect={onSelect} />);
    const item = screen.getByRole('button', { name: /テストセッション/ });
    fireEvent.keyDown(item, { key: 'Enter' });
    expect(onSelect).toHaveBeenCalledWith(1);
  });

  it('Spaceキーでセッション選択できる', () => {
    const onSelect = vi.fn();
    render(<AiSessionListItem {...baseProps} onSelect={onSelect} />);
    const item = screen.getByRole('button', { name: /テストセッション/ });
    fireEvent.keyDown(item, { key: ' ' });
    expect(onSelect).toHaveBeenCalledWith(1);
  });

  it('編集中はキーボードでセッション選択されない', () => {
    const onSelect = vi.fn();
    render(<AiSessionListItem {...baseProps} isEditing={true} editingTitle="編集中" onSelect={onSelect} />);
    const item = screen.getByRole('button', { name: /テストセッション/ });
    fireEvent.keyDown(item, { key: 'Enter' });
    expect(onSelect).not.toHaveBeenCalled();
  });
});
