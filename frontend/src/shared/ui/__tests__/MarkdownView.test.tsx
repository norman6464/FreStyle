import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import MarkdownView from '../MarkdownView';

// jsdom は CSS アニメを実行しないため「見た目のフェード」は検証できないが、
// このフェード方式の唯一の肝は「表示が進んでも既表示チャンクの span を再マウントしない
// （＝再フェードさせない）」こと。これは DOM ノードの同一性で直接検証できる。
describe('MarkdownView', () => {
  it('非ストリーミング時は素の Markdown（fade-seg span を作らない）', () => {
    const { container } = render(<MarkdownView content="今日は良い天気です" />);
    // 分割されないので getByText の完全一致が通る（既存テストの前提）。
    expect(screen.getByText('今日は良い天気です')).toBeInTheDocument();
    expect(container.querySelectorAll('.fade-seg')).toHaveLength(0);
  });

  it('ストリーミング時は本文を句読点チャンクの fade-seg span に分割する（テキストは保持）', () => {
    const { container } = render(<MarkdownView content="今日は晴れ。明日は雨。" isStreaming />);
    const spans = container.querySelectorAll('.fade-seg');
    expect(spans.length).toBe(2);
    // 分割してもレンダリング結果の全文は変わらない。
    expect(container.textContent).toContain('今日は晴れ。明日は雨。');
  });

  it('ストリーミング時でもコードブロックは分割しない（ハイライト/コピーを壊さない）', () => {
    const md = '本文です。\n\n```js\nconst a = 1;\n```';
    const { container } = render(<MarkdownView content={md} isStreaming />);
    // 本文側には fade-seg span がある。
    expect(container.querySelectorAll('.fade-seg').length).toBeGreaterThan(0);
    // pre(コードブロック)配下には fade-seg span を作らない。
    expect(container.querySelectorAll('pre .fade-seg')).toHaveLength(0);
    // コードの中身は保持される。
    expect(container.querySelector('pre')?.textContent).toContain('const a = 1;');
  });

  it('表示が進んでも既表示チャンクの span は再マウントされず DOM ノードが再利用される（再フェードしない）', () => {
    const { container, rerender } = render(<MarkdownView content="今日は晴れ。" isStreaming />);
    const firstBefore = container.querySelector('.fade-seg');
    expect(firstBefore).not.toBeNull();
    expect(firstBefore?.textContent).toBe('今日は晴れ。');

    // 次のチャンクが放出されて本文が伸びる。
    rerender(<MarkdownView content="今日は晴れ。明日は雨。" isStreaming />);
    const firstAfter = container.querySelector('.fade-seg');

    // 先頭チャンクの span は同一 DOM ノードのまま（React が再利用＝CSS アニメが再生し直されない）。
    expect(firstAfter).toBe(firstBefore);
    expect(firstAfter?.textContent).toBe('今日は晴れ。');
    // 末尾には新しいチャンク span が増えている。
    expect(container.querySelectorAll('.fade-seg')).toHaveLength(2);
  });

  it('見出しなどのブロック要素配下のテキストも fade-seg span になる', () => {
    const { container } = render(<MarkdownView content="# タイトル見出し" isStreaming />);
    const heading = container.querySelector('h1');
    expect(heading).not.toBeNull();
    expect(heading?.querySelectorAll('.fade-seg').length).toBeGreaterThan(0);
    expect(heading?.textContent).toBe('タイトル見出し');
  });

  it('ストリーミング時に GFM(リンク/表/太字)とチャンク分割が共存しても構造が壊れない', () => {
    const md = [
      'これは **強調** と [リンク](https://example.com) です',
      '',
      '| 名前 | 値 |',
      '| --- | --- |',
      '| foo | bar |',
    ].join('\n');
    const { container } = render(<MarkdownView content={md} isStreaming />);

    // リンクは href を保持したまま、リンクテキストも fade-seg 分割される。
    const link = container.querySelector('a[href="https://example.com"]');
    expect(link).not.toBeNull();
    expect(link?.textContent).toContain('リンク');

    // 太字(strong)は要素として残り、配下も分割される。
    const strong = container.querySelector('strong');
    expect(strong).not.toBeNull();
    expect(strong?.textContent).toBe('強調');

    // 表の構造(セル)が保たれ、セル内のテキストも fade-seg になる。
    expect(container.querySelector('table')).not.toBeNull();
    const cells = container.querySelectorAll('td');
    expect(cells.length).toBeGreaterThan(0);
    expect(container.querySelectorAll('td .fade-seg, th .fade-seg').length).toBeGreaterThan(0);
  });
});
