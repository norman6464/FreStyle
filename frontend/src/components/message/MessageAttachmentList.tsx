/**
 * 自分のメッセージに添付された画像 / ドキュメントの参照（送信中の Object URL も許容）。
 * MessageBubble 内では画像のみインラインプレビュー、それ以外はファイル名カードに落とす。
 *
 * `kind` は backend `domain.AttachmentKind*` と整合する `'image' | 'document'`。
 */
export interface MessageAttachmentView {
  key: string;
  filename: string;
  contentType: string;
  kind: 'image' | 'document';
  sizeBytes: number;
  /** 送信中チップから引き継ぐローカル Object URL */
  previewUrl?: string;
  /** 送信完了後 backend が返す CDN / S3 URL（あれば優先） */
  url?: string;
}

/**
 * 自分のメッセージに紐付く添付の表示。
 *
 * 画像は max-w 240px のサムネで横並び（Object URL でローカル送信中も表示できる）。
 * 画像以外（PR-G2 で増える PDF / CSV）は filename カードのフォールバック。
 */
export default function MessageAttachmentList({ attachments }: { attachments: MessageAttachmentView[] }) {
  return (
    <div className="flex flex-wrap gap-2 justify-end" aria-label="添付ファイル">
      {attachments.map((a) => {
        const src = a.url ?? a.previewUrl;
        if (a.kind === 'image' && src) {
          return (
            <img
              key={a.key}
              src={src}
              alt={a.filename}
              className="max-w-[240px] max-h-[240px] rounded-lg object-cover border border-[var(--color-surface-3)]"
            />
          );
        }
        return (
          <div
            key={a.key}
            className="px-3 py-2 rounded-lg border border-[var(--color-surface-3)] bg-[var(--color-surface-2)] text-xs text-[var(--color-text-primary)]"
          >
            {a.filename}
          </div>
        );
      })}
    </div>
  );
}
