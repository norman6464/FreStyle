package com.normanblog.frestyle.repository;

import com.normanblog.frestyle.entity.TeachingMaterial;
import java.util.List;
import org.springframework.data.jpa.repository.JpaRepository;

/** teaching_materials テーブルへのアクセス。並び順は order_in_course(同値時 id)昇順。 */
public interface TeachingMaterialRepository extends JpaRepository<TeachingMaterial, Long> {

  List<TeachingMaterial> findByCourseIdOrderByOrderInCourseAscIdAsc(Long courseId);

  List<TeachingMaterial> findByCourseIdAndIsPublishedTrueOrderByOrderInCourseAscIdAsc(Long courseId);
}
