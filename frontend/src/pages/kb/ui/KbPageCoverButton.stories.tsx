import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, waitFor, within } from 'storybook/test';
import KbPageCoverButton from './KbPageCoverButton';

/**
 * ページ頭部のカバー画像の設定・変更・解除口。4 状態を持つ
 * （読むだけ / 書ける × 未設定 / 設定済み）。画像そのものの表示は KbPageCover が別に担う
 * （この部品は操作だけ）。
 */
const meta = {
  title: 'pages/kb/KbPageCoverButton',
  component: KbPageCoverButton,
  parameters: { layout: 'centered' },
  args: {
    cover: null,
    canEdit: true,
    onChange: fn(async () => {}),
  },
} satisfies Meta<typeof KbPageCoverButton>;

export default meta;
type Story = StoryObj<typeof meta>;

const cover = { type: 'file' as const, url: 'https://cdn.example.com/cover.png' };

const pngFile = (name = 'cover.png', size = 1024) => {
  const file = new File([new Uint8Array(size)], name, { type: 'image/png' });
  return file;
};

function fileInputOf(canvasElement: HTMLElement): HTMLInputElement {
  const input = canvasElement.querySelector('input[type="file"]');
  if (!input) throw new Error('file input が見つかりません');
  return input as HTMLInputElement;
}

/** 読むだけ + 未設定。何も出さない。 */
export const 読むだけ_未設定: Story = {
  args: { canEdit: false, cover: null },
  play: async ({ canvasElement }) => {
    await expect(canvasElement.querySelector('button')).toBeNull();
  },
};

/** 読むだけ + 設定済み。画像は KbPageCover が別に出すので、ここでは何も出さない。 */
export const 読むだけ_設定済み: Story = {
  args: { canEdit: false, cover },
  play: async ({ canvasElement }) => {
    await expect(canvasElement.querySelector('button')).toBeNull();
  },
};

/** 書ける + 未設定。「カバー画像を追加」だけが出る。 */
export const 書ける_未設定: Story = {
  args: { canEdit: true, cover: null },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByRole('button', { name: 'カバー画像を追加' })).toBeInTheDocument();
    // 未設定なので「外す」は無い。
    await expect(canvas.queryByRole('button', { name: 'カバー画像を外す' })).toBeNull();
  },
};

/** 書ける + 設定済み。「変更」と「外す」の両方が出る。 */
export const 書ける_設定済み: Story = {
  args: { canEdit: true, cover },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByRole('button', { name: 'カバー画像を変更' })).toBeInTheDocument();
    await expect(canvas.getByRole('button', { name: 'カバー画像を外す' })).toBeInTheDocument();
  },
};

/** ファイルを選ぶとアップロードされて設定される（onChange に File が渡る）。 */
export const アップロードして設定: Story = {
  args: { canEdit: true, cover: null, onChange: fn(async () => {}) },
  play: async ({ canvasElement, args }) => {
    const file = pngFile();
    const input = fileInputOf(canvasElement);

    await userEvent.upload(input, file, { applyAccept: false });

    await waitFor(() => expect(args.onChange).toHaveBeenCalledWith(file));
  },
};

/** 大きすぎる画像は UI 側で早期に弾く（onChange は呼ばれない）。 */
export const 大きすぎる画像は弾く: Story = {
  args: { canEdit: true, cover: null, onChange: fn(async () => {}) },
  play: async ({ canvasElement, args }) => {
    // 10MiB を超えるダミーファイル。
    const tooLarge = pngFile('huge.png', 10 * 1024 * 1024 + 1);
    const input = fileInputOf(canvasElement);

    await userEvent.upload(input, tooLarge, { applyAccept: false });

    await waitFor(() => {
      expect(within(canvasElement).getByRole('alert')).toHaveTextContent('大きすぎます');
    });
    expect(args.onChange).not.toHaveBeenCalled();
  },
};

/** 対応していない形式も UI 側で弾く。 */
export const 非対応形式は弾く: Story = {
  args: { canEdit: true, cover: null, onChange: fn(async () => {}) },
  play: async ({ canvasElement, args }) => {
    const svg = new File(['<svg/>'], 'x.svg', { type: 'image/svg+xml' });
    const input = fileInputOf(canvasElement);

    await userEvent.upload(input, svg, { applyAccept: false });

    await waitFor(() => {
      expect(within(canvasElement).getByRole('alert')).toHaveTextContent('対応していない画像形式');
    });
    expect(args.onChange).not.toHaveBeenCalled();
  },
};

/** アップロード（設定）に失敗しても、ボタンは操作できる状態のまま残る。 */
export const 設定に失敗: Story = {
  args: {
    canEdit: true,
    cover: null,
    onChange: fn(async () => {
      throw new Error('upload failed');
    }),
  },
  play: async ({ canvasElement, args }) => {
    const file = pngFile();
    const input = fileInputOf(canvasElement);

    await userEvent.upload(input, file, { applyAccept: false });

    await waitFor(() => expect(args.onChange).toHaveBeenCalledWith(file));
    // 失敗しても操作可能なまま（disabled のまま固まらない）。
    await expect(
      within(canvasElement).getByRole('button', { name: 'カバー画像を追加' }),
    ).toBeEnabled();
  },
};

/** 「外す」を押すと onChange(null) が呼ばれる。 */
export const 外す: Story = {
  args: { canEdit: true, cover, onChange: fn(async () => {}) },
  play: async ({ canvasElement, args }) => {
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByRole('button', { name: 'カバー画像を外す' }));

    await waitFor(() => expect(args.onChange).toHaveBeenCalledWith(null));
  },
};

/**
 * 不正なファイルを選んで検証エラーを出した後に「外す」を押すと、そのエラーは消える
 * （外す自体はファイルの検証を経ないので、放置すると解除後も残ってしまう）。
 */
export const 不正なファイルの後に外すとエラーが消える: Story = {
  args: { canEdit: true, cover, onChange: fn(async () => {}) },
  play: async ({ canvasElement, args }) => {
    const canvas = within(canvasElement);
    const svg = new File(['<svg/>'], 'x.svg', { type: 'image/svg+xml' });
    const input = fileInputOf(canvasElement);

    await userEvent.upload(input, svg, { applyAccept: false });
    await waitFor(() => {
      expect(canvas.getByRole('alert')).toHaveTextContent('対応していない画像形式');
    });

    await userEvent.click(canvas.getByRole('button', { name: 'カバー画像を外す' }));
    await waitFor(() => expect(args.onChange).toHaveBeenCalledWith(null));
    await waitFor(() => expect(canvas.queryByRole('alert')).toBeNull());
  },
};
