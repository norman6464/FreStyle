import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import ScenarioCard from '../ScenarioCard';
import type { PracticeScenario } from '../../types';

const mockScenario: PracticeScenario = {
  id: 1,
  name: '本番障害の緊急報告',
  description: '本番環境で障害が発生し、顧客に緊急報告する場面です。',
  category: 'customer',
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

  it('カテゴリを日本語で表示する', () => {
    render(<ScenarioCard scenario={mockScenario} onSelect={vi.fn()} />);
    expect(screen.getByText('顧客折衝')).toBeDefined();
  });

  it('seniorカテゴリを日本語で表示する', () => {
    const seniorScenario = { ...mockScenario, category: 'senior' };
    render(<ScenarioCard scenario={seniorScenario} onSelect={vi.fn()} />);
    expect(screen.getByText('シニア・上司')).toBeDefined();
  });

  it('teamカテゴリを日本語で表示する', () => {
    const teamScenario = { ...mockScenario, category: 'team' };
    render(<ScenarioCard scenario={teamScenario} onSelect={vi.fn()} />);
    expect(screen.getByText('チーム内')).toBeDefined();
  });

  it('所要時間の目安を表示する', () => {
    render(<ScenarioCard scenario={mockScenario} onSelect={vi.fn()} />);
    expect(screen.getByText(/約5〜10分/)).toBeDefined();
  });

  it('難易度の説明テキストを表示する', () => {
    render(<ScenarioCard scenario={mockScenario} onSelect={vi.fn()} />);
    expect(screen.getByText(/利害関係の調整/)).toBeDefined();
  });

  it('初級の難易度説明を表示する', () => {
    const beginnerScenario = { ...mockScenario, difficulty: 'beginner' };
    render(<ScenarioCard scenario={beginnerScenario} onSelect={vi.fn()} />);
    expect(screen.getByText(/基本的な報連相/)).toBeDefined();
  });

  it('クリック時にonSelectが呼ばれる', async () => {
    const onSelect = vi.fn();
    render(<ScenarioCard scenario={mockScenario} onSelect={onSelect} />);
    const card = screen.getByText('本番障害の緊急報告').closest('div[class*="cursor-pointer"]');
    card?.click();
    expect(onSelect).toHaveBeenCalledWith(mockScenario);
  });

  it('ブックマークボタンを表示する', () => {
    render(<ScenarioCard scenario={mockScenario} onSelect={vi.fn()} isBookmarked={false} onToggleBookmark={vi.fn()} />);
    expect(screen.getByTitle('ブックマーク')).toBeDefined();
  });

  it('ブックマーク済みの場合は塗りつぶしアイコンを表示する', () => {
    render(<ScenarioCard scenario={mockScenario} onSelect={vi.fn()} isBookmarked={true} onToggleBookmark={vi.fn()} />);
    expect(screen.getByTitle('ブックマーク解除')).toBeDefined();
  });

  it('初級バッジにemeraldの色が適用される', () => {
    const beginnerScenario = { ...mockScenario, difficulty: 'beginner' };
    render(<ScenarioCard scenario={beginnerScenario} onSelect={vi.fn()} />);
    const badge = screen.getByText('初級');
    expect(badge.className).toContain('text-emerald-400');
  });

  it('中級バッジにamberの色が適用される', () => {
    render(<ScenarioCard scenario={mockScenario} onSelect={vi.fn()} />);
    const badge = screen.getByText('中級');
    expect(badge.className).toContain('text-amber-400');
  });

  it('上級バッジにroseの色が適用される', () => {
    const advancedScenario = { ...mockScenario, difficulty: 'advanced' };
    render(<ScenarioCard scenario={advancedScenario} onSelect={vi.fn()} />);
    const badge = screen.getByText('上級');
    expect(badge.className).toContain('text-rose-400');
  });
});
