import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, waitFor, within } from 'storybook/test';
import FirstTimeWelcome from './FirstTimeWelcome';

/**
 * はじめて来た人に「何ができるか・まず何をするか」を示すカード。
 *
 * 手順は 3〜5 個まで。それ以上並べると読まれない。
 *
 * この見本では `storageKey` を**わざと渡していない**。渡すと一度閉じたきり
 * 見本が空っぽになるため。
 */
const meta = {
  title: 'shared/FirstTimeWelcome',
  component: FirstTimeWelcome,
  parameters: { layout: 'padded' },
  args: {
    steps: [
      { title: '言語を選ぶ', description: '学びたい言語を 1 つ選びます。' },
      { title: '演習を解く', description: '画面の中でコードを書いて動かします。' },
      { title: '結果を振り返る', description: '通らなかったところの理由を読みます。' },
    ],
  },
  decorators: [
    (Story) => (
      <div className="max-w-2xl">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof FirstTimeWelcome>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 既定。題名と 3 つの手順。 */
export const 既定: Story = {
  args: {},
  play: async ({ canvasElement }) => {
    // このカードは 0.2 秒かけて浮かび上がる。出きる前に見ると opacity が 0 のままなので、
    // 「見えている」の判定は待ってから行う。
    await waitFor(async () => {
      await expect(
        within(canvasElement).getByRole('heading', { name: 'ようこそ FreStyle へ' }),
      ).toBeVisible();
    });
  },
};

/** 次の一歩を置いた形。 */
export const 最初の一歩つき: Story = {
  args: { primaryActionLabel: 'はじめて練習する', onPrimaryAction: fn() },
  play: async ({ args, canvasElement }) => {
    await userEvent.click(
      within(canvasElement).getByRole('button', { name: 'はじめて練習する' }),
    );
    await expect(args.onPrimaryAction).toHaveBeenCalledTimes(1);
  },
};

/** 題名を変えたところ。 */
export const 題名を変える: Story = {
  args: { title: 'ナレッジへようこそ' },
};

/** ✕ を押すと消える。 */
export const 閉じたとき: Story = {
  args: {},
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByRole('button', { name: 'このカードを閉じる' }));
    await waitFor(async () => {
      await expect(canvas.queryByRole('heading', { name: 'ようこそ FreStyle へ' })).toBeNull();
    });
  },
};
