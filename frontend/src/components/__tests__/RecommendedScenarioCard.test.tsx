import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import RecommendedScenarioCard from '../RecommendedScenarioCard';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return { ...actual, useNavigate: () => mockNavigate };
});

describe('RecommendedScenarioCard', () => {
  it('シナリオ情報が表示される', () => {
    render(
      <MemoryRouter>
        <RecommendedScenarioCard
          scenario={{ id: 1, name: '本番障害の緊急報告', description: '説明文', category: 'customer', roleName: '顧客', difficulty: 'intermediate' }}
          weakAxis="論理的構成力"
        />
      </MemoryRouter>
    );

    expect(screen.getByText('本番障害の緊急報告')).toBeInTheDocument();
    expect(screen.getByText(/論理的構成力/)).toBeInTheDocument();
  });

  it('弱点軸名が表示される', () => {
    render(
      <MemoryRouter>
        <RecommendedScenarioCard
          scenario={{ id: 2, name: 'テストシナリオ', description: '説明', category: 'senior', roleName: '上司', difficulty: 'beginner' }}
          weakAxis="配慮表現"
        />
      </MemoryRouter>
    );

    expect(screen.getByText(/配慮表現/)).toBeInTheDocument();
  });

  it('チャレンジボタンが表示される', () => {
    render(
      <MemoryRouter>
        <RecommendedScenarioCard
          scenario={{ id: 1, name: 'テスト', description: '説明', category: 'customer', roleName: '顧客', difficulty: 'intermediate' }}
          weakAxis="提案力"
        />
      </MemoryRouter>
    );

    expect(screen.getByText('このシナリオで練習する')).toBeInTheDocument();
  });

  it('ボタンクリックで練習ページに遷移する', () => {
    render(
      <MemoryRouter>
        <RecommendedScenarioCard
          scenario={{ id: 1, name: 'テスト', description: '説明', category: 'customer', roleName: '顧客', difficulty: 'intermediate' }}
          weakAxis="提案力"
        />
      </MemoryRouter>
    );

    screen.getByText('このシナリオで練習する').click();
    expect(mockNavigate).toHaveBeenCalledWith('/practice');
  });

  it('難易度バッジが表示される', () => {
    render(
      <MemoryRouter>
        <RecommendedScenarioCard
          scenario={{ id: 1, name: 'テスト', description: '説明', category: 'customer', roleName: '顧客', difficulty: 'intermediate' }}
          weakAxis="要約力"
        />
      </MemoryRouter>
    );

    expect(screen.getByText('中級')).toBeInTheDocument();
  });
});
