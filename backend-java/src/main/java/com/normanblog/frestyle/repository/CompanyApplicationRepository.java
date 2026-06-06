package com.normanblog.frestyle.repository;

import com.normanblog.frestyle.entity.CompanyApplication;
import java.util.List;
import org.springframework.data.jpa.repository.JpaRepository;

/** company_applications テーブルへのアクセスを担うリポジトリ。 */
public interface CompanyApplicationRepository extends JpaRepository<CompanyApplication, Long> {

  // 一覧は新しい順(super_admin が直近の申請から確認する)。
  List<CompanyApplication> findAllByOrderByCreatedAtDesc();
}
