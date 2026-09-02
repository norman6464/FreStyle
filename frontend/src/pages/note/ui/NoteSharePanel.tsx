import { useState } from 'react';
import { XMarkIcon } from '@heroicons/react/24/outline';
import type { NoteGrantablePrincipal, NoteGrantRole } from '@/entities/note';
import type { NoteShareRow } from '../model/useNoteShare';

/** 役割の選択肢（強い順）。値は backend の domain.GrantRole と同じ。 */
const ROLES: ReadonlyArray<{ value: NoteGrantRole; label: string }> = [
  { value: 'admin', label: '管理' },
  { value: 'editor', label: '編集' },
  { value: 'commenter', label: 'コメント' },
  { value: 'viewer', label: '閲覧' },
];

/**
 * 上の段から届いている人はここに出ない、と画面に書く一文。
 *
 * **これは飾りではない。** 一覧に出るのはこのページ自身に張った行だけなので、
 * 何も書かないと「このページを見られる人の一覧」に読める。実際はスペースの editor 全員が
 * 編集できるページでも、この一覧は空になり得る。空を「誰も見られない」と取り違えたまま
 * 機密を書き込む事故を、文言 1 つで塞ぐ。
 */
const INHERITED_NOTE =
  '上の段（ワークスペース・スペース・親ページ）から届いている人はここには出ません。';

export interface NoteSharePanelProps {
  /** いま開いているページの題名（どのページを共有しているかの手がかり）。 */
  pageTitle: string;
  /** このページ自身に張った権限。上の段から届いている相手は含まない。 */
  rows: NoteShareRow[];
  /** まだ権限を張っていない相手（追加の候補）。 */
  candidates: NoteGrantablePrincipal[];
  loading: boolean;
  /** 失敗の理由。null なら失敗していない。 */
  error: string | null;
  /** 書き込みが飛んでいる間 true（二重送信を止める）。 */
  saving: boolean;
  onGrant: (principalId: string, role: NoteGrantRole) => void | Promise<void>;
  onRevoke: (principalId: string) => void | Promise<void>;
  onClose: () => void;
}

/**
 * NoteSharePanel はページ単位の権限（既定の 3 段目）を見せて操作する。
 *
 * 設計の記録: https://claude.ai/code/artifact/7a173249-210b-4042-8bc4-d24ccacd303c
 *
 * 状態は受け取るだけで、自分では取りに行かない（取得は useNoteShare が持つ）。
 * こうしておくと、読み込み中・空・失敗の見た目を story でそのまま並べられる。
 *
 * 出すかどうかは呼び出し側が canManage を見て決める。権限が無い相手に押せるボタンを
 * 出しても、返るのは 404 だけで「権限が無い」ことすら伝わらない。
 */
export default function NoteSharePanel({
  pageTitle,
  rows,
  candidates,
  loading,
  error,
  saving,
  onGrant,
  onRevoke,
  onClose,
}: NoteSharePanelProps) {
  const [pickedPrincipal, setPickedPrincipal] = useState('');
  const [pickedRole, setPickedRole] = useState<NoteGrantRole>('editor');

  const handleAdd = async () => {
    if (!pickedPrincipal) return;
    await onGrant(pickedPrincipal, pickedRole);
    // 追加した相手は候補から外れるので、選択も戻す
    // （残すと次の追加で、もう候補に無い相手を指したままになる）。
    setPickedPrincipal('');
  };

  return (
    <section
      aria-label="共有"
      className="w-full max-w-md rounded-lg border border-surface-3 bg-surface-1 shadow-inkwell-8"
    >
      <header className="flex items-center gap-3 border-b border-surface-3 px-4 py-3">
        <h2 className="text-sm font-bold text-[var(--color-text-primary)]">共有</h2>
        <span className="min-w-0 flex-1 truncate text-right text-xs text-[var(--color-text-muted)]">
          {pageTitle}
        </span>
        <button
          type="button"
          onClick={onClose}
          aria-label="共有を閉じる"
          className="-mr-1 shrink-0 rounded p-1 text-[var(--color-text-muted)] transition-colors hover:bg-surface-2"
        >
          <XMarkIcon className="h-4 w-4" />
        </button>
      </header>

      <div className="flex flex-col gap-3 px-4 pb-4 pt-3">
        <div>
          <h3 className="text-[0.6875rem] font-bold tracking-wide text-[var(--color-text-muted)]">
            このページで足した権限
          </h3>
          {/*
            注記は行があるときだけ。空のときは下の一文が同じことをより強く言うので、
            両方出すと似た文が 2 つ並んで、どちらも読み飛ばされる。
          */}
          {!loading && !error && rows.length > 0 && (
            <p className="mt-0.5 text-xs leading-relaxed text-[var(--color-text-muted)]">
              {INHERITED_NOTE}
            </p>
          )}
        </div>

        {loading && (
          // 件数は分からないので 2 行に固定する（実際の件数に寄せると、
          // 読み込みのたびに高さが跳ねる）。
          <div className="flex flex-col gap-1.5" role="status" aria-label="権限を読み込み中">
            <div className="h-8 animate-skeleton rounded bg-surface-2" />
            <div className="h-8 animate-skeleton rounded bg-surface-2" />
          </div>
        )}

        {!loading && error && (
          <p role="alert" className="py-2 text-xs leading-relaxed text-red-600">
            {error}
          </p>
        )}

        {!loading && !error && rows.length === 0 && (
          <p className="mt-0.5 text-xs leading-relaxed text-[var(--color-text-muted)]">
            このページではまだ誰にも権限を足していません。上の段（ワークスペース・スペース・親ページ）から届いている人は、ここが空でもこのページを見られます。
          </p>
        )}

        {!loading && !error && rows.length > 0 && (
          <ul className="flex flex-col gap-0.5">
            {rows.map((row) => (
              <ShareRow
                key={row.principalId}
                row={row}
                disabled={saving}
                onChangeRole={(role) => onGrant(row.principalId, role)}
                onRemove={() => onRevoke(row.principalId)}
              />
            ))}
          </ul>
        )}

        <div className="h-px bg-surface-3" />

        <div>
          <h3 className="text-[0.6875rem] font-bold tracking-wide text-[var(--color-text-muted)]">
            相手を足す
          </h3>
          <div className="mt-2 flex gap-1.5">
            <select
              aria-label="足す相手"
              value={pickedPrincipal}
              onChange={(e) => setPickedPrincipal(e.target.value)}
              disabled={saving || candidates.length === 0}
              className="min-w-0 flex-1 rounded border border-surface-3 bg-surface-1 px-2 py-1.5 text-sm text-[var(--color-text-secondary)]"
            >
              <option value="">相手を選ぶ…</option>
              {candidates.map((c) => (
                <option key={c.id} value={c.id}>
                  {displayName(c.name, c.id)}
                </option>
              ))}
            </select>
            <select
              aria-label="与える役割"
              value={pickedRole}
              onChange={(e) => setPickedRole(e.target.value as NoteGrantRole)}
              disabled={saving}
              className="rounded border border-surface-3 bg-surface-1 px-2 py-1.5 text-sm text-[var(--color-text-secondary)]"
            >
              {ROLES.map((r) => (
                <option key={r.value} value={r.value}>
                  {r.label}
                </option>
              ))}
            </select>
            <button
              type="button"
              onClick={handleAdd}
              disabled={saving || !pickedPrincipal}
              className="shrink-0 rounded bg-brand-600 px-3 py-1.5 text-sm font-medium text-white transition-colors hover:bg-brand-700 disabled:opacity-45"
            >
              追加
            </button>
          </div>
        </div>
      </div>
    </section>
  );
}

interface ShareRowProps {
  row: NoteShareRow;
  disabled: boolean;
  onChangeRole: (role: NoteGrantRole) => void;
  onRemove: () => void;
}

function ShareRow({ row, disabled, onChangeRole, onRemove }: ShareRowProps) {
  const name = displayName(row.name, row.principalId);
  return (
    <li className="flex items-center gap-2 rounded px-1 py-1.5 hover:bg-surface-2">
      <span
        aria-hidden="true"
        className="grid h-7 w-7 shrink-0 place-items-center rounded-full border border-surface-3 bg-surface-2 text-[0.6875rem] font-bold text-[var(--color-text-tertiary)]"
      >
        {initials(row)}
      </span>
      <span className="min-w-0 flex-1">
        {/* 名前が引けなかった相手は ID が出る。切り詰められるので全体を title で補う。 */}
        <span
          title={name}
          className="block truncate text-sm font-medium text-[var(--color-text-primary)]"
        >
          {name}
        </span>
        <span className="block text-[0.6875rem] text-[var(--color-text-muted)]">
          {KIND_LABEL[row.kind]}
        </span>
      </span>
      <select
        aria-label={`${name} の役割`}
        value={row.role}
        onChange={(e) => onChangeRole(e.target.value as NoteGrantRole)}
        disabled={disabled}
        className="shrink-0 rounded border border-surface-3 bg-surface-1 px-1.5 py-1 text-sm text-[var(--color-text-secondary)]"
      >
        {ROLES.map((r) => (
          <option key={r.value} value={r.value}>
            {r.label}
          </option>
        ))}
      </select>
      <button
        type="button"
        onClick={onRemove}
        disabled={disabled}
        aria-label={`${name} を外す`}
        className="shrink-0 rounded p-1 text-[var(--color-text-muted)] transition-colors hover:bg-red-50 hover:text-red-600 disabled:opacity-45"
      >
        <XMarkIcon className="h-4 w-4" />
      </button>
    </li>
  );
}

const KIND_LABEL: Record<NoteShareRow['kind'], string> = {
  user: 'メンバー',
  group: 'グループ',
  space_all: 'スペースの全員',
  // 相手の一覧に居ない ID。引いた直後に主体が消えるとこうなる。行は残す
  // （消すと、取り消せない権限が画面から見えないまま残る）。
  unknown: '不明な相手',
};

/**
 * displayName は名前が空のとき ID を代わりに出す。
 *
 * backend は名前を引けなかった相手も空文字で返す（行を落とさない）。ここで
 * 「名前のない行」として出すと、どの権限を消せばよいのか人が選べない。
 */
function displayName(name: string, principalId: string): string {
  return name.trim() === '' ? principalId : name;
}

function initials(row: NoteShareRow): string {
  if (row.kind === 'space_all') return '全';
  if (row.kind === 'group') return 'G';
  const name = row.name.trim();
  return name === '' ? '?' : name.slice(0, 1);
}
