/**
 * 学習レポート（learning-report）entity のドメイン型。
 */

/** 学習レポート。Go backend `domain.LearningReport` と 1:1。
 *  非同期生成のため status (`pending` / `ready` / `failed`) と s3Key を持つ。 */
export interface LearningReport {
  id: number;
  userId: number;
  /** 対象期間の開始（月初、ISO 文字列）。表示の対象月はここから導出する。 */
  periodFrom: string;
  /** 対象期間の終了（翌月初、半開区間）。 */
  periodTo: string;
  /** 生成状態。pending = 作成中 / ready = 完成 / failed = 失敗。 */
  status: 'pending' | 'ready' | 'failed';
  s3Key?: string;
  createdAt: string;
}
