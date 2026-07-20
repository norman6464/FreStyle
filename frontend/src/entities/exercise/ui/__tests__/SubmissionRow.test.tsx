import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import SubmissionRow from '../SubmissionRow';
import type { ExerciseSubmission } from '../../model/types';

const submission = (over: Partial<ExerciseSubmission> = {}): ExerciseSubmission => ({
  isCorrect: true,
  // ローカルタイムで解釈させる（getFullYear 等でそのまま整形されるため）
  submittedAt: '2026-07-20T09:05:00',
  ...over,
} as ExerciseSubmission);

const renderRow = (s: ExerciseSubmission) =>
  render(
    <ul>
      <SubmissionRow submission={s} />
    </ul>
  );

describe('SubmissionRow', () => {
  it('提出日時を 0 埋めして表示する', () => {
    renderRow(submission());
    expect(screen.getByText('2026/07/20 09:05')).toBeInTheDocument();
  });

  it('合格時は「合格」を表示する', () => {
    renderRow(submission());
    expect(screen.getByText('合格')).toBeInTheDocument();
  });

  it('不合格時は「不合格」を表示する', () => {
    renderRow(submission({ isCorrect: false }));
    expect(screen.getByText('不合格')).toBeInTheDocument();
  });
});
