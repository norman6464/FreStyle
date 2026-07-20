import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import SubmitResultPanel from '../SubmitResultPanel';
import type { ExerciseTestCaseResult } from '../../model/types';

const result = (over: Partial<ExerciseTestCaseResult> = {}): ExerciseTestCaseResult => ({
  orderIndex: 1,
  passed: true,
  input: '3\n',
  expectedOutput: 'Fizz\n',
  actualOutput: 'Fizz\n',
  stderr: '',
  ...over,
} as ExerciseTestCaseResult);

describe('SubmitResultPanel', () => {
  it('合格数と総数を見出しに出す', () => {
    render(
      <SubmitResultPanel
        results={[result(), result({ orderIndex: 2, passed: false }), result({ orderIndex: 3 })]}
      />
    );
    expect(screen.getByText('テストケース採点結果 (2/3 合格)')).toBeInTheDocument();
  });

  it('テストケースごとに合否を表示する', () => {
    render(<SubmitResultPanel results={[result(), result({ orderIndex: 2, passed: false })]} />);
    expect(screen.getByText('テストケース 1')).toBeInTheDocument();
    expect(screen.getByText('テストケース 2')).toBeInTheDocument();
    expect(screen.getByText('合格')).toBeInTheDocument();
    expect(screen.getByText('不合格')).toBeInTheDocument();
  });

  it('入力・期待出力・実際の出力を表示する', () => {
    render(<SubmitResultPanel results={[result()]} />);
    expect(screen.getByText('入力')).toBeInTheDocument();
    expect(screen.getByText('期待出力')).toBeInTheDocument();
    expect(screen.getByText('実際の出力')).toBeInTheDocument();
  });

  it('入力が空のときは入力欄を出さない', () => {
    render(<SubmitResultPanel results={[result({ input: '' })]} />);
    expect(screen.queryByText('入力')).not.toBeInTheDocument();
  });

  it('実際の出力が空のときは (なし) と表示する', () => {
    render(<SubmitResultPanel results={[result({ actualOutput: '' })]} />);
    expect(screen.getByText('(なし)')).toBeInTheDocument();
  });

  it('stderr があるときだけ stderr 欄を出す', () => {
    const { unmount } = render(<SubmitResultPanel results={[result()]} />);
    expect(screen.queryByText('stderr')).not.toBeInTheDocument();
    unmount();
    render(<SubmitResultPanel results={[result({ stderr: 'panic: boom' })]} />);
    expect(screen.getByText('stderr')).toBeInTheDocument();
    expect(screen.getByText('panic: boom')).toBeInTheDocument();
  });
});
