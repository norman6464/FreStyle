import type { NoteGrantRole } from '@/entities/note';

/**
 * 付与で与える役割。ノートも教材も同じ 4 つで、backend の domain.GrantRole と対応する。
 *
 * ノート側の型を再輸出しているのは、**同じものが 2 つあると必ずずれる**ため。
 * 役割が増えたときに片方だけ増える、という壊れ方を型で防ぐ。
 */
export type ShareRole = NoteGrantRole;

/** 一覧に並ぶ 1 行（張った権限と、その相手の表示名を突き合わせたもの）。 */
export interface ShareRow {
  principalId: string;
  role: ShareRole;
  /** 表示名。引けなかった相手は空文字。 */
  name: string;
  kind: SharePrincipal['kind'] | 'unknown';
}

/**
 * 権限を張れる相手 1 件。
 *
 * name は表示名で、**引けなかった場合は空文字**（backend が行を落とさずそう返す）。
 * 画面もそれに合わせて行を消さない — 消すと、その相手に張った権限が一覧に出たまま
 * 選べなくなる。
 */
export interface SharePrincipal {
  id: string;
  kind: 'user' | 'group' | 'space_all';
  name: string;
}

/**
 * 共有パネルが必要とする状態と操作。
 *
 * 取得の仕方（どの API を叩くか）は持たない。ノートはページ単位、教材はコース / 章単位で
 * 口が違うが、**画面の見え方と操作は同じ**なので、この形だけを共通にする。
 */
export interface ShareState {
  rows: ShareRow[];
  candidates: SharePrincipal[];
  loading: boolean;
  /** 失敗の理由。null なら失敗していない。 */
  error: string | null;
  /** 書き込みが飛んでいる間 true（二重送信を止める）。 */
  saving: boolean;
  /** 付与。**成功したかを返す**（失敗したときに選択を消さないため）。 */
  grant: (principalId: string, role: ShareRole) => Promise<boolean>;
  revoke: (principalId: string) => Promise<boolean>;
}
