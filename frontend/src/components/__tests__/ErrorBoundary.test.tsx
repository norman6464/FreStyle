import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import ErrorBoundary from '../ErrorBoundary';

function ThrowError({ shouldThrow }: { shouldThrow: boolean }) {
  if (shouldThrow) {
    throw new Error('テストエラー');
  }
  return <div>正常なコンテンツ</div>;
}

describe('ErrorBoundary', () => {
  beforeEach(() => {
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  it('エラーがない場合は子コンポーネントを表示する', () => {
    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={false} />
      </ErrorBoundary>
    );

    expect(screen.getByText('正常なコンテンツ')).toBeInTheDocument();
  });

  it('エラー発生時にフレンドリーなメッセージを表示する', () => {
    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    expect(screen.getByText('エラーが発生しました')).toBeInTheDocument();
  });

  it('エラー発生時に再試行ボタンが表示される', () => {
    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    expect(screen.getByText('再試行')).toBeInTheDocument();
  });

  it('エラー発生時に説明テキストが表示される', () => {
    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    expect(screen.getByText(/予期せぬエラーが発生しました/)).toBeInTheDocument();
  });
});
