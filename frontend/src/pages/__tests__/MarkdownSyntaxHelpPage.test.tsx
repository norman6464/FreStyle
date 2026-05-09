import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import MarkdownSyntaxHelpPage from '../MarkdownSyntaxHelpPage';

function renderPage() {
  return render(
    <MemoryRouter>
      <MarkdownSyntaxHelpPage />
    </MemoryRouter>,
  );
}

describe('MarkdownSyntaxHelpPage', () => {
  it('h1 にページタイトルを描画する', () => {
    renderPage();
    expect(
      screen.getByRole('heading', { name: 'Markdown 記法ヘルプ', level: 1 }),
    ).toBeInTheDocument();
  });

  it('主要なセクション見出しを表示する', () => {
    renderPage();
    expect(screen.getByRole('heading', { name: '見出し', level: 2 })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: '文字装飾', level: 2 })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'リスト', level: 2 })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: '表 (GFM)', level: 2 })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'コードブロック (言語指定)', level: 2 })).toBeInTheDocument();
  });

  it('ノートに戻るリンクを表示する', () => {
    renderPage();
    const link = screen.getByRole('link', { name: /ノートに戻る/ });
    expect(link).toHaveAttribute('href', '/notes');
  });

  it('コピーボタンが各 Row に配置される', () => {
    renderPage();
    const copyButtons = screen.getAllByRole('button', { name: 'コードをコピー' });
    expect(copyButtons.length).toBeGreaterThanOrEqual(8);
  });
});
