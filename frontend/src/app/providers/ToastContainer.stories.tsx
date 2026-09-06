import { useEffect } from 'react';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, waitFor, within } from 'storybook/test';
import ToastContainer from './ToastContainer';
import { ToastProvider } from './ToastProvider';
import { useToast } from '@/shared/lib/hooks/useToast';
import type { ToastType } from '@/shared/ui/Toast';

/**
 * 知らせ（トースト）を画面上部の中央に積む置き場。
 *
 * 置き場そのものは押せないようにしてある（`pointer-events-none`）。画面の上端に見えない板が
 * 敷かれていると、その下にあるものを押せなくなるため。知らせ 1 枚 1 枚だけが押せる。
 *
 * 1 枚も無いときは何も描かない。空の箱を残すと、そこだけ操作が効かない帯になる。
 *
 * 同時に出せるのは 3 枚まで。連続で操作しても画面が埋まらないよう、古いものから落ちる。
 */
const meta = {
  title: 'app/ToastContainer',
  component: ToastContainer,
  parameters: { layout: 'fullscreen' },
} satisfies Meta<typeof ToastContainer>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 見本の中身を、開いた直後に流し込む小さな部品。 */
function Emit({ items }: { items: Array<{ type: ToastType; message: string }> }) {
  const { showToast } = useToast();
  useEffect(() => {
    items.forEach((item) => showToast(item.type, item.message));
  }, [items, showToast]);
  return null;
}

function Stage({ items }: { items: Array<{ type: ToastType; message: string }> }) {
  return (
    <ToastProvider>
      <div className="min-h-[360px] bg-surface p-8">
        <p className="text-sm text-[var(--color-text-secondary)]">
          この後ろの文字は、知らせが出ていても押せる（置き場は押下を素通しする）。
        </p>
      </div>
      <Emit items={items} />
      <ToastContainer />
    </ToastProvider>
  );
}

/** 1 枚だけ。 */
export const 一枚: Story = {
  render: () => <Stage items={[{ type: 'success', message: 'ページを保存しました' }]} />,
  play: async ({ canvasElement }) => {
    await waitFor(async () => {
      await expect(within(canvasElement).getByRole('alert')).toHaveTextContent(
        'ページを保存しました',
      );
    });
  },
};

/** 種類の違うものが積み重なったところ。 */
export const 積み重なる: Story = {
  render: () => (
    <Stage
      items={[
        { type: 'success', message: 'ページを保存しました' },
        { type: 'info', message: '共有リンクをコピーしました' },
        { type: 'error', message: '保存できませんでした' },
      ]}
    />
  ),
  play: async ({ canvasElement }) => {
    await waitFor(async () => {
      await expect(within(canvasElement).getAllByRole('alert')).toHaveLength(3);
    });
  },
};

/** 上限（3 枚）を超えたとき。古いものから落ちる。 */
export const 上限を超えたとき: Story = {
  render: () => (
    <Stage
      items={[
        { type: 'info', message: '1 枚目（落ちる）' },
        { type: 'info', message: '2 枚目（落ちる）' },
        { type: 'info', message: '3 枚目' },
        { type: 'info', message: '4 枚目' },
        { type: 'info', message: '5 枚目' },
      ]}
    />
  ),
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await waitFor(async () => {
      await expect(canvas.getAllByRole('alert')).toHaveLength(3);
    });
    await expect(canvas.queryByText('1 枚目（落ちる）')).toBeNull();
    // 知らせは 0.6 秒かけて上から落ちてくる。落ちきる前に見ると透明のままなので待つ。
    await waitFor(async () => {
      await expect(canvas.getByText('5 枚目')).toBeVisible();
    });
  },
};

/** 同じ知らせを続けて出したとき。増やさず 1 枚にまとめる。 */
export const 同じ知らせはまとめる: Story = {
  render: () => (
    <Stage
      items={[
        { type: 'error', message: '保存できませんでした' },
        { type: 'error', message: '保存できませんでした' },
        { type: 'error', message: '保存できませんでした' },
      ]}
    />
  ),
  play: async ({ canvasElement }) => {
    await waitFor(async () => {
      await expect(within(canvasElement).getAllByRole('alert')).toHaveLength(1);
    });
  },
};

/** 1 枚も無いとき。何も描かない。 */
export const 何も無いとき: Story = {
  render: () => <Stage items={[]} />,
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).queryByRole('alert')).toBeNull();
  },
};
