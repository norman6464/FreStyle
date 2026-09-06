import { useState } from 'react';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, within } from 'storybook/test';
import type { MasterExercise } from '@/entities/exercise';
import { withRouter } from '../../../../.storybook/decorators';
import PreviewExerciseView from './PreviewExerciseView';

/**
 * HTML と CSS を書きながら、その場で見た目を確かめる形の演習。
 *
 * 実行して出力を比べる採点はしない。見本と見比べて、本人が「できた」と言ったら提出になる。
 * 見た目の正しさは文字列で比べられないので、機械の判定に無理をさせていない。
 *
 * 書いたものは `sandbox=""` の枠の中で描く。スクリプトも同一オリジンも許していないので、
 * 打ち込んだ HTML がこのアプリの権限で動くことはない。**この設定は緩めないこと。**
 */
const meta = {
  title: 'pages/exercise-detail/PreviewExerciseView',
  component: PreviewExerciseView,
  parameters: { layout: 'fullscreen' },
  decorators: [withRouter],
} satisfies Meta<typeof PreviewExerciseView>;

export default meta;
type Story = StoryObj<typeof meta>;

const exercise: MasterExercise = {
  id: 77,
  slug: 'html-card',
  language: 'html',
  orderIndex: 2,
  category: '見た目',
  title: 'カードを作る',
  description: '見本と同じ見た目のカードを作ってください。',
  starterCode: '<div class="card">\n  <h2>題名</h2>\n</div>',
  hintText: 'border-radius を使います。',
  expectedOutput:
    '<div style="border:1px solid #ddd;border-radius:12px;padding:16px"><h2>題名</h2><p>説明</p></div>',
  mode: 'preview',
  explanation: '',
  difficulty: 1,
  isPublished: true,
  createdAt: '2026-09-01T00:00:00Z',
  updatedAt: '2026-09-01T00:00:00Z',
};

const baseArgs = {
  exercise,
  code: exercise.starterCode,
  onCodeChange: fn(),
  submitting: false,
  submitResult: null,
  submitError: null,
  solved: false,
  onSubmit: fn(),
  onReset: fn(),
};

/** 書きはじめたところ。 */
export const 書きはじめ: Story = {
  args: baseArgs,
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('heading', { name: 'カードを作る' })).toBeVisible();
  },
};

/** 書き進めたところ。右の枠にその場で反映される。 */
export const 書き進めたところ: Story = {
  args: {
    ...baseArgs,
    code: '<div style="border:1px solid #ddd;border-radius:12px;padding:16px">\n  <h2>題名</h2>\n  <p>説明</p>\n</div>',
  },
};

/** 送っている最中。 */
export const 送信中: Story = {
  args: { ...baseArgs, submitting: true },
};

/** できたと申告したあと。 */
export const 提出済み: Story = {
  args: {
    ...baseArgs,
    solved: true,
    submitResult: { submissionId: 1, isCorrect: true, results: [] },
  },
};

/** 送れなかったとき。 */
export const 送れなかった: Story = {
  args: { ...baseArgs, submitError: '送信に失敗しました。通信を確認してください。' },
};

/** 実際に打って、枠の中が変わるところ。 */
export const 打ってみる: Story = {
  args: baseArgs,
  render: (args) => {
    function Interactive() {
      const [code, setCode] = useState(args.code);
      return <PreviewExerciseView {...args} code={code} onCodeChange={setCode} />;
    }
    return <Interactive />;
  },
};
