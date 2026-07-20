/*
 * entities/audit の Public API。
 *
 * 外から使ってよいものだけを名前付きで re-export する（FSD 公式仕様。`export *` は使わない）。
 * この Slice の内部ファイルを直接 import してはいけない。
 */

export { AuditRepository } from './api/auditRepository';
export type { AuditEvent } from './api/auditRepository';
