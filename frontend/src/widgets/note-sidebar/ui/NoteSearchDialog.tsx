import { useEffect, useId, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { MagnifyingGlassIcon } from '@heroicons/react/24/outline';
import { NoteRepository, type NotePage, type NoteSpace } from '@/entities/note';
import { buildSearchView } from '../model/searchView';
import { NotePageIcon } from '@/shared/ui/icons/note';

export interface NoteSearchDialogProps {
  workspaceSlug: string;
  /** スペース名を引くための一覧（結果はスペースごとに見出しを付けて並べる）。 */
  spaces: NoteSpace[];
  onClose: () => void;
}


/**
 * NoteSearchDialog はワークスペース全体の題名検索のモーダル。
 *
 * サイドバー本体は場所（木）を示すことに徹し、検索は入口だけを置いてここで行う
 * （常設の入力欄が場所の面を圧迫し、木と結果が同じ狭い面で入れ替わる形をやめた）。
 * 検索はサーバーが行い、返るのは木と同じ規則で閲覧できる現役ページだけ。
 * 入力から 250ms 待って問い合わせ、世代番号で古い応答を捨てる。
 * ↑↓ で選び Enter で開く。Esc・外側クリックで閉じる。
 */
export default function NoteSearchDialog({ workspaceSlug, spaces, onClose }: NoteSearchDialogProps) {
  const navigate = useNavigate();
  const listboxId = useId();
  const [query, setQuery] = useState('');
  const [status, setStatus] = useState<'idle' | 'loading' | 'done' | 'error'>('idle');
  const [pages, setPages] = useState<NotePage[]>([]);
  const [selectedIndex, setSelectedIndex] = useState(0);
  // 速く打ったときに、古い応答が新しい結果を上書きしないための世代番号。
  const generation = useRef(0);
  // 再試行の引き金（値そのものに意味は無い。増えたら同じ問い合わせをもう一度投げる）。
  const [attempt, setAttempt] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  useEffect(() => {
    // 空入力に戻したときも世代を進める。進めないと、消す前に飛ばした検索の応答が
    // まだ有効な世代のまま届き、空の入力に古い結果が再表示される。
    const token = ++generation.current;
    const needle = query.trim();
    if (needle === '') {
      setStatus('idle');
      setPages([]);
      return undefined;
    }
    setStatus('loading');
    const timer = setTimeout(() => {
      NoteRepository.searchPages(workspaceSlug, needle)
        .then((found) => {
          if (token !== generation.current) return;
          setPages(found);
          setSelectedIndex(0);
          setStatus('done');
        })
        .catch(() => {
          if (token !== generation.current) return;
          setStatus('error');
        });
    }, 250);
    return () => clearTimeout(timer);
  }, [workspaceSlug, query, attempt]);

  const view = buildSearchView(pages, spaces);

  const open = (page: NotePage) => {
    navigate(`/p/${page.id}`);
    onClose();
  };

  const onKeyDown = (event: React.KeyboardEvent) => {
    // 日本語入力の変換キャンセル・確定はモーダルの操作にしない（打ちかけの検索語を守る）。
    if (event.nativeEvent.isComposing || event.keyCode === 229) return;
    if (event.key === 'Escape') {
      event.preventDefault();
      onClose();
      return;
    }
    if (view.flat.length === 0) return;
    if (event.key === 'ArrowDown') {
      event.preventDefault();
      setSelectedIndex((prev) => (prev + 1) % view.flat.length);
      return;
    }
    if (event.key === 'ArrowUp') {
      event.preventDefault();
      setSelectedIndex((prev) => (prev - 1 + view.flat.length) % view.flat.length);
      return;
    }
    if (event.key === 'Enter') {
      event.preventDefault();
      const selected = view.flat[selectedIndex];
      if (selected) open(selected);
    }
  };

  const optionId = (page: NotePage) => `${listboxId}-${page.id}`;
  const selectedPage = view.flat[selectedIndex];

  const renderRow = (page: NotePage) => (
    <li key={page.id} id={optionId(page)} role="option" aria-selected={page.id === selectedPage?.id}>
      <button
        type="button"
        tabIndex={-1}
        // mousedown での blur によりモーダルが振る舞いを変えないよう、click で開く。
        onClick={() => open(page)}
        className={`flex w-full min-w-0 items-center gap-1.5 rounded-md px-2 py-1.5 text-left text-sm ${
          page.id === selectedPage?.id
            ? 'bg-brand-500/10 font-medium text-brand-600'
            : 'text-[var(--color-text-primary)] hover:bg-surface-2'
        }`}
      >
        <NotePageIcon className="h-4 w-4 shrink-0 text-[var(--color-text-muted)]" />
        <span className="truncate">{page.title}</span>
      </button>
    </li>
  );

  return (
    <div className="fixed inset-0 z-[100] flex items-start justify-center pt-[18vh]">
      <div
        data-testid="note-search-overlay"
        className="absolute inset-0 bg-black/50"
        onClick={onClose}
      />
      <div
        role="dialog"
        aria-modal="true"
        aria-label="ページを検索"
        className="relative flex max-h-[60vh] w-full max-w-lg flex-col overflow-hidden rounded-xl border border-surface-3 bg-surface-1 shadow-2xl"
      >
        <div className="flex items-center gap-3 border-b border-surface-3 px-4 py-3">
          <MagnifyingGlassIcon
            className="h-5 w-5 shrink-0 text-[var(--color-text-muted)]"
            aria-hidden="true"
          />
          <input
            ref={inputRef}
            type="search"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            onKeyDown={onKeyDown}
            placeholder="題名で検索..."
            aria-label="ページを題名で検索"
            role="combobox"
            aria-expanded={view.flat.length > 0}
            aria-controls={listboxId}
            aria-activedescendant={selectedPage ? optionId(selectedPage) : undefined}
            className="flex-1 bg-transparent text-sm text-[var(--color-text-primary)] outline-none placeholder:text-[var(--color-text-muted)]"
          />
          <kbd className="hidden shrink-0 rounded border border-surface-3 bg-surface-2 px-1.5 py-0.5 font-mono text-[10px] text-[var(--color-text-muted)] sm:inline-flex">
            ESC
          </kbd>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto p-2">
          {status === 'idle' && (
            <p className="px-2 py-6 text-center text-xs text-[var(--color-text-muted)]">
              ワークスペース全体から題名で探します
            </p>
          )}
          {status === 'loading' && (
            <p className="px-2 py-6 text-center text-xs text-[var(--color-text-muted)]">検索中…</p>
          )}
          {status === 'error' && (
            <div className="px-2 py-6 text-center text-xs text-red-600">
              <p>検索に失敗しました</p>
              <button
                type="button"
                onClick={() => setAttempt((prev) => prev + 1)}
                className="mt-1 underline hover:no-underline"
              >
                再試行
              </button>
            </div>
          )}
          {status === 'done' && view.flat.length === 0 && (
            <p className="px-2 py-6 text-center text-xs text-[var(--color-text-muted)]">
              一致するページがありません
            </p>
          )}
          {status === 'done' && view.flat.length > 0 && (
            <ul id={listboxId} role="listbox" aria-label="検索結果">
              {view.groups.map(({ space, pages: groupPages }) => (
                <li key={space.id} role="group" aria-label={space.name}>
                  <h2 className="px-2 py-1 text-xs font-semibold uppercase tracking-wide text-[var(--color-text-muted)]">
                    {space.name}
                  </h2>
                  <ul role="presentation" className="space-y-px">
                    {groupPages.map(renderRow)}
                  </ul>
                </li>
              ))}
              {view.orphan.length > 0 && (
                // 見えるスペース一覧に無いスペースのページ。名前が引けないので見出しなし。
                <li role="group" aria-label="その他のページ">
                  <ul role="presentation" className="space-y-px">
                    {view.orphan.map(renderRow)}
                  </ul>
                </li>
              )}
            </ul>
          )}
        </div>
      </div>
    </div>
  );
}
