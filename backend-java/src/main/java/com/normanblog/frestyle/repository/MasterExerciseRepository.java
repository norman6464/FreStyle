package com.normanblog.frestyle.repository;

import com.normanblog.frestyle.entity.MasterExercise;
import java.util.List;
import java.util.Optional;
import org.springframework.data.jpa.repository.JpaRepository;

/** master_exercises テーブルへのアクセスを担うリポジトリ。 */
public interface MasterExerciseRepository extends JpaRepository<MasterExercise, Long> {

  Optional<MasterExercise> findBySlug(String slug);

  List<MasterExercise> findByIsPublishedTrueOrderByOrderIndexAsc();

  List<MasterExercise> findByIsPublishedTrueAndLanguageOrderByOrderIndexAsc(String language);
}
