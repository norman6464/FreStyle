/*
 * entities/course の Public API。
 *
 * 外から使ってよいものだけを名前付きで re-export する（FSD 公式仕様。`export *` は使わない）。
 * この Slice の内部ファイル（`@/entities/course/api/courseRepository` など）を
 * 直接 import してはいけない。
 */

export { default as CourseRepository } from './api/courseRepository';
export type { CoursePayload } from './api/courseRepository';
export { default as TeachingMaterialRepository } from './api/teachingMaterialRepository';
export { default as LessonProgressRepository } from './api/lessonProgressRepository';

export type {
  Course,
  CourseWithProgress,
  TeachingMaterial,
  UserLessonProgress,
  UserChapterView,
} from './model/types';

export { default as CourseProgressBar } from './ui/CourseProgressBar';

export { COURSE_CATEGORIES, findCourseCategory } from './config/courseCategories';
export { COURSE_LANGUAGES } from './config/courseLanguages';
export type { CourseCategoryDef } from './config/courseCategories';
export type { CourseLanguageDef } from './config/courseLanguages';
export type {
  TeachingMaterialCreatePayload,
  TeachingMaterialUpdatePayload,
} from './api/teachingMaterialRepository';
