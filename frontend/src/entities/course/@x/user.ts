/*
 * entities/user 向けのクロス import 用 Public API（FSD の `@x` 記法）。
 *
 * 同一レイヤーの Slice 同士は本来 import できないが、FSD 公式は entities 層に限り
 * `@x` 記法での明示的なクロス import を認めている。
 *
 * なぜ必要か: `UserDashboard`（entities/user）は「最近見た章」を持つため
 * `UserChapterView`（entities/course）を参照する。これは実データの構造上の依存で、
 * どちらかに寄せると「章の型を user が定義する」「ダッシュボードを course が知る」の
 * どちらかの不自然さが出る。
 *
 * ここに出すものは最小限に保つこと。増えてきたら Slice の切り方自体を見直す合図。
 */
export type { UserChapterView } from '../model/types';
