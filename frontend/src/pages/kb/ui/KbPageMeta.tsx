import type { KbEditorRef } from '@/entities/kb';
import { formatHourMinute, formatMonthDay } from '@/shared/lib/formatters';

export interface KbPageMetaProps {
  lastEditedBy?: KbEditorRef | null;
  lastEditedAt?: string | null;
}

/**
 * KbPageMeta は題名の下に出す「最終編集」の一行。
 *
 * lastEditedBy / lastEditedAt が無ければ何も出さない（旧応答・未保存のページの
 * どちらも該当し得る。無いことを匂わせる空欄を置かない）。
 * name が引けなければ「不明なユーザー」に倒す（ListGrantablePrincipals と同じ、
 * 行を消さず埋める約束）。
 */
export default function KbPageMeta({ lastEditedBy, lastEditedAt }: KbPageMetaProps) {
  if (!lastEditedBy || !lastEditedAt) return null;

  return (
    <p className="mb-6 text-xs text-[var(--color-text-muted)]">
      最終編集 {lastEditedBy.name || '不明なユーザー'} · {formatMonthDay(lastEditedAt)}{' '}
      {formatHourMinute(lastEditedAt)}
    </p>
  );
}
