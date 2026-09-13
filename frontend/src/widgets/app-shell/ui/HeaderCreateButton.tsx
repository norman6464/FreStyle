import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { PlusIcon } from '@heroicons/react/24/outline';
import { useToast } from '@/shared/lib/hooks/useToast';
import { emitKbTreeEvent, KbRepository, NOTE_NEW_PAGE_TITLE, useCurrentKbSpace } from '@/entities/kb';

/**
 * HeaderCreateButton はヘッダーの「作成」（段3・PR-3）。
 *
 * 今いるスペース（entities/kb の共有値）が分かっているときだけ出す —
 * `createPage` は spaceId を必須とし、ナレッジ以外の画面では作る先が無いため。
 * 作成後はサイドバーの木にも反映されるよう page-created を通知してから遷移する
 * （KbSidebar.createRootPage と同じ形。題名は「無題」のまま開き、本文と同じ
 * インライン編集でその場で書き直す — サイドバーの木のような専用の rename モードは
 * 持たない）。
 */
export default function HeaderCreateButton({ className }: { className: string }) {
  const current = useCurrentKbSpace();
  const navigate = useNavigate();
  const { showToast } = useToast();
  const [creating, setCreating] = useState(false);

  if (!current) return null;

  const handleCreate = async () => {
    if (creating) return;
    setCreating(true);
    try {
      const page = await KbRepository.createPage(current.workspaceSlug, current.spaceId, {
        title: NOTE_NEW_PAGE_TITLE,
      });
      emitKbTreeEvent({ type: 'page-created', page });
      navigate(`/kb/${page.id}`);
    } catch {
      showToast('error', 'ページを作成できませんでした');
    } finally {
      setCreating(false);
    }
  };

  return (
    <button type="button" onClick={() => void handleCreate()} disabled={creating} className={className}>
      <PlusIcon className="h-4 w-4" aria-hidden="true" />
      作成
    </button>
  );
}
