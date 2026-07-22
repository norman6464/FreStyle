import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import ExerciseHeader from '../ExerciseHeader';
import type { MasterExercise, ExerciseSubmitResult } from '../../model/types';

const exercise = (over: Partial<MasterExercise> = {}): MasterExercise => ({
  title: 'FizzBuzz',
  language: 'go',
  difficulty: 3,
  orderIndex: 7,
  ...over,
} as MasterExercise);

const renderHeader = (ex: MasterExercise, result: ExerciseSubmitResult | null = null) =>
  render(<ExerciseHeader exercise={ex} submitResult={result} />, { wrapper: MemoryRouter });

describe('ExerciseHeader', () => {
  it('タイトル・難易度・並び順を表示する', () => {
    renderHeader(exercise());
    expect(screen.getByRole('heading', { name: 'FizzBuzz' })).toBeInTheDocument();
    expect(screen.getByText('難易度 ★★★')).toBeInTheDocument();
    expect(screen.getByText('#7')).toBeInTheDocument();
  });

  it('難易度は 1〜5 に丸める', () => {
    const { unmount } = renderHeader(exercise({ difficulty: 0 }));
    expect(screen.getByText('難易度 ★')).toBeInTheDocument();
    unmount();
    renderHeader(exercise({ difficulty: 9 }));
    expect(screen.getByText('難易度 ★★★★★')).toBeInTheDocument();
  });

  it('採点結果がないときはバッジを出さない', () => {
    renderHeader(exercise());
    expect(screen.queryByText('全テストケース合格')).not.toBeInTheDocument();
    expect(screen.queryByText('不合格')).not.toBeInTheDocument();
  });

  it('採点結果があるときはバッジを出す', () => {
    renderHeader(exercise(), { isCorrect: true } as ExerciseSubmitResult);
    expect(screen.getByText('全テストケース合格')).toBeInTheDocument();
  });
});
