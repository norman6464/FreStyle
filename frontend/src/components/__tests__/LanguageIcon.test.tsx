import { describe, it, expect } from 'vitest';
import { render, fireEvent } from '@testing-library/react';
import LanguageIcon from '../LanguageIcon';

// 言語ロゴは public/lang/<key>.svg（Devicon）を <img> で読む。jsdom は実際に画像を
// 取得しないので、フォールバック経路は onError を発火させて検証する。
describe('LanguageIcon (FRESTYLE-152)', () => {
  it('言語に対応する SVG を読み込む', () => {
    const { container } = render(<LanguageIcon language="go" />);
    const img = container.querySelector('img');
    expect(img).toHaveAttribute('src', '/lang/go.svg');
    // 装飾用途なので支援技術からは隠す。
    expect(img).toHaveAttribute('aria-hidden', 'true');
  });

  it('読み込みに失敗したら汎用アイコンにフォールバックする', () => {
    const { container } = render(<LanguageIcon language="unknown-lang" />);
    fireEvent.error(container.querySelector('img')!);
    // img は消え、Heroicons の svg に置き換わる。
    expect(container.querySelector('img')).toBeNull();
    expect(container.querySelector('svg')).not.toBeNull();
  });

  it('失敗後に language が変わったら再度その言語のアイコンを読みにいく', () => {
    // SPA で language prop だけが変わるケース（mount されたまま言語だけ切替）。
    // 失敗状態を boolean で持つと、以降ずっとフォールバックのままになってしまう。
    const { container, rerender } = render(<LanguageIcon language="unknown-lang" />);
    fireEvent.error(container.querySelector('img')!);
    expect(container.querySelector('img')).toBeNull();

    rerender(<LanguageIcon language="go" />);
    expect(container.querySelector('img')).toHaveAttribute('src', '/lang/go.svg');
  });

  it('className を差し替えられる', () => {
    const { container } = render(<LanguageIcon language="php" className="w-4 h-4" />);
    expect(container.querySelector('img')).toHaveClass('w-4', 'h-4');
  });

  it('フォールバックでも className が反映される', () => {
    const { container } = render(<LanguageIcon language="nope" className="w-4 h-4" />);
    fireEvent.error(container.querySelector('img')!);
    expect(container.querySelector('svg')).toHaveClass('w-4', 'h-4');
  });
});
