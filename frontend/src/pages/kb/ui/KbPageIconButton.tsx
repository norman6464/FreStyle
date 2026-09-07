import { useEffect, useRef, useState } from 'react';
import { FaceSmileIcon } from '@heroicons/react/24/outline';
import type { KbIcon } from '@/entities/kb';
import KbPageIconPicker from './KbPageIconPicker';

export interface KbPageIconButtonProps {
  /** 未設定は null（明示的に外した）と undefined（旧応答）のどちらもあり得る。 */
  icon?: KbIcon | null;
  canEdit: boolean;
  /**
   * 設定・変更・解除をまとめて担う（null が解除）。**失敗は投げてくる**前提
   * （呼び出し側がトーストで知らせ、ピッカーは開いたままになる）。
   */
  onChange: (icon: KbIcon | null) => Promise<void>;
}

/**
 * KbPageIconButton はページ見出しの左に置くアイコンの表示・変更口。
 *
 * 4 状態:
 *   読むだけ + 未設定  → 何も出さない（無いものを匂わせる要素を置かない）
 *   読むだけ + 設定済み → 絵文字を役割 img で出すだけ（押せない）
 *   書ける + 未設定    → 「アイコンを追加」。md 以上は触れているかフォーカスが
 *                        当たっているときだけ現れる（DOM には常に居る — 外すと
 *                        Tab の順序が触れるたびに変わる。KbRowActions と同じ理由）。
 *                        md 未満は常に見える（ホバーが無い環境のため）。
 *   書ける + 設定済み   → 絵文字そのものが aria-expanded なボタン
 *
 * ピッカーの開閉と外側クリック・Escape での消し方は KbRowActions と同じ形。
 */
export default function KbPageIconButton({ icon, canEdit, onChange }: KbPageIconButtonProps) {
  const [open, setOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const onDocumentMouseDown = (event: MouseEvent) => {
      if (
        containerRef.current &&
        event.target instanceof Node &&
        containerRef.current.contains(event.target)
      ) {
        return;
      }
      setOpen(false);
    };
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setOpen(false);
    };
    document.addEventListener('mousedown', onDocumentMouseDown);
    document.addEventListener('keydown', onKeyDown);
    return () => {
      document.removeEventListener('mousedown', onDocumentMouseDown);
      document.removeEventListener('keydown', onKeyDown);
    };
  }, [open]);

  if (!canEdit) {
    if (!icon) return null;
    return (
      <span role="img" aria-label="ページのアイコン" className="mb-1 block text-4xl leading-none">
        {icon.value}
      </span>
    );
  }

  const picker = open && (
    <div className="absolute left-0 top-full z-20 mt-1">
      <KbPageIconPicker
        current={icon ?? null}
        onSelect={(next) => onChange(next)}
        onClear={() => onChange(null)}
        onClose={() => setOpen(false)}
      />
    </div>
  );

  if (!icon) {
    return (
      <div ref={containerRef} className="relative mb-1 inline-block">
        <button
          type="button"
          onClick={() => setOpen((prev) => !prev)}
          aria-label="アイコンを追加"
          aria-expanded={open}
          className={`flex items-center gap-1 rounded px-1.5 py-1 text-sm text-[var(--color-text-muted)] hover:bg-surface-2 md:transition-opacity ${
            open ? 'md:opacity-100' : 'opacity-100 md:opacity-0 md:group-hover:opacity-100 md:focus-visible:opacity-100'
          }`}
        >
          <FaceSmileIcon className="h-4 w-4" aria-hidden="true" />
          アイコンを追加
        </button>
        {picker}
      </div>
    );
  }

  return (
    <div ref={containerRef} className="relative mb-1 inline-block">
      <button
        type="button"
        onClick={() => setOpen((prev) => !prev)}
        aria-label="ページのアイコンを変更"
        aria-expanded={open}
        className="rounded text-4xl leading-none hover:bg-surface-2"
      >
        <span data-icon="emoji" aria-hidden="true">
          {icon.value}
        </span>
      </button>
      {picker}
    </div>
  );
}
