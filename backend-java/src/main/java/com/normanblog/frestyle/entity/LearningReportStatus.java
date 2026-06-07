package com.normanblog.frestyle.entity;

/** learning_reports.status の許容値。DB には文字列で保存する(既存 Go 実装と同値)。 */
public final class LearningReportStatus {

  // 生成ジョブ受付済み・未完了。
  public static final String PENDING = "pending";
  // 生成完了・成果物あり。
  public static final String READY = "ready";
  // 生成失敗。
  public static final String FAILED = "failed";

  private LearningReportStatus() {}
}
