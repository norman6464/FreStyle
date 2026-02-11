import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import EmptyState from '../EmptyState';
import { ChatBubbleLeftRightIcon } from '@heroicons/react/24/outline';

describe('EmptyState', () => {
  it('タイトルを表示する', () => {
    render(<EmptyState icon={ChatBubbleLeftRightIcon} title="データがありません" />);
    expect(screen.getByText('データがありません')).toBeDefined();
  });

  it('説明文を表示する', () => {
    render(
      <EmptyState
        icon={ChatBubbleLeftRightIcon}
        title="テスト"
        description="詳しい説明テキスト"
      />
    );
    expect(screen.getByText('詳しい説明テキスト')).toBeDefined();
  });

  it('アクションボタンを表示しクリックできる', () => {
    const onClick = vi.fn();
    render(
      <EmptyState
        icon={ChatBubbleLeftRightIcon}
        title="テスト"
        action={{ label: 'ユーザーを追加', onClick }}
      />
    );
    const button = screen.getByText('ユーザーを追加');
    expect(button).toBeDefined();
    fireEvent.click(button);
    expect(onClick).toHaveBeenCalledOnce();
  });

  it('アクションなしの場合ボタンを表示しない', () => {
    render(<EmptyState icon={ChatBubbleLeftRightIcon} title="テスト" />);
    expect(screen.queryByRole('button')).toBeNull();
  });
});
