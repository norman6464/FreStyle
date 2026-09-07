import { BubbleMenu } from '@tiptap/react/menus';
import type { Editor } from '@tiptap/react';
import FormatMenuBar from './FormatMenuBar';
import type { CommentAnchor } from './commentAnchor';

export interface BubbleFormatMenuProps {
  editor: Editor;
  /** 選択範囲からコメントを作りたいときに呼ばれる。渡さなければ「コメント」ボタンを出さない。 */
  onRequestComment?: (anchor: CommentAnchor) => void;
}

/**
 * BubbleFormatMenu はテキスト選択時に浮かぶ書式メニュー。
 * 位置決め（フローティング）だけを担い、中身は presentational な FormatMenuBar に委ねる。
 * 固定ツールバーを置かないインライン編集で、選択したときにだけ書式操作を出すための入れ物。
 */
export default function BubbleFormatMenu({ editor, onRequestComment }: BubbleFormatMenuProps) {
  return (
    <BubbleMenu editor={editor} className="rte-bubble">
      <FormatMenuBar editor={editor} onRequestComment={onRequestComment} />
    </BubbleMenu>
  );
}
