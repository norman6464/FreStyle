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
 * 階層: Company 1 ── * Course 1 ── * TeachingMaterial
 *
 * - company_admin: 自社の draft 含む全件 list / 編集 / 削除可
 * - trainee: 自社の `isPublished=true` コースのみ閲覧可
 */
export interface Course {
  id: number;
  companyId: number;
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
 * 必ず 1 つの Course に所属する Markdown 教材。
 *
 * - company_admin: 自社の draft 含む全件 list / 編集 / 削除可
 * - trainee: 自社の `isPublished=true` 教材かつ所属コース published のみ閲覧可
 */
export interface TeachingMaterial {
  id: number;
  companyId: number;
  courseId: number;
  createdByUserId: number;
  title: string;
  content: string;
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

/** 章閲覧記録（`GET /api/v2/me/dashboard` の recentChapterViews 要素）。 */
export interface UserChapterView {
  userId: number;
  teachingMaterialId: number;
  courseId: number;
  firstViewedAt: string;
  lastViewedAt: string;
  viewCount: number;
}
