import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import BookmarkedScenariosCard from '../BookmarkedScenariosCard';

vi.mock('../../hooks/useBookmarkedScenarios', () => ({
  useBookmarkedScenarios: vi.fn(),
}));

vi.mock('../../hooks/useStartPracticeSession', () => ({
  useStartPracticeSession: () => ({ startSession: vi.fn(), starting: false }),
}));

import { useBookmarkedScenarios } from '../../hooks/useBookmarkedScenarios';

const mockScenarios = [
  { id: 1, name: 'クレーム対応', description: '説明1', category: 'ビジネス', roleName: '顧客', difficulty: '中級', systemPrompt: '' },
  { id: 2, name: '会議ファシリテーション', description: '説明2', category: 'ビジネス', roleName: '同僚', difficulty: '上級', systemPrompt: '' },
];

describe('BookmarkedScenariosCard', () => {
  it('タイトルが表示される', () => {
    vi.mocked(useBookmarkedScenarios).mockReturnValue({ scenarios: mockScenarios, loading: false });

    render(<BrowserRouter><BookmarkedScenariosCard /></BrowserRouter>);

    expect(screen.getByText('ブックマーク済みシナリオ')).toBeInTheDocument();
  });

  it('シナリオ名が表示される', () => {
    vi.mocked(useBookmarkedScenarios).mockReturnValue({ scenarios: mockScenarios, loading: false });

    render(<BrowserRouter><BookmarkedScenariosCard /></BrowserRouter>);

    expect(screen.getByText('クレーム対応')).toBeInTheDocument();
    expect(screen.getByText('会議ファシリテーション')).toBeInTheDocument();
  });

  it('シナリオのカテゴリと難易度が表示される', () => {
    vi.mocked(useBookmarkedScenarios).mockReturnValue({ scenarios: mockScenarios, loading: false });

    render(<BrowserRouter><BookmarkedScenariosCard /></BrowserRouter>);

    expect(screen.getByText('ビジネス・中級')).toBeInTheDocument();
    expect(screen.getByText('ビジネス・上級')).toBeInTheDocument();
  });

  it('ブックマークがない場合は何も表示しない', () => {
    vi.mocked(useBookmarkedScenarios).mockReturnValue({ scenarios: [], loading: false });

    const { container } = render(<BrowserRouter><BookmarkedScenariosCard /></BrowserRouter>);

    expect(container.firstChild).toBeNull();
  });

  it('ローディング中は何も表示しない', () => {
    vi.mocked(useBookmarkedScenarios).mockReturnValue({ scenarios: [], loading: true });

    const { container } = render(<BrowserRouter><BookmarkedScenariosCard /></BrowserRouter>);

    expect(container.firstChild).toBeNull();
  });
});
