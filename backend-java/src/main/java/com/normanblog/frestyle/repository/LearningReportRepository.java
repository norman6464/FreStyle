package com.normanblog.frestyle.repository;

import com.normanblog.frestyle.entity.LearningReport;
import java.util.List;
import org.springframework.data.jpa.repository.JpaRepository;

/** learning_reports テーブルへのアクセスを担うリポジトリ。 */
public interface LearningReportRepository extends JpaRepository<LearningReport, Long> {

  List<LearningReport> findByUserIdOrderByPeriodToDesc(Long userId);
}
