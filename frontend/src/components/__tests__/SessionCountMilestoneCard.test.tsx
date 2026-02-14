import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import SessionCountMilestoneCard from '../SessionCountMilestoneCard';

describe('SessionCountMilestoneCard', () => {
  it('現在のセッション数が表示される', () => {
    render(<SessionCountMilestoneCard sessionCount={7} />);
    expect(screen.getByText('7')).toBeInTheDocument();
  });

  it('次のマイルストーンが表示される', () => {
    render(<SessionCountMilestoneCard sessionCount={7} />);
    expect(screen.getByText(/10/)).toBeInTheDocument();
  });

  it('タイトルが表示される', () => {
    render(<SessionCountMilestoneCard sessionCount={7} />);
    expect(screen.getByText('セッション達成')).toBeInTheDocument();
  });

  it('0セッションの場合も表示される', () => {
    render(<SessionCountMilestoneCard sessionCount={0} />);
    expect(screen.getByText('0')).toBeInTheDocument();
  });

  it('マイルストーン達成時に祝福メッセージが表示される', () => {
    render(<SessionCountMilestoneCard sessionCount={100} />);
    expect(screen.getByText(/100回以上の練習を達成/)).toBeInTheDocument();
  });

  it('プログレスバーが表示される', () => {
    const { container } = render(<SessionCountMilestoneCard sessionCount={5} />);
    const progressBar = container.querySelector('[role="progressbar"]');
    expect(progressBar).toBeTruthy();
  });

  it('アイコンが表示される', () => {
    const { container } = render(<SessionCountMilestoneCard sessionCount={5} />);
    const svg = container.querySelector('svg');
    expect(svg).toBeTruthy();
  });
});
