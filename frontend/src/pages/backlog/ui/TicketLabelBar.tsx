import { useState } from 'react';
import type { Label } from '@/entities/ticket';
import TicketLabelChip from './TicketLabelChip';
import TicketLabelPicker from './TicketLabelPicker';

export interface TicketLabelBarProps {
  /** このチケットに付いているラベル。 */
  attached: Label[];
  /** スペースに定義されている全ラベル（ピッカーの選択肢）。 */
  allLabels: Label[];
  canEdit: boolean;
  onToggle: (label: Label) => void;
  onCreate: (name: string, color: string) => Promise<Label>;
}

/**
 * チケットのラベル帯。読むだけのときはチップの列だけ（＋ボタンは出さない）。
 * 編集できるときだけ選択（TicketLabelPicker）を開ける。
 */
export default function TicketLabelBar({ attached, allLabels, canEdit, onToggle, onCreate }: TicketLabelBarProps) {
  const [pickerOpen, setPickerOpen] = useState(false);

  if (!canEdit && attached.length === 0) return null;

  return (
    <div className="mb-4">
      <div className="flex flex-wrap items-center gap-1.5">
        {attached.map((label) => (
          <TicketLabelChip key={label.id} label={label} />
        ))}
        {canEdit && (
          <button
            type="button"
            onClick={() => setPickerOpen((v) => !v)}
            aria-expanded={pickerOpen}
            aria-label="ラベルを付ける"
            className="grid h-5 w-5 place-items-center rounded border border-dashed border-surface-3 text-xs text-[var(--color-text-muted)] hover:bg-surface-2"
          >
            ＋
          </button>
        )}
      </div>
      {canEdit && pickerOpen && (
        <TicketLabelPicker
          labels={allLabels}
          attachedIds={attached.map((l) => l.id)}
          onToggle={onToggle}
          onCreate={onCreate}
        />
      )}
    </div>
  );
}
