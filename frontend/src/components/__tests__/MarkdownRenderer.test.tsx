import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import MarkdownRenderer from '../MarkdownRenderer';

describe('MarkdownRenderer', () => {
  it('箇条書き（- ）をリストとしてレンダリングする', () => {
    render(<MarkdownRenderer content={'- りんご\n- みかん\n- ぶどう'} />);
    const items = screen.getAllByRole('listitem');
    expect(items).toHaveLength(3);
    expect(items[0]).toHaveTextContent('りんご');
  });

  it('箇条書き（・）をリストとしてレンダリングする', () => {
    render(<MarkdownRenderer content={'・りんご\n・みかん'} />);
    const items = screen.getAllByRole('listitem');
    expect(items).toHaveLength(2);
  });

  it('番号付きリスト（1. ）をレンダリングする', () => {
    render(<MarkdownRenderer content={'1. 手順1\n2. 手順2\n3. 手順3'} />);
    const list = screen.getByRole('list');
    expect(list.tagName).toBe('OL');
    const items = screen.getAllByRole('listitem');
    expect(items).toHaveLength(3);
  });

  it('見出し（#）をレンダリングする', () => {
    render(<MarkdownRenderer content={'# 見出し1\n## 見出し2\n### 見出し3'} />);
    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('見出し1');
    expect(screen.getByRole('heading', { level: 2 })).toHaveTextContent('見出し2');
    expect(screen.getByRole('heading', { level: 3 })).toHaveTextContent('見出し3');
  });

  it('太字（**text**）をレンダリングする', () => {
    render(<MarkdownRenderer content="これは**太字**です" />);
    const bold = screen.getByText('太字');
    expect(bold.tagName).toBe('STRONG');
  });

  it('斜体（*text*）をレンダリングする', () => {
    render(<MarkdownRenderer content="これは*斜体*です" />);
    const italic = screen.getByText('斜体');
    expect(italic.tagName).toBe('EM');
  });

  it('通常テキストをそのまま表示する', () => {
    render(<MarkdownRenderer content="普通のテキスト" />);
    expect(screen.getByText('普通のテキスト')).toBeInTheDocument();
  });

  it('空文字列でもエラーにならない', () => {
    const { container } = render(<MarkdownRenderer content="" />);
    expect(container).toBeTruthy();
  });

  it('複数行の通常テキストを段落としてレンダリングする', () => {
    render(<MarkdownRenderer content={'行1\n行2'} />);
    expect(screen.getByText('行1')).toBeInTheDocument();
    expect(screen.getByText('行2')).toBeInTheDocument();
  });
});
