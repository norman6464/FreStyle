import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import SubmissionRow from './SubmissionRow';
import type { ExerciseSubmission } from '../model/types';

/**
 * 提出の履歴 1 件ぶんの行。日時と合否だけを出す。
 *
 * 直近の何件かを並べて「さっきは通らなかったが今は通る」を見せるためのもの。
 * 中身のコードまでは出さない — 一覧が縦に伸びて、履歴として読めなくなる。
 */
const meta = {
  title: 'entities/exercise/SubmissionRow',
  component: SubmissionRow,
  parameters: { layout: 'padded' },
  decorators: [
    (Story) => (
      // 実物は <ul> の中に並ぶ。li を裸で置くと入れ子の決まりから外れる。
      <ul className="w-96 space-y-1">
        <Story />
      </ul>
    ),
  ],
} satisfies Meta<typeof SubmissionRow>;

export default meta;
type Story = StoryObj<typeof meta>;

const submission = (over: Partial<ExerciseSubmission> = {}): ExerciseSubmission => ({
  id: 1,
  userId: 1,
  exerciseKind: 'master',
  exerciseId: 10,
  submittedCode: 'print(1)',
  stdout: '1\n',
  stderr: '',
  exitCode: 0,
  isCorrect: true,
  submittedAt: '2026-09-06T09:41:00Z',
  ...over,
});

/** 通ったとき。 */
export const 合格: Story = {
  args: { submission: submission() },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText('合格')).toBeVisible();
  },
};

/** 通らなかったとき。 */
export const 不合格: Story = {
  args: { submission: submission({ id: 2, isCorrect: false, exitCode: 1 }) },
};

/** 何件か並べたところ（新しい順に積む想定）。 */
export const 履歴として並べる: Story = {
  args: { submission: submission() },
  render: () => (
    <>
      <SubmissionRow submission={submission({ id: 3, submittedAt: '2026-09-06T09:41:00Z' })} />
      <SubmissionRow
        submission={submission({ id: 2, isCorrect: false, submittedAt: '2026-09-06T09:12:00Z' })}
      />
      <SubmissionRow
        submission={submission({ id: 1, isCorrect: false, submittedAt: '2026-09-05T22:03:00Z' })}
      />
    </>
  ),
};
