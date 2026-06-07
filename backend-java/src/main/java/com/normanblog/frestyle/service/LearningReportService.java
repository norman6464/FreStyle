package com.normanblog.frestyle.service;

import com.normanblog.frestyle.entity.LearningReport;
import com.normanblog.frestyle.entity.LearningReportStatus;
import com.normanblog.frestyle.repository.LearningReportRepository;
import com.normanblog.frestyle.infra.sqs.ReportEnqueuer;
import java.time.Instant;
import java.time.LocalDate;
import java.time.ZoneOffset;
import java.util.List;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

/** 学習レポートの一覧取得と、月次レポート生成ジョブの受付(DB 行作成 + キュー投入)を担う。 */
@Service
public class LearningReportService {

  private final LearningReportRepository reports;
  private final ReportEnqueuer enqueuer;

  public LearningReportService(LearningReportRepository reports, ReportEnqueuer enqueuer) {
    this.reports = reports;
    this.enqueuer = enqueuer;
  }

  /** current user のレポートを期間(period_to)降順で返す。 */
  public List<LearningReport> list(Long userId) {
    return reports.findByUserIdOrderByPeriodToDesc(userId);
  }

  /**
   * 指定月のレポート生成ジョブを pending 行で作り、キューに積む。
   *
   * <p>year + month を月初〜翌月初の半開区間 [periodFrom, periodTo) に変換する。
   */
  @Transactional
  public LearningReport requestMonthly(Long userId, int year, int month) {
    Instant from = LocalDate.of(year, month, 1).atStartOfDay(ZoneOffset.UTC).toInstant();
    Instant to = LocalDate.of(year, month, 1).plusMonths(1).atStartOfDay(ZoneOffset.UTC).toInstant();

    LearningReport report =
        reports.save(
            LearningReport.builder()
                .userId(userId)
                .periodFrom(from)
                .periodTo(to)
                .status(LearningReportStatus.PENDING)
                .createdAt(Instant.now())
                .build());
    enqueuer.enqueue(report.getId());

    return report;
  }
}
