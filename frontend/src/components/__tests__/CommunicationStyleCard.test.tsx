import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import CommunicationStyleCard from '../CommunicationStyleCard';

interface AxisScore {
  axis: string;
  score: number;
  comment: string;
}

interface Session {
  scores: AxisScore[];
}

describe('CommunicationStyleCard', () => {
  it('セッションがない場合は未判定メッセージを表示する', () => {
    render(<CommunicationStyleCard sessions={[]} />);
    expect(screen.getByText('まだスタイルが判定できません')).toBeInTheDocument();
  });

  it('論理的構成力が最高の場合は論理型を表示する', () => {
    const sessions: Session[] = [{
      scores: [
        { axis: '論理的構成力', score: 9, comment: '' },
        { axis: '配慮表現', score: 5, comment: '' },
        { axis: '要約力', score: 5, comment: '' },
        { axis: '提案力', score: 5, comment: '' },
        { axis: '質問・傾聴力', score: 5, comment: '' },
      ],
    }];
    render(<CommunicationStyleCard sessions={sessions} />);
    expect(screen.getByText('論理型コミュニケーター')).toBeInTheDocument();
  });

  it('配慮表現が最高の場合は共感型を表示する', () => {
    const sessions: Session[] = [{
      scores: [
        { axis: '論理的構成力', score: 5, comment: '' },
        { axis: '配慮表現', score: 9, comment: '' },
        { axis: '要約力', score: 5, comment: '' },
        { axis: '提案力', score: 5, comment: '' },
        { axis: '質問・傾聴力', score: 5, comment: '' },
      ],
    }];
    render(<CommunicationStyleCard sessions={sessions} />);
    expect(screen.getByText('共感型コミュニケーター')).toBeInTheDocument();
  });

  it('要約力が最高の場合は簡潔型を表示する', () => {
    const sessions: Session[] = [{
      scores: [
        { axis: '論理的構成力', score: 5, comment: '' },
        { axis: '配慮表現', score: 5, comment: '' },
        { axis: '要約力', score: 9, comment: '' },
        { axis: '提案力', score: 5, comment: '' },
        { axis: '質問・傾聴力', score: 5, comment: '' },
      ],
    }];
    render(<CommunicationStyleCard sessions={sessions} />);
    expect(screen.getByText('簡潔型コミュニケーター')).toBeInTheDocument();
  });

  it('提案力が最高の場合は提案型を表示する', () => {
    const sessions: Session[] = [{
      scores: [
        { axis: '論理的構成力', score: 5, comment: '' },
        { axis: '配慮表現', score: 5, comment: '' },
        { axis: '要約力', score: 5, comment: '' },
        { axis: '提案力', score: 9, comment: '' },
        { axis: '質問・傾聴力', score: 5, comment: '' },
      ],
    }];
    render(<CommunicationStyleCard sessions={sessions} />);
    expect(screen.getByText('提案型コミュニケーター')).toBeInTheDocument();
  });

  it('質問・傾聴力が最高の場合は傾聴型を表示する', () => {
    const sessions: Session[] = [{
      scores: [
        { axis: '論理的構成力', score: 5, comment: '' },
        { axis: '配慮表現', score: 5, comment: '' },
        { axis: '要約力', score: 5, comment: '' },
        { axis: '提案力', score: 5, comment: '' },
        { axis: '質問・傾聴力', score: 9, comment: '' },
      ],
    }];
    render(<CommunicationStyleCard sessions={sessions} />);
    expect(screen.getByText('傾聴型コミュニケーター')).toBeInTheDocument();
  });

  it('複数セッションの平均からスタイルを判定する', () => {
    const sessions: Session[] = [
      {
        scores: [
          { axis: '論理的構成力', score: 8, comment: '' },
          { axis: '配慮表現', score: 6, comment: '' },
          { axis: '要約力', score: 7, comment: '' },
          { axis: '提案力', score: 5, comment: '' },
          { axis: '質問・傾聴力', score: 6, comment: '' },
        ],
      },
      {
        scores: [
          { axis: '論理的構成力', score: 6, comment: '' },
          { axis: '配慮表現', score: 4, comment: '' },
          { axis: '要約力', score: 9, comment: '' },
          { axis: '提案力', score: 5, comment: '' },
          { axis: '質問・傾聴力', score: 6, comment: '' },
        ],
      },
    ];
    // 平均: 論理7, 配慮5, 要約8, 提案5, 傾聴6 → 要約力が最高 → 簡潔型
    render(<CommunicationStyleCard sessions={sessions} />);
    expect(screen.getByText('簡潔型コミュニケーター')).toBeInTheDocument();
  });
});
