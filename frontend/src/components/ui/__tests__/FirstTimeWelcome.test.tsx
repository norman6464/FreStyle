import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import FirstTimeWelcome from '../FirstTimeWelcome';
import { createMockStorage } from '../../../test/mockStorage';

const STEPS = [
  { title: '練習モードを選ぶ', description: '12 のシナリオから選択' },
  { title: 'AI と会話する', description: '実務を想定したロールプレイ' },
  { title: 'スコアを確認する', description: '5 軸評価で弱点がわかる' },
];

describe('FirstTimeWelcome', () => {
  beforeEach(() => {
    vi.stubGlobal('localStorage', createMockStorage());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('既定タイトルを h2 として描画する', () => {
    render(<FirstTimeWelcome steps={STEPS} />);
    const heading = screen.getByRole('heading', { name: 'ようこそ FreStyle へ' });
    expect(heading.tagName).toBe('H2');
  });

  it('全ステップを番号付きで描画する', () => {
    render(<FirstTimeWelcome steps={STEPS} />);
    STEPS.forEach((step) => {
      expect(screen.getByText(step.title)).toBeInTheDocument();
      expect(screen.getByText(step.description)).toBeInTheDocument();
    });
    // 1, 2, 3 の番号が見える
    expect(screen.getByText('1')).toBeInTheDocument();
    expect(screen.getByText('2')).toBeInTheDocument();
    expect(screen.getByText('3')).toBeInTheDocument();
  });

  it('onPrimaryAction が未指定なら CTA ボタンを出さない', () => {
    render(<FirstTimeWelcome steps={STEPS} />);
    expect(screen.queryByRole('button', { name: /はじめて練習する/ })).not.toBeInTheDocument();
  });

  it('onPrimaryAction が指定されれば CTA ボタンを出し、クリックで呼ばれる', () => {
    const onPrimary = vi.fn();
    render(<FirstTimeWelcome steps={STEPS} onPrimaryAction={onPrimary} />);
    fireEvent.click(screen.getByRole('button', { name: 'はじめて練習する' }));
    expect(onPrimary).toHaveBeenCalledOnce();
  });

  it('閉じるボタンでカードが消え、storageKey があれば永続化される', () => {
    render(<FirstTimeWelcome steps={STEPS} storageKey="welcome:menu" />);
    fireEvent.click(screen.getByRole('button', { name: 'このカードを閉じる' }));
    expect(screen.queryByRole('heading', { name: 'ようこそ FreStyle へ' })).not.toBeInTheDocument();
    expect(window.localStorage.getItem('welcome:menu')).toBe('dismissed');
  });

  it('既に dismissed 済みなら初回から非表示', () => {
    window.localStorage.setItem('welcome:menu', 'dismissed');
    render(<FirstTimeWelcome steps={STEPS} storageKey="welcome:menu" />);
    expect(screen.queryByRole('heading', { name: 'ようこそ FreStyle へ' })).not.toBeInTheDocument();
  });
});
