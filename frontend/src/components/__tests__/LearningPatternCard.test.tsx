import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import LearningPatternCard from '../LearningPatternCard';

describe('LearningPatternCard', () => {
  it('練習日がない場合は何も表示しない', () => {
    const { container } = render(<LearningPatternCard practiceDates={[]} />);
    expect(container.firstChild).toBeNull();
  });

  it('学習パターンタイトルが表示される', () => {
    const dates = ['2026-02-10T10:00:00Z', '2026-02-11T10:00:00Z'];
    render(<LearningPatternCard practiceDates={dates} />);
    expect(screen.getByText('学習パターン')).toBeInTheDocument();
  });

  it('曜日ラベルが表示される', () => {
    const dates = ['2026-02-10T10:00:00Z'];
    render(<LearningPatternCard practiceDates={dates} />);
    expect(screen.getByText('月')).toBeInTheDocument();
    expect(screen.getByText('日')).toBeInTheDocument();
  });

  it('平日集中型パターンが判定される', () => {
    // 月〜金の日付を複数設定
    const dates = [
      '2026-02-09T10:00:00Z', // 月
      '2026-02-10T10:00:00Z', // 火
      '2026-02-11T10:00:00Z', // 水
      '2026-02-12T10:00:00Z', // 木
      '2026-02-13T10:00:00Z', // 金
    ];
    render(<LearningPatternCard practiceDates={dates} />);
    expect(screen.getByText('平日集中型')).toBeInTheDocument();
  });

  it('週末集中型パターンが判定される', () => {
    // 土日の日付を複数設定
    const dates = [
      '2026-02-07T10:00:00Z', // 土
      '2026-02-08T10:00:00Z', // 日
      '2026-02-14T10:00:00Z', // 土
      '2026-02-15T10:00:00Z', // 日
    ];
    render(<LearningPatternCard practiceDates={dates} />);
    expect(screen.getByText('週末集中型')).toBeInTheDocument();
  });

  it('毎日コツコツ型パターンが判定される', () => {
    // 全曜日にまんべんなく分布
    const dates = [
      '2026-02-09T10:00:00Z', // 月
      '2026-02-10T10:00:00Z', // 火
      '2026-02-11T10:00:00Z', // 水
      '2026-02-12T10:00:00Z', // 木
      '2026-02-13T10:00:00Z', // 金
      '2026-02-14T10:00:00Z', // 土
      '2026-02-15T10:00:00Z', // 日
    ];
    render(<LearningPatternCard practiceDates={dates} />);
    expect(screen.getByText('毎日コツコツ型')).toBeInTheDocument();
  });

  it('不定期型パターンが判定される', () => {
    // 少数の曜日に偏らず、全体的に少ない
    const dates = [
      '2026-02-10T10:00:00Z', // 火
      '2026-02-14T10:00:00Z', // 土
    ];
    render(<LearningPatternCard practiceDates={dates} />);
    expect(screen.getByText('不定期型')).toBeInTheDocument();
  });

  it('パターンに応じた説明メッセージが表示される', () => {
    const dates = [
      '2026-02-09T10:00:00Z',
      '2026-02-10T10:00:00Z',
      '2026-02-11T10:00:00Z',
      '2026-02-12T10:00:00Z',
      '2026-02-13T10:00:00Z',
    ];
    render(<LearningPatternCard practiceDates={dates} />);
    expect(screen.getByText(/平日の空き時間/)).toBeInTheDocument();
  });
});
