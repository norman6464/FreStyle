import { memo, useRef, useState } from 'react';
import MessageActionRow from './message/MessageActionRow';
import MessageAttachmentList, {
  MessageAttachmentView as _MessageAttachmentView,
} from './message/MessageAttachmentList';
import MarkdownView from './message/MarkdownView';
import { useSmoothReveal } from '../hooks/useSmoothReveal';
import { useFadeOnVisible } from '../hooks/useFadeOnVisible';

// `MessageAttachmentView` は MessageBubbleAi 経由で外部から import されている。
// 互換性のため同名で re-export する。
export type MessageAttachmentView = _MessageAttachmentView;

interface MessageBubbleProps {
  isSender: boolean;
  type?: 'text' | 'image' | 'bot';
  content: string;
  id: string;
  senderName?: string;
  createdAt?: string;
  attachments?: MessageAttachmentView[];
  onDelete?: ((id: string) => void) | null;
  onCopy?: ((id: string, content: string) => void) | null;
  isCopied?: boolean;
  isDeleted?: boolean;
  /** 旧 API 互換のため受けるが本コンポーネントでは使わない */
  onRephrase?: ((content: string) => void) | null;
}

/**
 * メッセージ表示コンポーネント。
 *
 * 設計:
 *   - 自分のメッセージは右寄せのコンパクトな塊（軽いハイライト背景）
 *   - アシスタント / 他者のメッセージは左寄せで **背景なしのフラット表示**。
 *     見出し / リスト / コード / 表など Markdown 要素を本文として描画する
 *   - カードや角丸の装飾はあえて入れない（一般的な汎用 AI チャット UI に寄せる）
 *
 * Markdown:
 *   - GFM（表 / タスクリスト / strikethrough）対応
 *   - コードブロックは highlight.js でシンタックスハイライト
 *
 * 4 つの分岐:
 *   1. type='image' → 画像のみ
 *   2. isDeleted → 削除済み プレース ホルダー
 *   3. isSender → 自分の メッセージ (右 寄せ + 添付 + アクション)
 *   4. その他 → アシスタント / 他者 (左 寄せ + Markdown + 考え 中 / 完了 マーカー)
 */
export default memo(function MessageBubble({
  isSender,
  type = 'text',
  content,
  id,
  senderName,
  createdAt,
  attachments,
  onDelete,
  onCopy,
  isCopied = false,
  isDeleted = false,
}: MessageBubbleProps) {
  const [showActions, setShowActions] = useState(false);

  // ストリーミング中(placeholder)判定。 useAskAi が done まで id を `streaming-…` に保つので
  // それを流用する。 数値 id(履歴/テスト)でも落ちないよう String 化。
  const isStreaming = String(id).startsWith('streaming-');
  // ストリーミング本文を Gemini 実物と同じリズム(句読点チャンク + 適応間隔)で放出する(FRESTYLE-146)。
  // アシスタントのテキスト応答のみ有効。それ以外(自分の発話・画像・履歴)は全文即時表示。
  const { text: shown, settled } = useSmoothReveal(
    content,
    isStreaming && !isSender && type === 'text' && !isDeleted,
  );

  // 画面外で mount したチャンクのフェードは、スクロールで見えた瞬間まで保留する(FRESTYLE-153)。
  // 自動スクロール追従をやめた(FRESTYLE-149)ため、これが無いと画面外でフェードを再生し終えてしまい、
  // 自分でスクロールして読むとアニメーションが一切見えない。
  const bodyRef = useRef<HTMLDivElement | null>(null);
  useFadeOnVisible(bodyRef, isStreaming || !settled, shown);

  if (type === 'image') {
    return (
      <div
        className={`my-3 flex ${isSender ? 'justify-end' : 'justify-start'}`}
        role="article"
        aria-label="画像メッセージ"
      >
        <img src={content} alt="画像" className="max-w-[85%] rounded-lg shadow-md" />
      </div>
    );
  }

  if (isDeleted) {
    return (
      <div
        className={`my-3 text-xs text-[var(--color-text-muted)] italic ${
          isSender ? 'text-right' : 'text-left'
        }`}
      >
        メッセージを削除しました
      </div>
    );
  }

  if (isSender) {
    return (
      <div
        className="my-4 group flex justify-end"
        onMouseEnter={() => setShowActions(true)}
        onMouseLeave={() => setShowActions(false)}
        role="article"
        aria-label="自分のメッセージ"
      >
        <div className="max-w-[85%] flex flex-col items-end gap-1">
          {attachments && attachments.length > 0 && (
            <MessageAttachmentList attachments={attachments} />
          )}
          {content && (
            <div className="px-4 py-2 rounded-2xl bg-[var(--color-surface-3)] text-[var(--color-text-primary)] text-sm whitespace-pre-wrap break-words">
              {content}
            </div>
          )}
          <MessageActionRow
            isSender
            id={id}
            content={content}
            createdAt={createdAt}
            isCopied={isCopied}
            onCopy={onCopy}
            onDelete={onDelete}
            visible={showActions}
          />
        </div>
      </div>
    );
  }

  // アシスタント / 他者のメッセージ: フラット、本文に Markdown。
  // 本文が空のときは SSE で最初の token が来るまでの「考え中」状態とみなし、
  // favicon をパルスさせながら "考え中..." ラベルを出す（Claude / ChatGPT 同様の UX）。
  // 最初の token が来た瞬間に上部のアバターアイコンは消し、本文と末尾の bookend マーカーで
  // 「処理中 → 応答中 → 完了」の状態遷移を視覚化する。
  // 「考え中...」はペーシング後の表示テキスト(shown)基準で判定する。 最初のチャンクが
  // 放出されるまでインジケータを保ち、 空本文のバブルがちらつかないようにする。
  const isThinking = type === 'text' && shown.trim() === '';
  return (
    <div
      className="my-6 group flex gap-3"
      onMouseEnter={() => setShowActions(true)}
      onMouseLeave={() => setShowActions(false)}
      role="article"
      aria-label={senderName ? `${senderName}のメッセージ` : 'AIのメッセージ'}
    >
      {isThinking ? (
        <img
          src="/favicon.svg"
          alt=""
          aria-hidden="true"
          className="w-5 h-5 flex-shrink-0 animate-thinking"
        />
      ) : (
        // 応答開始後はアバター枠を持たず本文を左端から始める。
        // ただし flex のレイアウトが崩れないよう同サイズの空きを確保する。
        <span className="w-5 h-5 flex-shrink-0" aria-hidden="true" />
      )}
      <div className="flex-1 min-w-0">
        {senderName && (
          <p className="text-xs text-[var(--color-text-muted)] mb-1">{senderName}</p>
        )}
        {isThinking ? (
          <p
            className="text-sm text-[var(--color-text-muted)] italic"
            aria-live="polite"
          >
            考え中...
          </p>
        ) : (
          <>
            <div
              ref={bodyRef}
              className="prose prose-sm max-w-none text-[var(--color-text-primary)] leading-relaxed"
            >
              {type === 'bot' ? (
                <p className="italic opacity-80">{content}</p>
              ) : (
                // done 直後の drain 中(!settled)もフェードプラグインを維持し、 残りチャンクに
                // フェードを効かせる。 出し切ったら素の Markdown(span なし)に戻る。
                <MarkdownView content={shown} isStreaming={isStreaming || !settled} />
              )}
            </div>
            <MessageActionRow
              isSender={false}
              id={id}
              content={content}
              createdAt={createdAt}
              isCopied={isCopied}
              onCopy={onCopy}
              visible={showActions}
            />
            {/* 完了マーカー: ストリーミング placeholder ではない（= done 確定済 / 履歴ロード済）
                AI 応答の末尾に favicon を 1 つ置いて「ここで応答が締まった」ことを視覚化する。
                Claude 等のメジャー AI チャットで採用されている bookend 表現に倣う。
                streaming 中(placeholder)は出さず、 done 後に残りを流し切って(settled)から出す。 */}
            {!isStreaming && settled && (
              <div className="mt-8 flex" aria-hidden="true">
                <img
                  src="/favicon.svg"
                  alt=""
                  className="w-7 h-7"
                />
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
});
