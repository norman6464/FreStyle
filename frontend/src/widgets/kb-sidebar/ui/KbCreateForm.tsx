import { useId, useState } from 'react';
import { slugify } from '../model/slugify';

export interface KbCreateFormProps {
  /** 何を作るのか（「ワークスペース」「スペース」）。ラベルと文言に使う。 */
  what: string;
  /** URL に出る短い名前の欄のラベル。 */
  keyLabel: string;
  /** 確定。**失敗は投げてくる**ので、握り潰さない限り必ず表に出る。 */
  onCreate: (input: { key: string; name: string }) => Promise<void>;
}

/**
 * KbCreateForm はワークスペース / スペースを作る小さな入力欄。
 *
 * 名前と「URL に使う短い名前」の 2 つを受け取る。短い名前は名前から下書きを作るが、
 * **書き換えられる**ようにしてある。日本語の名前からは何も作れない（英数字が残らない）ので、
 * 自動生成だけに頼ると先へ進めなくなる。
 *
 * 失敗しても入力を消さない。消すと、打ち直しになるうえ「何が悪かったのか」も分からない。
 */
export default function KbCreateForm({ what, keyLabel, onCreate }: KbCreateFormProps) {
  const nameId = useId();
  const keyId = useId();
  const [name, setName] = useState('');
  // 短い名前を人が触ったか。触っていない間だけ、名前から下書きを作り直す。
  const [keyTouched, setKeyTouched] = useState(false);
  const [key, setKey] = useState('');
  const [saving, setSaving] = useState(false);

  const canSubmit = name.trim() !== '' && key.trim() !== '' && !saving;

  return (
    <form
      className="space-y-2 px-2 py-3"
      onSubmit={async (event) => {
        event.preventDefault();
        if (!canSubmit) return;
        setSaving(true);
        try {
          await onCreate({ key: key.trim(), name: name.trim() });
        } catch {
          // 入力はそのまま残す。知らせは呼び出し側が出す。
          setSaving(false);
          return;
        }
        setSaving(false);
        setName('');
        setKey('');
        setKeyTouched(false);
      }}
    >
      <div>
        <label htmlFor={nameId} className="mb-0.5 block text-xs text-[var(--color-text-muted)]">
          {what}の名前
        </label>
        <input
          id={nameId}
          type="text"
          value={name}
          onChange={(event) => {
            setName(event.target.value);
            if (!keyTouched) setKey(slugify(event.target.value));
          }}
          className="w-full rounded border border-surface-3 bg-surface-1 px-2 py-1 text-sm text-[var(--color-text-primary)] focus:border-brand-400 focus:outline-none"
        />
      </div>
      <div>
        <label htmlFor={keyId} className="mb-0.5 block text-xs text-[var(--color-text-muted)]">
          {keyLabel}
        </label>
        <input
          id={keyId}
          type="text"
          value={key}
          onChange={(event) => {
            setKeyTouched(true);
            setKey(event.target.value);
          }}
          placeholder="eng"
          className="w-full rounded border border-surface-3 bg-surface-1 px-2 py-1 font-mono text-sm text-[var(--color-text-primary)] focus:border-brand-400 focus:outline-none"
        />
      </div>
      <button
        type="submit"
        disabled={!canSubmit}
        className="w-full rounded-lg bg-brand-500 px-3 py-1.5 text-sm font-medium text-white transition-colors hover:bg-brand-600 disabled:opacity-40"
      >
        {saving ? '作成中…' : `${what}を作る`}
      </button>
    </form>
  );
}
