import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import Badge from '../Badge';

describe('Badge', () => {
  it('テキストが表示される', () => {
    render(<Badge>初級</Badge>);
    expect(screen.getByText('初級')).toBeInTheDocument();
  });

  it('デフォルトでneutralバリアントが適用される', () => {
    render(<Badge>テスト</Badge>);
    const badge = screen.getByText('テスト');
    expect(badge.className).toContain('bg-surface-3');
  });

  it('successバリアントが適用される', () => {
    render(<Badge variant="success">成功</Badge>);
    const badge = screen.getByText('成功');
    expect(badge.className).toContain('bg-emerald-900/30');
    expect(badge.className).toContain('text-emerald-400');
  });

  it('warningバリアントが適用される', () => {
    render(<Badge variant="warning">警告</Badge>);
    const badge = screen.getByText('警告');
    expect(badge.className).toContain('bg-amber-900/30');
    expect(badge.className).toContain('text-amber-400');
  });

  it('dangerバリアントが適用される', () => {
    render(<Badge variant="danger">危険</Badge>);
    const badge = screen.getByText('危険');
    expect(badge.className).toContain('bg-rose-900/30');
    expect(badge.className).toContain('text-rose-400');
  });

  it('infoバリアントが適用される', () => {
    render(<Badge variant="info">情報</Badge>);
    const badge = screen.getByText('情報');
    expect(badge.className).toContain('bg-blue-900/30');
    expect(badge.className).toContain('text-blue-400');
  });

  it('デフォルトでsmサイズが適用される', () => {
    render(<Badge>小</Badge>);
    const badge = screen.getByText('小');
    expect(badge.className).toContain('text-[10px]');
    expect(badge.className).toContain('px-2');
    expect(badge.className).toContain('py-0.5');
  });

  it('mdサイズが適用される', () => {
    render(<Badge size="md">中</Badge>);
    const badge = screen.getByText('中');
    expect(badge.className).toContain('text-xs');
    expect(badge.className).toContain('px-2.5');
    expect(badge.className).toContain('py-1');
  });

  it('rounded-fullクラスが適用される', () => {
    render(<Badge>丸</Badge>);
    const badge = screen.getByText('丸');
    expect(badge.className).toContain('rounded-full');
  });

  it('追加のclassNameが適用される', () => {
    render(<Badge className="ml-2">追加</Badge>);
    const badge = screen.getByText('追加');
    expect(badge.className).toContain('ml-2');
  });
});
