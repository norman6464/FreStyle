import type { Label } from '@/entities/ticket';
import { labelPaint } from '../lib/labelTextColor';

export interface TicketLabelChipProps {
  label: Label;
}

/**
 * ラベル 1 件の見た目。塗り方は色の明るさで決まる（labelTextColor）。
 *
 * 地に敷けない明るさのときは枠線と点にだけ色を使う。同じ画面の中で塗り方が
 * 揃わないことになるが、読めない文字を出すよりよい。
 */
export default function TicketLabelChip({ label }: TicketLabelChipProps) {
  const paint = labelPaint(label.color);

  if (paint.kind === 'solid') {
    return (
      <span
        className="inline-flex items-center rounded px-1.5 py-0.5 text-[11px] font-semibold leading-relaxed"
        style={{ backgroundColor: paint.background, color: paint.color }}
      >
        {label.name}
      </span>
    );
  }

  return (
    <span
      className="inline-flex items-center gap-1 rounded border bg-surface-1 px-1.5 py-0.5 text-[11px] font-semibold leading-relaxed text-[var(--color-text-primary)]"
      style={{ borderColor: paint.borderColor }}
    >
      <span
        aria-hidden="true"
        className="h-2 w-2 flex-none rounded-full"
        style={{ backgroundColor: paint.borderColor }}
      />
      {label.name}
    </span>
  );
}
