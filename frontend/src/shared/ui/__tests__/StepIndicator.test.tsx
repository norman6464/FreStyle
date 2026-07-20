import { describe, it, expect } from 'vitest';
import { render, screen, within } from '@testing-library/react';
import StepIndicator from '../StepIndicator';

const STEPS = [
  { label: 'シナリオを選ぶ', description: '12 件から選択' },
  { label: 'AI と会話する' },
  { label: 'スコアを確認する' },
];

describe('StepIndicator', () => {
  it('すべてのステップラベルを描画する', () => {
    render(<StepIndicator steps={STEPS} currentStep={1} />);
    STEPS.forEach((step) => {
      expect(screen.getByText(step.label)).toBeInTheDocument();
    });
  });

  it('補足説明が指定されていれば描画する', () => {
    render(<StepIndicator steps={STEPS} currentStep={0} />);
    expect(screen.getByText('12 件から選択')).toBeInTheDocument();
  });

  it('現在アクティブなステップに aria-current="step" が付く', () => {
    render(<StepIndicator steps={STEPS} currentStep={1} />);
    const list = screen.getByRole('list', { name: '進行状況' });
    const items = within(list).getAllByRole('listitem');
    expect(items[0]).not.toHaveAttribute('aria-current');
    expect(items[1]).toHaveAttribute('aria-current', 'step');
    expect(items[2]).not.toHaveAttribute('aria-current');
  });

  it('完了済み / 実行中 / 未着手 をスクリーンリーダー向けに区別する', () => {
    render(<StepIndicator steps={STEPS} currentStep={1} />);
    expect(screen.getByText(/ステップ 1 \/ 3・完了/)).toBeInTheDocument();
    expect(screen.getByText(/ステップ 2 \/ 3・実行中/)).toBeInTheDocument();
    // 未着手ステップは完了・実行中の接尾辞が付かない（全角括弧で閉じられる）
    expect(screen.getByText('（ステップ 3 / 3）')).toBeInTheDocument();
  });

  it('ol 要素にラベル「進行状況」が付く', () => {
    render(<StepIndicator steps={STEPS} currentStep={0} />);
    expect(screen.getByRole('list', { name: '進行状況' })).toBeInTheDocument();
  });
});
