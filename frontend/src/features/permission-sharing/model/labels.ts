import type { ShareRole } from './types';

/**
 * 役割の選択肢（強い順）。値は backend の domain.GrantRole と同じ。
 *
 * パネルと行の両方が同じ並びを出す必要があるので、ここに 1 つだけ置く
 * （写すと、片方だけ選択肢が増えたときに「一覧では選べるのに追加では選べない」になる）。
 */
export const ROLES: ReadonlyArray<{ value: ShareRole; label: string }> = [
  { value: 'admin', label: '管理' },
  { value: 'editor', label: '編集' },
  { value: 'commenter', label: 'コメント' },
  { value: 'viewer', label: '閲覧' },
];

/**
 * displayName は名前が空のとき ID を代わりに出す。
 *
 * backend は名前を引けなかった相手も空文字で返す（行を落とさない）。ここで
 * 「名前のない行」として出すと、どの権限を消せばよいのか人が選べない。
 */
export function displayName(name: string, principalId: string): string {
  return name.trim() === '' ? principalId : name;
}
