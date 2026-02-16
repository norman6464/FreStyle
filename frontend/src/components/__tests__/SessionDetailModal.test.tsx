import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import SessionDetailModal from '../SessionDetailModal';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}));

const session = {
  sessionId: 1,
  sessionTitle: '障害報告の練習',
  overallScore: 7.5,
  scores: [
    { axis: '論理的構成力', score: 8, comment: '構成が明確でした' },
    { axis: '配慮表現', score: 6, comment: 'もう少し丁寧に' },
    { axis: '要約力', score: 8.5, comment: '簡潔にまとめられていました' },
  ],
  createdAt: '2026-02-13T10:00:00',
};

describe('SessionDetailModal', () => {
  it('セッションタイトルと日時が表示される', () => {
    render(<SessionDetailModal session={session} onClose={vi.fn()} />);

    expect(screen.getByText('障害報告の練習')).toBeInTheDocument();
    expect(screen.getByText(/2026年/)).toBeInTheDocument();
  });

  it('総合スコアが表示される', () => {
    render(<SessionDetailModal session={session} onClose={vi.fn()} />);

    expect(screen.getByText('7.5')).toBeInTheDocument();
  });

  it('各軸のスコアとコメントが表示される', () => {
    render(<SessionDetailModal session={session} onClose={vi.fn()} />);

    expect(screen.getByText('論理的構成力')).toBeInTheDocument();
    expect(screen.getByText('構成が明確でした')).toBeInTheDocument();
    expect(screen.getByText('もう少し丁寧に')).toBeInTheDocument();
  });

  it('閉じるボタンでonCloseが呼ばれる', () => {
    const onClose = vi.fn();
    render(<SessionDetailModal session={session} onClose={onClose} />);

    fireEvent.click(screen.getByRole('button', { name: '閉じる' }));
    expect(onClose).toHaveBeenCalledOnce();
  });

  it('scoresがnullでもエラーにならない', () => {
    const nullSession = {
      ...session,
      scores: null as unknown as typeof session.scores,
    };
    render(<SessionDetailModal session={nullSession} onClose={vi.fn()} />);
    expect(screen.getByText('障害報告の練習')).toBeInTheDocument();
  });

  it('オーバーレイクリックでonCloseが呼ばれる', () => {
    const onClose = vi.fn();
    render(<SessionDetailModal session={session} onClose={onClose} />);

    fireEvent.click(screen.getByTestId('modal-overlay'));
    expect(onClose).toHaveBeenCalledOnce();
  });

  it('「会話を見る」ボタンをクリックするとセッション画面に遷移する', () => {
    const onClose = vi.fn();
    render(<SessionDetailModal session={session} onClose={onClose} />);

    fireEvent.click(screen.getByRole('button', { name: '会話を見る' }));
    expect(mockNavigate).toHaveBeenCalledWith('/chat/ask-ai/1');
  });
});
