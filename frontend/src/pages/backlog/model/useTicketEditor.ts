import { useCallback, useEffect, useRef, useState } from 'react';
import type { Ticket, TicketPriority, UpdateTicketInput } from '@/entities/ticket';
import type { SaveStatus } from '@/shared/ui/RichTextEditor';

/** 打鍵が止まってから保存を撃つまでの間合い（useKbPageDoc と同じ値）。 */
const SAVE_DEBOUNCE_MS = 800;

interface Draft {
  title: string;
  doc: unknown;
  priority: TicketPriority;
  startDate: string | null;
  dueDate: string | null;
}

/**
 * useTicketEditor は詳細パネルでのチケット編集（title / doc / priority / 日付）を持つ。
 *
 * `PUT .../tickets/:id` は全置換（設計 Ⅳ-D）— 1 項目だけ変えても他の項目を送り返す
 * 必要がある。呼び出し側（TicketDetailPanel）は `key={ticket.id}` でこの hook ごと
 * マウントし直す前提で書いてある（チケットを切り替えるたびに下書きを作り直す。
 * KB の複数ページ間切り替えほど頻繁ではないので、KbPageDoc のような宛先ごとの
 * 保留キューは持たない — アンマウント時に保留があれば flush するだけで足りる）。
 */
export function useTicketEditor(
  ticket: Ticket,
  canEdit: boolean,
  onUpdate: (input: UpdateTicketInput) => Promise<Ticket>,
) {
  const [title, setTitle] = useState(ticket.title);
  const [doc, setDoc] = useState(ticket.doc);
  const [priority, setPriority] = useState(ticket.priority);
  const [startDate, setStartDate] = useState(ticket.startDate);
  const [dueDate, setDueDate] = useState(ticket.dueDate);
  const [saveStatus, setSaveStatus] = useState<SaveStatus>('idle');

  const draftRef = useRef<Draft>({ title, doc, priority, startDate, dueDate });
  useEffect(() => {
    draftRef.current = { title, doc, priority, startDate, dueDate };
  });

  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const commit = useCallback(
    (override?: Partial<Draft>) => {
      if (!canEdit) return;
      const snapshot = { ...draftRef.current, ...override };
      setSaveStatus('saving');
      onUpdate({
        title: snapshot.title,
        doc: snapshot.doc,
        typeId: ticket.typeId,
        priority: snapshot.priority,
        startDate: snapshot.startDate ?? undefined,
        dueDate: snapshot.dueDate ?? undefined,
      })
        .then(() => setSaveStatus('saved'))
        .catch(() => setSaveStatus('unsaved'));
    },
    [canEdit, onUpdate, ticket.typeId],
  );

  const scheduleCommit = useCallback(() => {
    if (!canEdit) return;
    setSaveStatus('unsaved');
    if (timer.current) clearTimeout(timer.current);
    timer.current = setTimeout(() => {
      timer.current = null;
      commit();
    }, SAVE_DEBOUNCE_MS);
  }, [canEdit, commit]);

  // アンマウント（別チケットの選択・パネルを閉じる）時に保留中の編集を捨てない。
  useEffect(
    () => () => {
      if (timer.current) {
        clearTimeout(timer.current);
        timer.current = null;
        commit();
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [],
  );

  const changeTitle = useCallback((value: string) => setTitle(value), []);
  const commitTitle = useCallback(() => {
    if (title !== ticket.title) commit();
  }, [title, ticket.title, commit]);

  const changeDoc = useCallback(
    (value: unknown) => {
      setDoc(value);
      scheduleCommit();
    },
    [scheduleCommit],
  );

  const changePriority = useCallback(
    (value: TicketPriority) => {
      setPriority(value);
      commit({ priority: value });
    },
    [commit],
  );

  const changeStartDate = useCallback(
    (value: string | null) => {
      setStartDate(value);
      commit({ startDate: value });
    },
    [commit],
  );

  const changeDueDate = useCallback(
    (value: string | null) => {
      setDueDate(value);
      commit({ dueDate: value });
    },
    [commit],
  );

  return {
    title,
    doc,
    priority,
    startDate,
    dueDate,
    saveStatus,
    changeTitle,
    commitTitle,
    changeDoc,
    changePriority,
    changeStartDate,
    changeDueDate,
  };
}
