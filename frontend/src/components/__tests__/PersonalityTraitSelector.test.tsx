import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import PersonalityTraitSelector from '../PersonalityTraitSelector';

const traits = ['社交的', '慎重', '論理的', '共感的', '積極的'];

describe('PersonalityTraitSelector', () => {
  it('すべての性格特性が表示される', () => {
    render(
      <PersonalityTraitSelector
        options={traits}
        selected={[]}
        onToggle={vi.fn()}
      />
    );
    traits.forEach((trait) => {
      expect(screen.getByText(trait)).toBeInTheDocument();
    });
  });

  it('選択済みの特性にアクティブスタイルが適用される', () => {
    render(
      <PersonalityTraitSelector
        options={traits}
        selected={['社交的']}
        onToggle={vi.fn()}
      />
    );
    const button = screen.getByText('社交的');
    expect(button.className).toContain('bg-primary-500');
    expect(button.className).toContain('text-white');
  });

  it('未選択の特性に非アクティブスタイルが適用される', () => {
    render(
      <PersonalityTraitSelector
        options={traits}
        selected={[]}
        onToggle={vi.fn()}
      />
    );
    const button = screen.getByText('社交的');
    expect(button.className).toContain('bg-surface-3');
  });

  it('クリックでonToggleが呼ばれる', () => {
    const onToggle = vi.fn();
    render(
      <PersonalityTraitSelector
        options={traits}
        selected={[]}
        onToggle={onToggle}
      />
    );
    fireEvent.click(screen.getByText('論理的'));
    expect(onToggle).toHaveBeenCalledWith('論理的');
  });

  it('ラベルが表示される', () => {
    render(
      <PersonalityTraitSelector
        options={traits}
        selected={[]}
        onToggle={vi.fn()}
        label="性格特性"
      />
    );
    expect(screen.getByText('性格特性')).toBeInTheDocument();
  });

  it('ボタンがtype=buttonである', () => {
    render(
      <PersonalityTraitSelector
        options={traits}
        selected={[]}
        onToggle={vi.fn()}
      />
    );
    const button = screen.getByText('社交的');
    expect(button).toHaveAttribute('type', 'button');
  });

  it('複数の特性を選択できる', () => {
    render(
      <PersonalityTraitSelector
        options={traits}
        selected={['社交的', '論理的']}
        onToggle={vi.fn()}
      />
    );
    expect(screen.getByText('社交的').className).toContain('bg-primary-500');
    expect(screen.getByText('論理的').className).toContain('bg-primary-500');
    expect(screen.getByText('慎重').className).toContain('bg-surface-3');
  });
});
