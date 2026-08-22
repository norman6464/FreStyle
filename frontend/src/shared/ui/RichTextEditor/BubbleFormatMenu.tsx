import { BubbleMenu } from '@tiptap/react/menus';
import type { Editor } from '@tiptap/react';
import FormatMenuBar from './FormatMenuBar';

/**
 * BubbleFormatMenu はテキスト選択時に浮かぶ書式メニュー。
 * 位置決め（フローティング）だけを担い、中身は presentational な FormatMenuBar に委ねる。
 * 固定ツールバーを置かないインライン編集で、選択したときにだけ書式操作を出すための入れ物。
 */
export default function BubbleFormatMenu({ editor }: { editor: Editor }) {
  return (
    <BubbleMenu editor={editor} className="rte-bubble">
      <FormatMenuBar editor={editor} />
    </BubbleMenu>
  );
}
