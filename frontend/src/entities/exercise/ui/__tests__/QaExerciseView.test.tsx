import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import QaExerciseView from '../QaExerciseView';
import type { MasterExercise, ExerciseSubmitResult } from '../../model/types';

// MarkdownView は remark 一式を読み込むため、ここでは本文が渡ることだけ確認する
vi.mock('@/components/message/MarkdownView', () => ({
  default: ({ content }: { content: string }) => <div data-testid="markdown">{content}</div>,
}));

const exercise = (over: Partial<MasterExercise> = {}): MasterExercise => ({
  title: 'コンテナ一覧',
  language: 'docker',
  difficulty: 2,
  orderIndex: 1,
  description: '起動中のコンテナを一覧表示するコマンドは？',
  expectedOutput: 'docker ps',
  explanation: '`docker ps` は起動中のコンテナだけを表示します。',
  ...over,
} as MasterExercise);

const defaults = {
  starterCode: '',
  onCodeChange: vi.fn(),
  submitting: false,
  submitResult: null as ExerciseSubmitResult | null,
  submitError: null as string | null,
  onSubmit: vi.fn(),
  onReset: vi.fn(),
};

const renderView = (props: Partial<typeof defaults> & { exercise?: MasterExercise } = {}) =>
  render(
    <QaExerciseView exercise={props.exercise ?? exercise()} {...{ ...defaults, ...props }} />,
    { wrapper: MemoryRouter }
  );

describe('QaExerciseView', () => {
  beforeEach(() => vi.clearAllMocks());

  it('問題文とコマンド入力欄を表示する', () => {
    renderView();
    expect(screen.getByText('起動中のコンテナを一覧表示するコマンドは？')).toBeInTheDocument();
    expect(screen.getByLabelText('コマンドを入力')).toBeInTheDocument();
  });

  it('入力を変更すると onCodeChange が呼ばれる', () => {
    const onCodeChange = vi.fn();
    renderView({ onCodeChange });
    fireEvent.change(screen.getByLabelText('コマンドを入力'), { target: { value: 'docker ps' } });
    expect(onCodeChange).toHaveBeenCalledWith('docker ps');
  });

  it('入力が空のときは解答ボタンを押せない', () => {
    renderView({ starterCode: '   ' });
    expect(screen.getByRole('button', { name: /解答する/ })).toBeDisabled();
  });

  it('入力があるとき submit で onSubmit が呼ばれる', () => {
    const onSubmit = vi.fn();
    renderView({ starterCode: 'docker ps', onSubmit });
    fireEvent.click(screen.getByRole('button', { name: /解答する/ }));
    expect(onSubmit).toHaveBeenCalledTimes(1);
  });

  it('採点中はボタン文言が「採点中...」になる', () => {
    renderView({ starterCode: 'docker ps', submitting: true });
    expect(screen.getByRole('button', { name: /採点中/ })).toBeInTheDocument();
  });

  it('リセットで onReset が呼ばれる', () => {
    const onReset = vi.fn();
    renderView({ onReset });
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(onReset).toHaveBeenCalledTimes(1);
  });

  it('正解時は正解バナーと解説を表示する', () => {
    renderView({ submitResult: { isCorrect: true } as ExerciseSubmitResult });
    expect(screen.getByRole('status')).toHaveTextContent('正解です。');
    expect(screen.getByText('回答は正解です')).toBeInTheDocument();
    expect(screen.getByText('docker ps')).toBeInTheDocument();
    expect(screen.getByText('説明')).toBeInTheDocument();
  });

  it('不正解時は再入力を促し、解説は出さない', () => {
    renderView({ submitResult: { isCorrect: false } as ExerciseSubmitResult });
    expect(screen.getByText(/もう一度入力してください/)).toBeInTheDocument();
    expect(screen.queryByText('回答は正解です')).not.toBeInTheDocument();
  });

  it('解説が無い正解では説明欄を出さない', () => {
    renderView({
      exercise: exercise({ explanation: '' }),
      submitResult: { isCorrect: true } as ExerciseSubmitResult,
    });
    expect(screen.getByText('回答は正解です')).toBeInTheDocument();
    expect(screen.queryByText('説明')).not.toBeInTheDocument();
  });

  it('submitError があれば alert として表示する', () => {
    renderView({ submitError: '通信に失敗しました' });
    expect(screen.getByRole('alert')).toHaveTextContent('通信に失敗しました');
  });
});
