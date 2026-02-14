import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import SkipLink from '../SkipLink';

describe('SkipLink', () => {
  it('リンクがレンダリングされる', () => {
    render(<SkipLink targetId="main-content" />);
    expect(screen.getByText('メインコンテンツへスキップ')).toBeInTheDocument();
  });

  it('デフォルトでは視覚的に非表示', () => {
    render(<SkipLink targetId="main-content" />);
    const link = screen.getByText('メインコンテンツへスキップ');
    expect(link.className).toContain('sr-only');
  });

  it('フォーカス時に表示される', () => {
    render(<SkipLink targetId="main-content" />);
    const link = screen.getByText('メインコンテンツへスキップ');
    fireEvent.focus(link);
    expect(link.className).toContain('focus:not-sr-only');
  });

  it('href属性がターゲットIDを参照する', () => {
    render(<SkipLink targetId="main-content" />);
    const link = screen.getByText('メインコンテンツへスキップ');
    expect(link).toHaveAttribute('href', '#main-content');
  });

  it('カスタムラベルが表示される', () => {
    render(<SkipLink targetId="main-content" label="ナビゲーションをスキップ" />);
    expect(screen.getByText('ナビゲーションをスキップ')).toBeInTheDocument();
  });

  it('クリック時にターゲット要素にフォーカスが移動する', () => {
    const target = document.createElement('div');
    target.id = 'main-content';
    target.tabIndex = -1;
    document.body.appendChild(target);

    render(<SkipLink targetId="main-content" />);
    const link = screen.getByText('メインコンテンツへスキップ');
    fireEvent.click(link);
    expect(document.activeElement).toBe(target);

    document.body.removeChild(target);
  });

  it('z-50クラスでナビゲーション上に表示される', () => {
    render(<SkipLink targetId="main-content" />);
    const link = screen.getByText('メインコンテンツへスキップ');
    expect(link.className).toContain('z-50');
  });

  it('role=linkが暗黙的に設定される（aタグ）', () => {
    render(<SkipLink targetId="main-content" />);
    const link = screen.getByText('メインコンテンツへスキップ');
    expect(link.tagName).toBe('A');
  });

  it('ターゲット要素が存在しない場合クリックしてもエラーにならない', () => {
    render(<SkipLink targetId="nonexistent" />);
    const link = screen.getByText('メインコンテンツへスキップ');
    expect(() => fireEvent.click(link)).not.toThrow();
  });

  it('フォーカス時にprimary-500背景クラスが適用される', () => {
    render(<SkipLink targetId="main-content" />);
    const link = screen.getByText('メインコンテンツへスキップ');
    expect(link.className).toContain('focus:bg-primary-500');
  });

  it('異なるtargetIdでhrefが正しく設定される', () => {
    render(<SkipLink targetId="custom-target" />);
    const link = screen.getByText('メインコンテンツへスキップ');
    expect(link).toHaveAttribute('href', '#custom-target');
  });
});
