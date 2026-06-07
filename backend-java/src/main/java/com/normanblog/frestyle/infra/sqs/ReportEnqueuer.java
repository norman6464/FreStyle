package com.normanblog.frestyle.infra.sqs;

/** 学習レポート生成ジョブを非同期キュー(SQS)に投入する境界。 */
public interface ReportEnqueuer {

  /** 指定レポートの生成ジョブをキューに積む。 */
  void enqueue(Long reportId);
}
