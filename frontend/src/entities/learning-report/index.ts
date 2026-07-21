/*
 * entities/learning-report の Public API。
 *
 * 外から使ってよいものだけを名前付きで re-export する（FSD 公式仕様。`export *` は使わない）。
 * この Slice の内部ファイルを直接 import してはいけない。
 */

export { LearningReportRepository } from './api/learningReportRepository';

export type {
  LearningReport,
} from './model/types';

export { default as ReportCard } from './ui/ReportCard';
