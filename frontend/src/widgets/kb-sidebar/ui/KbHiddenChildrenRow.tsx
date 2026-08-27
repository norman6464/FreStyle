import { KB_INDENT_PX, KB_TOGGLE_WIDTH_PX } from './KbPageRow';

export interface KbHiddenChildrenRowProps {
  depth: number;
  count: number;
}

/**
 * KbHiddenChildrenRow は「この段に、自分には見えないページが n 枚ある」ことだけを示す行。
 *
 * **題名は出さない。** そもそも API が返してこない（返してもいけない）。
 * ただ消すと木に穴が空いた理由が分からず「壊れている」と読まれるので、居ることだけを示す。
 *
 * 押せる行ではないので treeitem にはしない。ツリーの項目として数えられると、
 * 上下移動で止まるのに何も起きない行になる。
 */
export default function KbHiddenChildrenRow({ depth, count }: KbHiddenChildrenRowProps) {
  return (
    <li role="none">
      <p
        className="py-0.5 text-xs text-[var(--color-text-muted)]"
        style={{ paddingLeft: depth * KB_INDENT_PX + KB_TOGGLE_WIDTH_PX }}
      >
        {count} ページは表示できません
      </p>
    </li>
  );
}
