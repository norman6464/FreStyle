import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import HelpPage from '../HelpPage';
import { createMockStorage } from '../../test/mockStorage';

function renderHelp() {
  return render(
    <MemoryRouter>
      <HelpPage />
    </MemoryRouter>
  );
}

describe('HelpPage', () => {
  beforeEach(() => {
    vi.stubGlobal('localStorage', createMockStorage());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('PageIntro のタイトルが h1 として描画される', () => {
    renderHelp();
    const heading = screen.getByRole('heading', { name: '使い方ガイド', level: 1 });
    expect(heading).toBeInTheDocument();
  });

  it('全 7 章の見出しが表示される', () => {
    renderHelp();
    expect(screen.getByRole('heading', { name: /1\. FreStyle ってなに？/ })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /2\. 最初の1日にやること/ })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /3\..*の使い方/ })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /4\..*の読み方/ })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /5\. AI チャットとは/ })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /6\. メモ・テンプレート機能/ })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /7\. 困ったとき/ })).toBeInTheDocument();
  });

  it('Step 1 の ActionCard が /practice へのリンクとして描画される', () => {
    renderHelp();
    const link = screen.getByRole('link', { name: /練習モードを開く/ });
    expect(link).toHaveAttribute('href', '/practice');
  });

  it('Step 3 の ActionCard が /scores へのリンクとして描画される', () => {
    renderHelp();
    const link = screen.getByRole('link', { name: /スコア履歴を確認する/ });
    expect(link).toHaveAttribute('href', '/scores');
  });

  it('AI アシスタントへの導線が /chat/ask-ai を指す', () => {
    renderHelp();
    const link = screen.getByRole('link', { name: /AI アシスタントに相談する/ });
    expect(link).toHaveAttribute('href', '/chat/ask-ai');
  });

  it('メモ・テンプレートのカードがそれぞれ /notes と /templates を指す', () => {
    renderHelp();
    expect(screen.getByRole('link', { name: /メモを書く/ })).toHaveAttribute('href', '/notes');
    expect(screen.getByRole('link', { name: /テンプレートを使う/ })).toHaveAttribute('href', '/templates');
  });

  it('末尾の CTA がホームに戻るリンクになっている', () => {
    renderHelp();
    expect(screen.getByRole('link', { name: /ホームに戻って練習を始める/ })).toHaveAttribute('href', '/');
  });

  it('5軸評価の 5 つの評価項目が dt として全て描画される', () => {
    renderHelp();
    expect(screen.getByText('論理的構成力')).toBeInTheDocument();
    expect(screen.getByText('配慮表現')).toBeInTheDocument();
    expect(screen.getByText('要約力')).toBeInTheDocument();
    expect(screen.getByText('提案力')).toBeInTheDocument();
    expect(screen.getByText('質問・傾聴力')).toBeInTheDocument();
  });

  it('FAQ セクションの details 要素が 4 件ある', () => {
    const { container } = renderHelp();
    const detailsElements = container.querySelectorAll('details');
    expect(detailsElements).toHaveLength(4);
  });

  it('GuidedHint が初回表示で見え、閉じると localStorage に保存される', () => {
    renderHelp();
    expect(screen.getByText('このページの読み方')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'ヒントを閉じる' }));
    expect(window.localStorage.getItem('hint:help:howto-v1')).toBe('dismissed');
  });

  it('再訪問時（dismissed が記録済み）は GuidedHint を表示しない', () => {
    window.localStorage.setItem('hint:help:howto-v1', 'dismissed');
    renderHelp();
    expect(screen.queryByText('このページの読み方')).not.toBeInTheDocument();
  });
});
