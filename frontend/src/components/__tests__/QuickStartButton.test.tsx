import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import QuickStartButton from '../QuickStartButton';

const mockStartSession = vi.fn();
vi.mock('../../hooks/useStartPracticeSession', () => ({
  useStartPracticeSession: () => ({
    startSession: mockStartSession,
    starting: false,
  }),
}));

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

const defaultScenario = {
  id: 1,
  name: '本番障害の緊急報告',
  description: '本番環境で障害が発生し、顧客に緊急報告する場面です。',
  category: 'customer',
  roleName: '顧客',
  difficulty: 'intermediate',
};

describe('QuickStartButton', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('ボタンが表示される', () => {
    render(
      <MemoryRouter>
        <QuickStartButton scenario={null} />
      </MemoryRouter>
    );

    expect(screen.getByRole('button', { name: /練習を始める/i })).toBeInTheDocument();
  });

  it('推奨シナリオがある場合、シナリオ名が表示される', () => {
    render(
      <MemoryRouter>
        <QuickStartButton scenario={defaultScenario} />
      </MemoryRouter>
    );

    expect(screen.getByText('本番障害の緊急報告')).toBeInTheDocument();
  });

  it('推奨シナリオがある場合、クリックでstartSessionが呼ばれる', () => {
    render(
      <MemoryRouter>
        <QuickStartButton scenario={defaultScenario} />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByRole('button', { name: /練習を始める/i }));

    expect(mockStartSession).toHaveBeenCalledWith(defaultScenario);
  });

  it('推奨シナリオがない場合、クリックで練習ページに遷移する', () => {
    render(
      <MemoryRouter>
        <QuickStartButton scenario={null} />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByRole('button', { name: /練習を始める/i }));

    expect(mockNavigate).toHaveBeenCalledWith('/practice');
    expect(mockStartSession).not.toHaveBeenCalled();
  });

  it('推奨シナリオがない場合、「シナリオを選んで練習」と表示される', () => {
    render(
      <MemoryRouter>
        <QuickStartButton scenario={null} />
      </MemoryRouter>
    );

    expect(screen.getByText('シナリオを選んで練習')).toBeInTheDocument();
  });
});
