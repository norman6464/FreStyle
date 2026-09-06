import { useState } from 'react';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';
import { withRouter } from '../../../../.storybook/decorators';
import QaExerciseView from './QaExerciseView';
import type { MasterExercise } from '../model/types';

/**
 * 「コマンドを書き取る」形式の演習の画面。
 *
 * docker や kubernetes のように、その場で安全に動かすのが難しい題材のための形。
 * コードを実行せず、打った 1 行と答えを直接くらべる。
 *
 * 間違えても打った内容は消さない。消すと打ち直しになり、どこが惜しかったのかも分からなくなる。
 */
const meta = {
  title: 'entities/exercise/QaExerciseView',
  component: QaExerciseView,
  parameters: { layout: 'fullscreen' },
  decorators: [withRouter],
} satisfies Meta<typeof QaExerciseView>;

export default meta;
type Story = StoryObj<typeof meta>;

const exercise: MasterExercise = {
  id: 42,
  slug: 'docker-ps-all',
  language: 'docker',
  orderIndex: 3,
  category: 'コンテナ',
  title: '停止中のコンテナも含めて一覧を出す',
  description: '停止しているものも含め、すべてのコンテナを一覧するコマンドを書いてください。',
  starterCode: '',
  hintText: '-a を付けます。',
  expectedOutput: 'docker ps -a',
  mode: 'qa',
  explanation:
    '`docker ps` は動いているコンテナだけを出します。`-a`（--all）を付けると、停止しているものも含めて全部出ます。',
  difficulty: 1,
  isPublished: true,
  createdAt: '2026-09-01T00:00:00Z',
  updatedAt: '2026-09-01T00:00:00Z',
};

const baseArgs = {
  exercise,
  starterCode: '',
  onCodeChange: fn(),
  submitting: false,
  submitResult: null,
  submitError: null,
  onSubmit: fn(),
  onReset: fn(),
};

/** 開いた直後。まだ何も打っていない。 */
export const 打つ前: Story = {
  args: baseArgs,
  play: async ({ canvasElement }) => {
    await expect(
      within(canvasElement).getByRole('heading', { name: '停止中のコンテナも含めて一覧を出す' }),
    ).toBeVisible();
  },
};

/** 打っている途中。 */
export const 打っている途中: Story = {
  args: { ...baseArgs, starterCode: 'docker ps' },
};

/** 送っている最中。二重に送れないようにする。 */
export const 送信中: Story = {
  args: { ...baseArgs, starterCode: 'docker ps -a', submitting: true },
};

/** 正解したとき。解説が出る。 */
export const 正解: Story = {
  args: {
    ...baseArgs,
    starterCode: 'docker ps -a',
    submitResult: { submissionId: 1, isCorrect: true, results: [] },
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText(/--all/)).toBeVisible();
  },
};

/** 間違えたとき。打った内容はそのまま残る。 */
export const 不正解: Story = {
  args: {
    ...baseArgs,
    starterCode: 'docker ps',
    submitResult: { submissionId: 1, isCorrect: false, results: [] },
  },
};

/** そもそも送れなかったとき。 */
export const 送れなかった: Story = {
  args: {
    ...baseArgs,
    starterCode: 'docker ps -a',
    submitError: '送信に失敗しました。通信を確認してください。',
  },
};

/** 実際に打って送るところ。 */
export const 打って送る: Story = {
  args: baseArgs,
  render: (args) => {
    function Interactive() {
      const [code, setCode] = useState('');
      return <QaExerciseView {...args} starterCode={code} onCodeChange={setCode} />;
    }
    return <Interactive />;
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement);
    const input = canvas.getByRole('textbox');
    await userEvent.type(input, 'docker ps -a');
    await expect(input).toHaveValue('docker ps -a');
    await userEvent.keyboard('{Enter}');
    await expect(args.onSubmit).toHaveBeenCalled();
  },
};
