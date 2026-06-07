package com.normanblog.frestyle.repository;

import com.normanblog.frestyle.entity.MasterExerciseExample;
import java.util.List;
import org.springframework.data.jpa.repository.JpaRepository;

/** master_exercise_examples テーブルへのアクセスを担うリポジトリ。 */
public interface MasterExerciseExampleRepository
    extends JpaRepository<MasterExerciseExample, Long> {

  List<MasterExerciseExample> findByExerciseIdOrderByOrderIndexAsc(Long exerciseId);
}
