import type { Meta, StoryObj } from '@storybook/react-vite';
import LanguageIcon from './LanguageIcon';

/**
 * 言語のロゴ。
 *
 * ロゴは各社の公式マークなので自作せず、Devicon（MIT ライセンス）の絵を
 * `public/lang/<言語>.svg` に置いて使っている。
 *
 * 知らない言語や、絵が読めなかったときは汎用のコード記号に差し替わる。
 * 「絵が無いせいでカードが壊れて見える」状態を作らないため。
 *
 * なお、この絵は飾りなので読み上げソフトからは隠してある（隣に必ず言語名の文字がある）。
 */
const meta = {
  title: 'shared/LanguageIcon',
  component: LanguageIcon,
  parameters: { layout: 'centered' },
} satisfies Meta<typeof LanguageIcon>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 1 つだけ。 */
export const 既定: Story = {
  args: { language: 'go' },
};

/** 大きさは className で変える。 */
export const 大きさ: Story = {
  args: { language: 'go' },
  render: () => (
    <div className="flex items-end gap-4">
      <LanguageIcon language="go" className="h-6 w-6" />
      <LanguageIcon language="go" className="h-10 w-10" />
      <LanguageIcon language="go" className="h-16 w-16" />
    </div>
  ),
};

/** 知らない言語。汎用のコード記号に落ちる。 */
export const 未知の言語: Story = {
  args: { language: '存在しない言語' },
};

/** 一覧での見え方。 */
export const 並べたところ: Story = {
  args: { language: 'go' },
  render: () => (
    <div className="flex flex-wrap items-center gap-5">
      {['go', 'typescript', 'javascript', 'python', 'java', 'php', 'ruby', 'rust'].map((language) => (
        <div key={language} className="flex flex-col items-center gap-1">
          <LanguageIcon language={language} />
          <span className="text-xs text-[var(--color-text-muted)]">{language}</span>
        </div>
      ))}
    </div>
  ),
};
