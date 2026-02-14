import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import PersonalBestCard from '../PersonalBestCard';

const mockHistory = [
  { sessionId: 1, overallScore: 7.5, createdAt: '2025-01-10T10:00:00' },
  { sessionId: 2, overallScore: 9.2, createdAt: '2025-01-15T14:00:00' },
  { sessionId: 3, overallScore: 8.0, createdAt: '2025-01-20T09:00:00' },
];

describe('PersonalBestCard', () => {
  it('自己ベストスコアが表示される', () => {
    render(<PersonalBestCard history={mockHistory} />);
    expect(screen.getByText('9.2')).toBeInTheDocument();
  });

  it('自己ベストのタイトルが表示される', () => {
    render(<PersonalBestCard history={mockHistory} />);
    expect(screen.getByText('自己ベスト')).toBeInTheDocument();
  });

  it('達成日が表示される', () => {
    render(<PersonalBestCard history={mockHistory} />);
    expect(screen.getByText('2025/1/15')).toBeInTheDocument();
  });

  it('履歴が空の場合は何も表示しない', () => {
    const { container } = render(<PersonalBestCard history={[]} />);
    expect(container.firstChild).toBeNull();
  });

  it('単一セッションでも表示される', () => {
    render(<PersonalBestCard history={[mockHistory[0]]} />);
    expect(screen.getByText('7.5')).toBeInTheDocument();
  });

  it('トロフィーアイコンが表示される', () => {
    const { container } = render(<PersonalBestCard history={mockHistory} />);
    const svg = container.querySelector('svg');
    expect(svg).toBeTruthy();
  });

  it('達成ラベルが表示される', () => {
    render(<PersonalBestCard history={mockHistory} />);
    expect(screen.getByText('達成日')).toBeInTheDocument();
  });
});
