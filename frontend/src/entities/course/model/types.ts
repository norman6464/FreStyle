import type { RichDocContent } from '@/shared/ui/RichTextEditor';
/**
 * コース（course）entity のドメイン型。
 *
 * 教材（TeachingMaterial）はコースの章であり単独では存在しないため同じ Slice に置く。
 * 別 Slice にすると courseRepository が教材の型を参照した時点で
 * 同一レイヤーの Slice 間 import になり FSD 違反になる。
 */

/**
 * Course は教材を束ねる「コース（プロジェクト）」。 backend `domain.Course` と 1:1。
 *
 * 階層: Workspace 1 ── * Course 1 ── * TeachingMaterial
 *
 * - company_admin: 自社の draft 含む全件 list / 編集 / 削除可
 * - trainee: 自社の `isPublished=true` コースのみ閲覧可
 */
export interface Course {
  id: number;
  workspaceId?: string;
  createdByUserId: number;
  title: string;
  description: string;
  /** 学習領域カテゴリ（空 = 未分類。entities/course/config/courseCategories の key と対応） */
  category: string;
  /** 主に扱う言語・技術（例: 'go' / 'docker'。空 = 言語が主題でない → バッジ非表示） */
  language: string;
  sortOrder: number;
  isPublished: boolean;
  createdAt: string;
  updatedAt: string;
  /**
   * 書き換えられるか（編集 UI を出すかの判定）。
   *
   * **画面が自分で判断しない。** 以前はアプリのロール（company_admin なら編集できる）で
   * 出していたが、可否は対象ごとの付与で決まるので、ロールで出すと「ボタンは出るのに
   * 保存が弾かれる」状態になる。サーバーの答えに従う。
   */
  canEdit?: boolean;
  /** 権限そのものを変えられるか（共有ボタンを出すかの判定）。 */
  canManage?: boolean;
}

/**
 * CourseWithProgress はコース一覧 API (`GET /api/v2/courses`) の要素。
 * backend `usecase.CourseWithProgress` と 1:1(Course に章数と完了章数を合成したフラット JSON)。
 */
export interface CourseWithProgress extends Course {
  /** コース内の教材(章)数。trainee は published のみ、admin 系は下書き込み。 */
  materialCount: number;
  /** current user が完了した章数(現存する published 章のみ。常に materialCount 以下)。 */
  completedCount: number;
}

/**
 * TeachingMaterial は Go backend `domain.TeachingMaterial` と 1:1。
 * 必ず 1 つの Course に所属する教材（本文はリッチ本文 doc が正本）。
 *
 * - company_admin: 自社の draft 含む全件 list / 編集 / 削除可
 * - trainee: 自社の `isPublished=true` 教材かつ所属コース published のみ閲覧可
 */
export interface TeachingMaterial {
  id: number;
  workspaceId?: string;
  courseId: number;
  createdByUserId: number;
  title: string;
  /**
   * リッチ本文（tiptap JSON）。詳細 GET のみが返す（一覧には含まれない）。
   * まだ本文を保存していない新規章は null（空 doc として扱う）。
   */
  doc?: RichDocContent | null;
  /** doc 更新の楽観ロック版数。詳細 GET のみが返す。 */
  revision?: number;
  orderInCourse: number;
  isPublished: boolean;
  createdAt: string;
  updatedAt: string;
}

/**
 * UserLessonProgress は trainee 自身の教材（レッスン）完了記録。 backend `domain.UserLessonProgress` と 1:1。
 * current user 固定で、 他人の進捗は取得・操作できない。
 */
export interface UserLessonProgress {
  id: number;
  userId: number;
  teachingMaterialId: number;
  courseId: number;
  completedAt: string;
  createdAt: string;
}

/** 章閲覧記録（コース詳細のレジューム表示用。`GET .../courses/:id/last-viewed` が返す）。 */
export interface UserChapterView {
  userId: number;
  teachingMaterialId: number;
  courseId: number;
  firstViewedAt: string;
  lastViewedAt: string;
  viewCount: number;
}

/** 教材の付与で与える役割。強い順に admin > editor > commenter > viewer。 */
export type MaterialGrantRole = 'admin' | 'editor' | 'commenter' | 'viewer';

/**
 * コース / 教材に張られた付与 1 件。
 *
 * **「これを編集できる人」ではない。** 返るのはその段に張った行だけで、
 * ワークスペースの admin も、コースから章へ降りてくる分も含まれない。
 */
export interface MaterialGrant {
  principalId: string;
  role: MaterialGrantRole;
  createdAt: string;
  updatedAt: string;
}

/**
 * 権限を張れる相手 1 件。
 *
 * name は表示名で、**引けなかった場合は空文字**（backend が行を落とさずそう返す）。
 * 画面もそれに合わせて行を消さない。
 */
export interface MaterialPrincipal {
  id: string;
  kind: 'user' | 'group' | 'space_all';
  name: string;
}
