import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import Loading from '../Loading';

describe('Loading', () => {
  it('スピナーが表示される', () => {
    render(<Loading />);

    expect(screen.getByRole('status')).toBeInTheDocument();
  });

  it('メッセージが表示される', () => {
    render(<Loading message="読み込み中..." />);

    expect(screen.getByText('読み込み中...')).toBeInTheDocument();
  });

  it('フルスクリーンモードで表示される', () => {
    render(<Loading fullscreen message="処理中..." />);

    expect(screen.getByText('処理中...')).toBeInTheDocument();
  });

  it('メッセージなしの場合テキストが表示されない', () => {
    const { container } = render(<Loading />);

    expect(container.querySelector('p')).toBeNull();
  });
});
