import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';
import ShareRow from './ShareRow';
import type { ShareRow as ShareRowData } from '../model/types';

/**
 * 共有パネルの 1 行（誰に・どの役割で・外す）。
 *
 * 名前が引けなかった相手は **ID をそのまま出して行を残す**。行ごと消すと、その人に
 * 与えたままの権限が画面から見えなくなり、「誰が見られるのか」を人が説明できなくなる。
 *
 * 書き込み中は役割の選択も「外す」も止める。二重に送ると、取り消したはずの権限が
 * 戻ることがあるため。
 */
const meta = {
  title: 'features/permission-sharing/ShareRow',
  component: ShareRow,
  parameters: { layout: 'padded' },
  args: { disabled: false, onChangeRole: fn(), onRemove: fn() },
  decorators: [
    (Story) => (
      // 実物は <ul> の中に並ぶ。
      <ul className="w-96">
        <Story />
      </ul>
    ),
  ],
} satisfies Meta<typeof ShareRow>;

export default meta;
type Story = StoryObj<typeof meta>;

const row = (over: Partial<ShareRowData> = {}): ShareRowData => ({
  principalId: 'user:12',
  role: 'editor',
  name: '川野 拓馬',
  kind: 'user',
  ...over,
});

/** ふつうのメンバー。 */
export const メンバー: Story = {
  args: { row: row() },
  play: async ({ canvasElement }) => {
    await expect(
      within(canvasElement).getByRole('combobox', { name: '川野 拓馬 の役割' }),
    ).toHaveValue('editor');
  },
};

/** グループ。 */
export const グループ: Story = {
  args: { row: row({ principalId: 'group:3', name: 'バックエンド班', kind: 'group', role: 'viewer' }) },
};

/** スペースにいる全員。 */
export const スペースの全員: Story = {
  args: { row: row({ principalId: 'space:1', name: '営業定例', kind: 'space_all', role: 'commenter' }) },
};

/** 名前が引けなかった相手。ID を出して、行は残す。 */
export const 名前が引けない: Story = {
  args: { row: row({ principalId: 'user:999', name: '', kind: 'unknown' }) },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByText('user:999')).toBeVisible();
    await expect(canvas.getByText('不明な相手')).toBeVisible();
  },
};

/** 書き込み中。役割も「外す」も止まる。 */
export const 書き込み中: Story = {
  args: { row: row(), disabled: true },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByRole('combobox', { name: '川野 拓馬 の役割' })).toBeDisabled();
    await expect(canvas.getByRole('button', { name: '川野 拓馬 を外す' })).toBeDisabled();
  },
};

/** 役割を変えると、新しい役割が親へ渡る。 */
export const 役割を変える: Story = {
  args: { row: row() },
  play: async ({ args, canvasElement }) => {
    const select = within(canvasElement).getByRole('combobox', { name: '川野 拓馬 の役割' });
    await userEvent.selectOptions(select, 'viewer');
    await expect(args.onChangeRole).toHaveBeenCalledWith('viewer');
  },
};

/** 「外す」を押したとき。 */
export const 外す: Story = {
  args: { row: row() },
  play: async ({ args, canvasElement }) => {
    await userEvent.click(within(canvasElement).getByRole('button', { name: '川野 拓馬 を外す' }));
    await expect(args.onRemove).toHaveBeenCalledTimes(1);
  },
};

/** 一覧での見え方。名前は長ければ切り詰める。 */
export const 並べたところ: Story = {
  args: { row: row() },
  render: (args) => (
    <>
      <ShareRow {...args} row={row({ role: 'admin' })} />
      <ShareRow {...args} row={row({ principalId: 'user:13', name: '山田 太郎', role: 'editor' })} />
      <ShareRow
        {...args}
        row={row({
          principalId: 'group:3',
          name: 'とても長いグループ名がここに入って切り詰められる例',
          kind: 'group',
          role: 'viewer',
        })}
      />
      <ShareRow {...args} row={row({ principalId: 'user:999', name: '', kind: 'unknown' })} />
    </>
  ),
};
