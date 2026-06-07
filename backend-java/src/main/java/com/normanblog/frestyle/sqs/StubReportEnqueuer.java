package com.normanblog.frestyle.sqs;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

/**
 * SQS への実送信を伴わない no-op 実装(Go 版の stubEnqueuer と同等)。
 *
 * <p>レポート生成ワーカー(SQS コンシューマ)の契約が未確定のため、現状はキュー投入をログのみに留める。
 * 実 SQS 送信(AWS SDK)は別 PR で差し替える。
 */
@Component
public class StubReportEnqueuer implements ReportEnqueuer {

  private static final Logger log = LoggerFactory.getLogger(StubReportEnqueuer.class);

  @Override
  public void enqueue(Long reportId) {
    log.debug("learning-report enqueue (stub no-op): reportId={}", reportId);
  }
}
