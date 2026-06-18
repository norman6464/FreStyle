import { useEffect, useRef, useState, useCallback, useMemo } from 'react';
import MessageBubbleAi from '../components/MessageBubbleAi';
import MessageInput from '../components/MessageInput';
import ConfirmModal from '../components/ConfirmModal';
import SecondaryPanel from '../components/layout/SecondaryPanel';
import AiSessionListItem from '../components/AiSessionListItem';
import Loading from '../components/Loading';
import {
  PlusIcon,
  Bars3Icon,
  MagnifyingGlassIcon,
  ArrowDownIcon,
  ChevronDownIcon,
  PencilIcon,
  TrashIcon,
} from '@heroicons/react/24/outline';
import { useAskAi } from '../hooks/useAskAi';
import { useMobilePanelState } from '../hooks/useMobilePanelState';
import { useCopyToClipboard } from '../hooks/useCopyToClipboard';

// 自動スクロールの「下端付近」と判定する閾値（px）。これ以上スクロールアップしたら
// auto-scroll を停止し、ユーザーが過去メッセージを読みやすくする。
const STICK_TO_BOTTOM_THRESHOLD = 80;

/**
 * 汎用 AI チャット画面。
 *
 * 旧版で並んでいた練習モード / スコアカード / シナリオ受け渡し / セッションノート /
 * 言い換え提案などはすべて削除し、純粋な「セッション一覧 + メッセージ表示 + 入力」だけを残す。
 *
 * Markdown 表示・カード装飾の除去・SSE ストリーミングは PR-B / PR-C で追加。
 */
export default function AskAiPage() {
  const { isOpen: mobilePanelOpen, open: openMobilePanel, close: closeMobilePanel } =
    useMobilePanelState();
  const { copiedId, copyToClipboard } = useCopyToClipboard();
  const {
    sessions,
    filteredSessions,
    messages,
    loading,
    currentSessionId,
    deleteModal,
    editingSessionId,
    editingTitle,
    setEditingTitle,
    sessionSearchQuery,
    setSessionSearchQuery,
    handleNewSession,
    handleSelectSession,
    handleDeleteSession,
    confirmDeleteSession,
    cancelDeleteSession,
    handleStartEditTitle,
    handleSaveTitle,
    handleCancelEditTitle,
    handleSend,
  } = useAskAi();

  // 自動スクロール制御:
  //   stickToBottom=true → messages 変化のたびに最下部へスクロール（streaming 中の追従）
  //   stickToBottom=false → ユーザーが上にスクロールしたので auto-scroll を停止
  // ユーザーが手動で底まで戻したら stickToBottom を再 true に戻す（onScroll で検知）。
  // ChatGPT / Claude.ai と同じ UX。
  const containerRef = useRef<HTMLDivElement | null>(null);
  const messagesEndRef = useRef<HTMLDivElement | null>(null);
  const [stickToBottom, setStickToBottom] = useState(true);
  // ref 版を 並行で持つ。 streaming 中は messages が ms 単位で更新されるため、
  // setStickToBottom (非同期) の反映を待たず ref を即時参照することで、
  // ユーザーが wheel した直後の "残り chunk" による 強制下スクロールを防ぐ。
  const stickToBottomRef = useRef(true);

  const updateStickToBottom = useCallback((next: boolean) => {
    stickToBottomRef.current = next;
    setStickToBottom(next);
  }, []);

  const handleContainerScroll = useCallback(() => {
    const el = containerRef.current;
    if (!el) return;
    const distanceFromBottom = el.scrollHeight - el.scrollTop - el.clientHeight;
    updateStickToBottom(distanceFromBottom < STICK_TO_BOTTOM_THRESHOLD);
  }, [updateStickToBottom]);

  // ユーザーが上へ wheel / touchmove したら、 即座に stickToBottom=false に切替えて
  // 進行中の smooth scroll を含めて 「追従」 を止める。 onScroll より早く拾える。
  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;
    const onWheel = (e: WheelEvent) => {
      if (e.deltaY < 0) {
        // 上方向へのスクロール → 追従停止
        updateStickToBottom(false);
      }
    };
    const onTouchMove = () => {
      // touchmove は方向判定が面倒なので、 一旦止める。 底に戻れば onScroll で再 true。
      const distanceFromBottom = el.scrollHeight - el.scrollTop - el.clientHeight;
      if (distanceFromBottom > STICK_TO_BOTTOM_THRESHOLD) {
        updateStickToBottom(false);
      }
    };
    el.addEventListener('wheel', onWheel, { passive: true });
    el.addEventListener('touchmove', onTouchMove, { passive: true });
    return () => {
      el.removeEventListener('wheel', onWheel);
      el.removeEventListener('touchmove', onTouchMove);
    };
  }, [updateStickToBottom]);

  // messages 変化時に stick している場合のみ最下部へ。stick していない時はそのまま
  // ユーザーが見ている位置を保つ。 ref で 同期チェックして race を回避。
  useEffect(() => {
    if (!stickToBottomRef.current) return;
    // smooth animation は wheel と競合しやすいので auto に落として 1 frame で完了させる
    messagesEndRef.current?.scrollIntoView({ behavior: 'auto' });
  }, [messages]);

  // セッション切替で先頭に戻ったときは強制的に底へ（履歴の最後を見せる）。
  useEffect(() => {
    updateStickToBottom(true);
  }, [currentSessionId, updateStickToBottom]);

  const jumpToBottom = useCallback(() => {
    updateStickToBottom(true);
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [updateStickToBottom]);

  // 左上に出すタイトル。 currentSessionId に該当する session の title が無ければ
  // 「新しいチャット」をフォールバックにする（ first message が確定するまでの間）。
  const currentSessionTitle = useMemo(() => {
    if (!currentSessionId) return '新しいチャット';
    const found = sessions.find((s) => s.id === currentSessionId);
    return found?.title?.trim() || '新しいチャット';
  }, [sessions, currentSessionId]);

  // タイトルクリックで開くアクションメニュー（名前変更 / 削除）。
  const [titleMenuOpen, setTitleMenuOpen] = useState(false);
  const titleMenuRef = useRef<HTMLDivElement | null>(null);

  // メニュー外クリックで閉じる。
  useEffect(() => {
    if (!titleMenuOpen) return;
    const handler = (e: MouseEvent) => {
      if (!titleMenuRef.current) return;
      if (!titleMenuRef.current.contains(e.target as Node)) {
        setTitleMenuOpen(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [titleMenuOpen]);

  const handleTitleRename = () => {
    setTitleMenuOpen(false);
    if (!currentSessionId) return;
    const found = sessions.find((s) => s.id === currentSessionId);
    handleStartEditTitle({ id: currentSessionId, title: found?.title });
    // モバイルでは編集 UI が左パネル内なので、 メニュー操作と同時にパネルを開いて確認しやすくする。
    openMobilePanel();
  };

  const handleTitleDelete = () => {
    setTitleMenuOpen(false);
    if (!currentSessionId) return;
    handleDeleteSession(currentSessionId);
  };

  if (loading && sessions.length === 0) {
    return <Loading message="読み込み中..." className="min-h-[calc(100vh-3.5rem)]" />;
  }

  return (
    <div className="flex h-full">
      {/* セカンダリパネル: セッション一覧 */}
      <SecondaryPanel
        title="セッション"
        badge={`${sessions.length}件`}
        mobileOpen={mobilePanelOpen}
        onMobileClose={closeMobilePanel}
        headerContent={
          <div className="space-y-2">
            <div className="relative">
              <MagnifyingGlassIcon className="w-4 h-4 absolute left-2.5 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)]" />
              <input
                type="text"
                placeholder="セッションを検索..."
                aria-label="セッションを検索"
                value={sessionSearchQuery}
                onChange={(e) => setSessionSearchQuery(e.target.value)}
                className="w-full pl-8 pr-3 py-1.5 bg-surface-2 border border-surface-3 rounded-lg text-sm text-[var(--color-text-primary)] placeholder:text-[var(--color-text-muted)] focus:outline-none focus:border-brand-400 transition-colors"
              />
            </div>
            <button
              onClick={handleNewSession}
              className="w-full bg-brand-500 text-white py-2 px-4 rounded-lg text-sm font-medium hover:bg-brand-600 transition-colors flex items-center justify-center gap-2"
            >
              <PlusIcon className="w-4 h-4" />
              新しいチャット
            </button>
          </div>
        }
      >
        <div className="p-2 space-y-0.5">
          {filteredSessions.map((session) => (
            <AiSessionListItem
              key={session.id}
              id={session.id}
              title={session.title}
              createdAt={session.createdAt}
              isActive={currentSessionId === session.id}
              isEditing={editingSessionId === session.id}
              editingTitle={editingTitle}
              onSelect={(id: number) => {
                handleSelectSession(id);
                closeMobilePanel();
              }}
              onStartEdit={handleStartEditTitle}
              onDelete={handleDeleteSession}
              onSaveTitle={handleSaveTitle}
              onCancelEdit={handleCancelEditTitle}
              onEditingTitleChange={setEditingTitle}
            />
          ))}
          {filteredSessions.length === 0 && (
            <p className="text-center text-xs text-[var(--color-text-muted)] py-8">
              セッションがありません
            </p>
          )}
        </div>
      </SecondaryPanel>

      {/* メイン: チャット */}
      <div className="flex-1 flex flex-col min-w-0 relative">
        {/* 上部ヘッダー: 区切り線は出さず、 タイトルはピル型ボタンで「名前を変更 / 削除」を出すドロップダウンにする */}
        <div className="flex items-center gap-2 px-4 pt-3 pb-2">
          <button
            onClick={openMobilePanel}
            className="md:hidden p-1.5 rounded hover:bg-[var(--color-surface-2)]"
            aria-label="セッションを開く"
          >
            <Bars3Icon className="w-5 h-5" />
          </button>
          <div className="relative" ref={titleMenuRef}>
            <button
              type="button"
              onClick={() => {
                if (!currentSessionId) return;
                setTitleMenuOpen((v) => !v);
              }}
              disabled={!currentSessionId}
              aria-haspopup="menu"
              aria-expanded={titleMenuOpen}
              className="flex items-center gap-1 px-2.5 py-1 rounded-md text-sm font-medium text-[var(--color-text-primary)] hover:bg-[var(--color-surface-2)] disabled:cursor-default disabled:hover:bg-transparent transition-colors max-w-[60vw]"
            >
              <span className="truncate">{currentSessionTitle}</span>
              {currentSessionId && (
                <ChevronDownIcon
                  className={`w-3.5 h-3.5 flex-shrink-0 text-[var(--color-text-muted)] transition-transform ${
                    titleMenuOpen ? 'rotate-180' : ''
                  }`}
                  aria-hidden="true"
                />
              )}
            </button>
            {titleMenuOpen && currentSessionId && (
              <div
                role="menu"
                className="absolute left-0 top-full mt-1 z-30 min-w-[180px] bg-[var(--color-surface-1)] border border-[var(--color-surface-3)] rounded-lg shadow-lg py-1"
              >
                <button
                  type="button"
                  role="menuitem"
                  onClick={handleTitleRename}
                  className="w-full flex items-center gap-2 px-3 py-2 text-sm text-[var(--color-text-primary)] hover:bg-[var(--color-surface-2)] transition-colors"
                >
                  <PencilIcon className="w-4 h-4 text-[var(--color-text-muted)]" />
                  名前を変更
                </button>
                <div className="my-1 border-t border-[var(--color-surface-3)]" />
                <button
                  type="button"
                  role="menuitem"
                  onClick={handleTitleDelete}
                  className="w-full flex items-center gap-2 px-3 py-2 text-sm text-red-600 hover:bg-red-500/10 transition-colors"
                >
                  <TrashIcon className="w-4 h-4" />
                  削除
                </button>
              </div>
            )}
          </div>
        </div>

        {/* スクロール領域上部のフェードオーバーレイ。 ヘッダー直下にスクロールしてきた本文が
            じわっと消えるように見える。 ヘッダーと本文の境界線を引かない代わりにここで視覚分離する。 */}
        <div
          aria-hidden="true"
          className="pointer-events-none absolute left-0 right-0 top-[44px] h-6 bg-gradient-to-b from-[var(--color-surface)] to-transparent z-10"
        />

        <div
          ref={containerRef}
          onScroll={handleContainerScroll}
          className="flex-1 overflow-y-auto px-4 py-6 relative"
        >
          {messages.length === 0 ? (
            <WelcomeGreeting />
          ) : (
            <div className="max-w-3xl mx-auto space-y-4">
              {messages.map((message) => (
                <MessageBubbleAi
                  key={message.id}
                  id={message.id}
                  type="text"
                  content={message.content}
                  attachments={message.attachments}
                  isSender={message.role === 'user'}
                  isCopied={copiedId === message.id}
                  onCopy={copyToClipboard}
                />
              ))}
              <div ref={messagesEndRef} />
            </div>
          )}
        </div>

        {/* stick が解除されているとき（= 上にスクロールしている）だけ表示する
             「最下部へ戻る」ボタン */}
        {!stickToBottom && messages.length > 0 && (
          <div className="absolute right-6 bottom-32 z-10">
            <button
              type="button"
              onClick={jumpToBottom}
              aria-label="最下部にスクロール"
              className="flex items-center justify-center w-10 h-10 rounded-full bg-[var(--color-surface-1)] text-[var(--color-text-primary)] shadow-md hover:bg-[var(--color-surface-2)] border border-[var(--color-surface-3)]"
            >
              <ArrowDownIcon className="w-5 h-5" />
            </button>
          </div>
        )}

        {/* 入力エリア: 角丸カード風 (主要 AI チャットの compose UI に倣う) */}
        <div className="px-4 pb-4 pt-2">
          <div className="max-w-3xl mx-auto">
            <MessageInput onSend={handleSend} />
          </div>
        </div>
      </div>

      {/* 削除確認モーダル */}
      <ConfirmModal
        isOpen={deleteModal.isOpen}
        title="セッションを削除しますか？"
        message="この操作は取り消せません。"
        confirmText="削除"
        cancelText="キャンセル"
        onConfirm={confirmDeleteSession}
        onCancel={cancelDeleteSession}
      />
    </div>
  );
}

/**
 * 新規セッション時にメッセージ欄が空のときに出すウェルカムグリーティング。
 *
 * デザイン:
 *   - 大きめのブランドアイコン (favicon.svg = 三角の翼ロゴ) を見出しの左に置き、
 *     「FreStyle 式の AI チャット」であることを最初の一画面で伝える。
 *   - 中央寄せ、本文は控えめに 1〜2 行。
 */
function WelcomeGreeting() {
  return (
    <div className="h-full flex items-center justify-center px-4">
      <div className="max-w-xl w-full text-center">
        <div className="flex items-center justify-center gap-3 mb-4">
          <img
            src="/favicon.svg"
            alt=""
            aria-hidden="true"
            className="w-10 h-10 flex-shrink-0"
          />
          <h2 className="text-2xl font-medium text-[var(--color-text-primary)]">
            なにをお手伝いしましょうか？
          </h2>
        </div>
        <p className="text-sm text-[var(--color-text-muted)]">
          質問・要約・コードレビューなど、自由にメッセージを送ってください。
        </p>
      </div>
    </div>
  );
}
