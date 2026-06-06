package com.normanblog.frestyle.repository;

import com.normanblog.frestyle.entity.Course;
import java.util.List;
import org.springframework.data.jpa.repository.JpaRepository;

/** courses テーブルへのアクセスを担うリポジトリ。並び順は sort_order(同値時 id)昇順。 */
public interface CourseRepository extends JpaRepository<Course, Long> {

  List<Course> findByCompanyIdOrderBySortOrderAscIdAsc(Long companyId);

  List<Course> findByCompanyIdAndIsPublishedTrueOrderBySortOrderAscIdAsc(Long companyId);
}
