import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import ScenarioCard from '../ScenarioCard';
import type { PracticeScenario } from '../../types';

const mockScenario: PracticeScenario = {
  id: 1,
  name: '本番障害の緊急報告',
  description: '本番環境で障害が発生し、顧客に緊急報告する場面です。',
  category: '顧客折衝',
  roleName: '不満を持つ顧客担当者',
  difficulty: 'intermediate',
};

describe('ScenarioCard', () => {
  it('シナリオ名を表示する', () => {
    render(<ScenarioCard scenario={mockScenario} onSelect={vi.fn()} />);
    expect(screen.getByText('本番障害の緊急報告')).toBeDefined();
  });

  it('シナリオの説明を表示する', () => {
    render(<ScenarioCard scenario={mockScenario} onSelect={vi.fn()} />);
    expect(screen.getByText('本番環境で障害が発生し、顧客に緊急報告する場面です。')).toBeDefined();
  });

  it('相手役名を表示する', () => {
    render(<ScenarioCard scenario={mockScenario} onSelect={vi.fn()} />);
    expect(screen.getByText(/不満を持つ顧客担当者/)).toBeDefined();
  });

  it('難易度を日本語で表示する', () => {
    render(<ScenarioCard scenario={mockScenario} onSelect={vi.fn()} />);
    expect(screen.getByText('中級')).toBeDefined();
  });

  it('初級の難易度を表示する', () => {
    const beginnerScenario = { ...mockScenario, difficulty: 'beginner' };
    render(<ScenarioCard scenario={beginnerScenario} onSelect={vi.fn()} />);
    expect(screen.getByText('初級')).toBeDefined();
  });

  it('上級の難易度を表示する', () => {
    const advancedScenario = { ...mockScenario, difficulty: 'advanced' };
    render(<ScenarioCard scenario={advancedScenario} onSelect={vi.fn()} />);
    expect(screen.getByText('上級')).toBeDefined();
  });

  it('カテゴリを表示する', () => {
    render(<ScenarioCard scenario={mockScenario} onSelect={vi.fn()} />);
    expect(screen.getByText('顧客折衝')).toBeDefined();
  });

  it('クリック時にonSelectが呼ばれる', async () => {
    const onSelect = vi.fn();
    render(<ScenarioCard scenario={mockScenario} onSelect={onSelect} />);
    const card = screen.getByText('本番障害の緊急報告').closest('div[class*="cursor-pointer"]');
    card?.click();
    expect(onSelect).toHaveBeenCalledWith(mockScenario);
  });
});
