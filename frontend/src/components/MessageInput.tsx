import { useState, useRef, useEffect, KeyboardEvent, ChangeEvent } from 'react';
import { ArrowUpIcon, PlusIcon } from '@heroicons/react/24/solid';
import { XMarkIcon, PhotoIcon } from '@heroicons/react/24/outline';
import { useAutoResizeTextarea } from '../hooks/useAutoResizeTextarea';
import aiChatRepository from '../repositories/AiChatRepository';
import type { AiAttachment, AiAttachmentFormat } from '../types';

interface MessageInputProps {
  onSend: (text: string, attachments?: AiAttachment[]) => void;
  isSending?: boolean;
}

// 画像のみ受け付ける（PR-G1）。PR-G2 で application/pdf, text/csv を追加予定。
const ACCEPTED_FILE_TYPES = 'image/png,image/jpeg,image/gif,image/webp';

const ACCEPTED_CONTENT_TYPES = new Set([
  'image/png',
  'image/jpeg',
  'image/gif',
  'image/webp',
]);

const MAX_ATTACHMENTS = 4;
const MAX_BYTES_PER_FILE = 5 * 1024 * 1024; // Bedrock image upper bound

const FORMAT_FROM_CONTENT_TYPE: Record<string, AiAttachmentFormat> = {
  'image/png': 'png',
  'image/jpeg': 'jpeg',
  'image/gif': 'gif',
  'image/webp': 'webp',
};

interface PendingAttachment extends AiAttachment {
  /** アップロード進行中フラグ（送信ボタンの活性制御用） */
  uploading?: boolean;
  /** アップロード失敗時のメッセージ */
  error?: string;
}

export default function MessageInput({ onSend, isSending = false }: MessageInputProps) {
  const [text, setText] = useState('');
  const [pending, setPending] = useState<PendingAttachment[]>([]);
  const [validationError, setValidationError] = useState<string | null>(null);
  const [isComposing, setIsComposing] = useState(false);
  const textareaRef = useAutoResizeTextarea({ text });
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  // pending の最新状態を unmount cleanup から参照するための ref。
  // setPending と並行して更新し、cleanup が stale な closure に閉じ込められても
  // 最新の Object URL を revoke できるようにする。
  const pendingRef = useRef<PendingAttachment[]>([]);
  useEffect(() => {
    pendingRef.current = pending;
  }, [pending]);

  const prevIsSendingRef = useRef(isSending);
  useEffect(() => {
    if (prevIsSendingRef.current && !isSending) {
      textareaRef.current?.focus();
    }
    prevIsSendingRef.current = isSending;
  }, [isSending, textareaRef]);

  // unmount で残った Object URL を確実に revoke する（メモリリーク防止）。
  useEffect(() => {
    return () => {
      pendingRef.current.forEach((a) => {
        if (a.previewUrl) URL.revokeObjectURL(a.previewUrl);
      });
    };
  }, []);

  const anyUploading = pending.some((a) => a.uploading);
  // 送信可能条件: テキストがあるか、エラーでないアップロード済 attachment が 1 件以上ある。
  // errored attachment しか無い状態では送信不可（送ってもユーザーの意図とずれるため）。
  const hasReadyAttachment = pending.some((a) => !a.uploading && !a.error);
  const canSend = !isSending && !anyUploading && (text.trim().length > 0 || hasReadyAttachment);

  const handleSend = () => {
    if (!canSend) return;
    const ready = pending.filter((a) => !a.uploading && !a.error);
    // 失敗・送信中の item の previewUrl は不要なので破棄する。
    // 送信に乗せた item の URL は MessageBubble が描画に使うため revoke せず、
    // 親（useAskAi の messages 配列）が所有権を引き継ぐ。
    // セッション内で送信された画像 URL は long-lived（ページ離脱時にブラウザが
    // 自動回収）。長時間チャットでの累積は PR-G3 (UX 仕上げ) で revoke を
    // 親側に移して対応する想定。
    pending.forEach((a) => {
      if (a.previewUrl && (a.error || a.uploading)) {
        URL.revokeObjectURL(a.previewUrl);
      }
    });
    onSend(text, ready);
    setText('');
    setPending([]);
    setValidationError(null);
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey && !isComposing) {
      e.preventDefault();
      handleSend();
    }
  };

  const handlePickClick = () => {
    fileInputRef.current?.click();
  };

  const handleFileChange = async (e: ChangeEvent<HTMLInputElement>) => {
    const list = e.target.files;
    if (!list || list.length === 0) return;
    const remaining = MAX_ATTACHMENTS - pending.length;
    const files = Array.from(list).slice(0, remaining);

    if (Array.from(list).length > remaining) {
      setValidationError(`添付できるのは ${MAX_ATTACHMENTS} 件までです`);
    } else {
      setValidationError(null);
    }

    // input を即時リセットして同じファイルを再選択できるようにする。
    e.target.value = '';

    for (const file of files) {
      const validation = validateFile(file);
      if (validation) {
        setValidationError(validation);
        continue;
      }
      const previewUrl = URL.createObjectURL(file);
      const tempKey = `local-${Date.now()}-${file.name}`;
      const placeholder: PendingAttachment = {
        key: tempKey,
        filename: file.name,
        contentType: file.type,
        kind: 'image',
        format: FORMAT_FROM_CONTENT_TYPE[file.type] ?? 'png',
        sizeBytes: file.size,
        previewUrl,
        uploading: true,
      };
      setPending((prev) => [...prev, placeholder]);

      try {
        const { uploadUrl, key } = await aiChatRepository.issueAttachmentUploadUrl({
          filename: file.name,
          contentType: file.type,
          sizeBytes: file.size,
        });
        const putRes = await fetch(uploadUrl, {
          method: 'PUT',
          headers: { 'Content-Type': file.type },
          body: file,
        });
        if (!putRes.ok) {
          throw new Error(`upload failed: ${putRes.status}`);
        }
        setPending((prev) =>
          prev.map((a) =>
            a.key === tempKey ? { ...a, key, uploading: false } : a
          )
        );
      } catch (err) {
        setPending((prev) =>
          prev.map((a) =>
            a.key === tempKey
              ? { ...a, uploading: false, error: 'アップロードに失敗しました' }
              : a
          )
        );
        console.error('attachment upload failed', err);
      }
    }
  };

  const handleRemove = (key: string) => {
    setPending((prev) => {
      const target = prev.find((a) => a.key === key);
      if (target?.previewUrl) URL.revokeObjectURL(target.previewUrl);
      return prev.filter((a) => a.key !== key);
    });
  };

  return (
    // 角丸カード型の compose UI。 ChatGPT / Claude.ai に倣い、
    //   1 段目: textarea のみ
    //   2 段目: 左に + (添付)、右に送信ボタンと送信中ラベル
    // を縦に並べて、入力にスペースを取りつつ操作系をボトムバーに集約する。
    <div className="w-full bg-[var(--color-surface-1)] border border-[var(--color-surface-3)] rounded-2xl shadow-sm focus-within:border-[var(--color-text-muted)] transition-colors">
      {pending.length > 0 && (
        <div
          className="flex flex-wrap gap-2 px-4 pt-3"
          aria-label="添付ファイル"
        >
          {pending.map((a) => (
            <AttachmentChip key={a.key} attachment={a} onRemove={handleRemove} />
          ))}
        </div>
      )}
      {validationError && (
        <p className="px-4 pt-2 text-xs text-red-500" role="alert">
          {validationError}
        </p>
      )}

      <input
        ref={fileInputRef}
        type="file"
        accept={ACCEPTED_FILE_TYPES}
        multiple
        className="hidden"
        onChange={handleFileChange}
        aria-label="添付ファイルを選択"
      />

      <div className="px-4 pt-3">
        <textarea
          ref={textareaRef}
          rows={1}
          className="w-full bg-transparent text-[var(--color-text-primary)] outline-none resize-none placeholder:text-[var(--color-text-muted)] text-sm leading-6"
          placeholder="メッセージを入力..."
          value={text}
          onChange={(e) => setText(e.target.value)}
          onKeyDown={handleKeyDown}
          onCompositionStart={() => setIsComposing(true)}
          onCompositionEnd={() => setIsComposing(false)}
          disabled={isSending}
        />
      </div>

      <div className="flex items-center justify-between px-3 py-2">
        <div className="flex items-center gap-1">
          <button
            type="button"
            onClick={handlePickClick}
            disabled={pending.length >= MAX_ATTACHMENTS || isSending}
            className="text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-surface-2)] p-1.5 rounded-full transition-colors duration-150 flex-shrink-0 disabled:opacity-40 disabled:cursor-not-allowed"
            aria-label="添付ファイル"
          >
            <PlusIcon className="h-5 w-5" />
          </button>
        </div>

        <div className="flex items-center gap-2">
          {isSending && (
            <span className="text-xs text-[var(--color-text-muted)]">送信中...</span>
          )}
          <button
            onClick={handleSend}
            className="text-white bg-[var(--color-text-primary)] p-1.5 rounded-full hover:opacity-80 transition-opacity duration-150 flex-shrink-0 disabled:bg-[var(--color-surface-3)] disabled:text-[var(--color-text-muted)] disabled:cursor-not-allowed"
            disabled={!canSend}
            aria-label="送信"
          >
            <ArrowUpIcon className="h-4 w-4" />
          </button>
        </div>
      </div>
    </div>
  );
}

function validateFile(file: File): string | null {
  if (!ACCEPTED_CONTENT_TYPES.has(file.type)) {
    return `${file.name}: 対応していない形式です`;
  }
  if (file.size > MAX_BYTES_PER_FILE) {
    return `${file.name}: 5MB を超える画像は送信できません`;
  }
  if (file.size <= 0) {
    return `${file.name}: ファイルが空です`;
  }
  return null;
}

interface AttachmentChipProps {
  attachment: PendingAttachment;
  onRemove: (key: string) => void;
}

/**
 * 送信前の添付チップ。
 *
 * 画像は 56px 正方形サムネ。onError で blob URL が読めない場合は row 型にフォールバック。
 * 画像以外・エラー・読込失敗は row 型（PhotoIcon + ファイル名 + サイズ）。
 */
function AttachmentChip({ attachment, onRemove }: AttachmentChipProps) {
  const [imgFailed, setImgFailed] = useState(false);
  const canShowImage =
    attachment.kind === 'image' && !!attachment.previewUrl && !attachment.error && !imgFailed;

  if (canShowImage) {
    return (
      <div className="relative flex-shrink-0 w-14 h-14">
        <img
          src={attachment.previewUrl}
          alt=""
          onError={() => setImgFailed(true)}
          className="w-full h-full rounded-xl border border-[var(--color-surface-3)] object-cover bg-[var(--color-surface-2)]"
        />
        {attachment.uploading && (
          <div className="absolute inset-0 rounded-xl bg-black/40 flex items-center justify-center">
            <svg
              className="animate-spin h-4 w-4 text-white"
              aria-hidden="true"
              viewBox="0 0 24 24"
              fill="none"
            >
              <circle
                className="opacity-25"
                cx="12" cy="12" r="10"
                stroke="currentColor"
                strokeWidth="4"
              />
              <path
                className="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
              />
            </svg>
          </div>
        )}
        <button
          type="button"
          onClick={() => onRemove(attachment.key)}
          aria-label={`${attachment.filename} を削除`}
          className="absolute -top-1.5 -right-1.5 w-[18px] h-[18px] flex items-center justify-center rounded-full bg-[var(--color-text-primary)] text-white shadow-sm"
        >
          <XMarkIcon className="w-2.5 h-2.5" />
        </button>
      </div>
    );
  }

  // 画像以外 / 読込失敗 / アップロードエラー時の row 型チップ
  return (
    <div
      className={`relative flex items-center gap-2 pl-2.5 pr-7 py-1.5 rounded-xl border max-w-[200px] ${
        attachment.error
          ? 'border-red-400/60 bg-red-500/8'
          : 'border-[var(--color-surface-3)] bg-[var(--color-surface-2)]'
      }`}
    >
      <div className="w-7 h-7 rounded-lg bg-[var(--color-surface-3)] flex items-center justify-center flex-shrink-0">
        <PhotoIcon className="w-4 h-4 text-[var(--color-text-muted)]" />
      </div>
      <div className="min-w-0">
        <p className="text-xs font-medium text-[var(--color-text-primary)] truncate">
          {attachment.filename}
        </p>
        <p className="text-[10px] text-[var(--color-text-muted)]">
          {attachment.uploading
            ? 'アップロード中...'
            : attachment.error
            ? attachment.error
            : formatBytes(attachment.sizeBytes)}
        </p>
      </div>
      <button
        type="button"
        onClick={() => onRemove(attachment.key)}
        aria-label={`${attachment.filename} を削除`}
        className="absolute right-1.5 top-1/2 -translate-y-1/2 w-4 h-4 flex items-center justify-center rounded-full hover:bg-[var(--color-surface-3)] text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] transition-colors"
      >
        <XMarkIcon className="w-3 h-3" />
      </button>
    </div>
  );
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
