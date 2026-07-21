/*
 * entities/company の Public API。
 *
 * 外から使ってよいものだけを名前付きで re-export する（FSD 公式仕様。`export *` は使わない）。
 * この Slice の内部ファイルを直接 import してはいけない。
 */

export { default as CompanyRepository } from './api/companyRepository';
export type { Company } from './api/companyRepository';
export type { CompanyStat } from './api/companyRepository';
export { default as CompanySettingsRepository } from './api/companySettingsRepository';
export type { CompanySettings } from './api/companySettingsRepository';
export { CompanyApplicationRepository } from './api/companyApplicationRepository';
export type { CompanyApplicationForm } from './api/companyApplicationRepository';
export type { CompanyApplicationStatus } from './api/companyApplicationRepository';
export type { CompanyApplication } from './api/companyApplicationRepository';
