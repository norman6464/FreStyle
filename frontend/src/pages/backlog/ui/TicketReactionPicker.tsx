import { REACTION_EMOJIS } from '../config/reactionEmojis';

export interface TicketReactionPickerProps {
  onPick: (emoji: string) => void;
}

/** 厳選 12 種の絵文字の格子。自由入力は無い（config/reactionEmojis.ts）。 */
export default function TicketReactionPicker({ onPick }: TicketReactionPickerProps) {
  return (
    <div role="group" aria-label="反応を選ぶ" className="mt-1 grid grid-cols-6 gap-1 rounded-lg border border-surface-3 bg-surface-1 p-1.5">
      {REACTION_EMOJIS.map((emoji) => (
        <button
          key={emoji}
          type="button"
          aria-label={`${emoji} の反応を付ける`}
          onClick={() => onPick(emoji)}
          className="grid h-7 w-7 place-items-center rounded text-base hover:bg-surface-2"
        >
          <span aria-hidden="true">{emoji}</span>
        </button>
      ))}
    </div>
  );
}
