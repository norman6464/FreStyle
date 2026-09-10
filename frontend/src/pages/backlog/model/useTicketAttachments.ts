import { useCallback, useEffect, useRef, useState } from 'react';
import { TicketRepository, type TicketAttachment } from '@/entities/ticket';
import { isAcceptedAttachmentContentType, MAX_ATTACHMENT_UPLOAD_BYTES } from '../config/attachmentUpload';

const LOAD_FAILED = '添付を読み込めませんでした。時間をおいて開き直すと最新の状態が出ます。';
const REJECTED_TYPE = '対応していない形式のファイルです。';
const REJECTED_SIZE = 'ファイルが大きすぎます（上限 25 MB）。';
const UPLOAD_FAILED = 'アップロードに失敗しました。';

/** アップロード中・失敗の 1 件。成功すると消え、`attachments` 側に移る。 */
export interface PendingAttachment {
  /** 手元だけの ID（サーバーはまだ何も知らない）。 */
  clientId: string;
  file: File;
  status: 'uploading' | 'failed';
  error: string | null;
}

/**
 * useTicketAttachments はチケット 1 件の添付（一覧・追加・削除）を読み書きする。
 *
 * 追加は 3 段（presign 発行 → Cloud Storage へ直接 PUT → メタデータの記録）で、
 * 手元では `pending` に「アップロード中/失敗」の行として持ち、成功した分だけ
 * `attachments`（確定済み）へ移す。一覧の取得し直しを待たずに進捗を見せるため、
 * ラベルのような素朴な CRUD 一覧より 1 段複雑な形になっている。
 *
 * 楽観更新はしない（確定済み側は応答をそのまま反映）。失敗した pending 行は
 * 消さずに残し、`retry` でファイルを持ち回したまま再送できるようにする。
 */
export function useTicketAttachments(workspaceSlug: string | undefined, ticketId: string | undefined) {
  const [attachments, setAttachments] = useState<TicketAttachment[]>([]);
  const [pending, setPending] = useState<PendingAttachment[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [busyId, setBusyId] = useState<string | null>(null);
  const active = useRef<string | null>(null);

  const load = useCallback(async (slug: string, id: string) => {
    const key = `${slug} ${id}`;
    setLoading(true);
    setError(null);
    try {
      const list = await TicketRepository.fetchTicketAttachments(slug, id);
      if (active.current !== key) return;
      setAttachments(list);
    } catch {
      if (active.current !== key) return;
      setError(LOAD_FAILED);
    } finally {
      if (active.current === key) setLoading(false);
    }
  }, []);

  useEffect(() => {
    const key = workspaceSlug && ticketId ? `${workspaceSlug} ${ticketId}` : null;
    active.current = key;
    setPending([]);
    if (!workspaceSlug || !ticketId) {
      setAttachments([]);
      return;
    }
    void load(workspaceSlug, ticketId);
  }, [workspaceSlug, ticketId, load]);

  const refresh = useCallback(() => {
    if (workspaceSlug && ticketId) void load(workspaceSlug, ticketId);
  }, [workspaceSlug, ticketId, load]);

  const runUpload = useCallback(async (clientId: string, slug: string, id: string, file: File) => {
    const key = `${slug} ${id}`;
    try {
      const issued = await TicketRepository.issueTicketAttachmentUploadUrl(slug, id, file.type, file.size);
      await TicketRepository.putTicketAttachmentFile(issued.url, file);
      const created = await TicketRepository.createTicketAttachment(slug, id, {
        key: issued.key,
        filename: file.name,
        contentType: file.type,
        sizeBytes: file.size,
      });
      if (active.current !== key) return;
      setAttachments((prev) => [...prev, created]);
      setPending((prev) => prev.filter((p) => p.clientId !== clientId));
    } catch {
      if (active.current !== key) return;
      setPending((prev) =>
        prev.map((p) => (p.clientId === clientId ? { ...p, status: 'failed', error: UPLOAD_FAILED } : p)),
      );
    }
  }, []);

  /** ファイル 1 件を選ぶたびに呼ぶ（複数選択は呼び出し側が 1 つずつ回す）。 */
  const upload = useCallback(
    (file: File) => {
      if (!workspaceSlug || !ticketId) return;
      const clientId = crypto.randomUUID();
      if (!isAcceptedAttachmentContentType(file.type)) {
        setPending((prev) => [...prev, { clientId, file, status: 'failed', error: REJECTED_TYPE }]);
        return;
      }
      if (file.size <= 0 || file.size > MAX_ATTACHMENT_UPLOAD_BYTES) {
        setPending((prev) => [...prev, { clientId, file, status: 'failed', error: REJECTED_SIZE }]);
        return;
      }
      setPending((prev) => [...prev, { clientId, file, status: 'uploading', error: null }]);
      void runUpload(clientId, workspaceSlug, ticketId, file);
    },
    [workspaceSlug, ticketId, runUpload],
  );

  const retry = useCallback(
    (clientId: string) => {
      if (!workspaceSlug || !ticketId) return;
      setPending((prev) => {
        const target = prev.find((p) => p.clientId === clientId);
        if (target) void runUpload(clientId, workspaceSlug, ticketId, target.file);
        return prev.map((p) => (p.clientId === clientId ? { ...p, status: 'uploading', error: null } : p));
      });
    },
    [workspaceSlug, ticketId, runUpload],
  );

  const dismiss = useCallback((clientId: string) => {
    setPending((prev) => prev.filter((p) => p.clientId !== clientId));
  }, []);

  /** 確定済みの添付を削除する。失敗は投げる（呼び出し側がトーストで知らせる）。 */
  const remove = useCallback(
    async (attachmentId: string) => {
      if (!workspaceSlug || !ticketId) throw new Error('ticket attachments: no active scope');
      const key = `${workspaceSlug} ${ticketId}`;
      setBusyId(attachmentId);
      try {
        await TicketRepository.deleteTicketAttachment(workspaceSlug, ticketId, attachmentId);
        if (active.current === key) {
          setAttachments((prev) => prev.filter((a) => a.id !== attachmentId));
          setBusyId(null);
        }
      } catch (cause) {
        if (active.current === key) setBusyId(null);
        throw cause;
      }
    },
    [workspaceSlug, ticketId],
  );

  return { attachments, pending, loading, error, busyId, refresh, upload, retry, dismiss, remove };
}
