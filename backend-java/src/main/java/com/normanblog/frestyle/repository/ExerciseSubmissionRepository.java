package com.normanblog.frestyle.repository;

import com.normanblog.frestyle.entity.ExerciseSubmission;
import java.util.List;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

/** exercise_submissions テーブルへのアクセスを担うリポジトリ。 */
public interface ExerciseSubmissionRepository extends JpaRepository<ExerciseSubmission, Long> {

  // ユーザーの特定演習への提出履歴(新しい順)。
  List<ExerciseSubmission> findByUserIdAndExerciseKindAndExerciseIdOrderBySubmittedAtDesc(
      Long userId, String exerciseKind, Long exerciseId);

  // ユーザーが提出した master 演習 id(着手済み判定用)。
  @Query(
      "select distinct s.exerciseId from ExerciseSubmission s "
          + "where s.userId = :userId and s.exerciseKind = 'master'")
  List<Long> findAttemptedMasterExerciseIds(@Param("userId") Long userId);

  // ユーザーが正答した master 演習 id(solved 判定用)。
  @Query(
      "select distinct s.exerciseId from ExerciseSubmission s "
          + "where s.userId = :userId and s.exerciseKind = 'master' and s.isCorrect = true")
  List<Long> findSolvedMasterExerciseIds(@Param("userId") Long userId);

  // 演習ごとの提出総数 [exerciseId, count]。
  @Query(
      "select s.exerciseId, count(s) from ExerciseSubmission s "
          + "where s.exerciseKind = 'master' group by s.exerciseId")
  List<Object[]> aggregateMasterTotalSubmissions();

  // 演習ごとの正答ユーザー数 [exerciseId, count(distinct userId)]。
  @Query(
      "select s.exerciseId, count(distinct s.userId) from ExerciseSubmission s "
          + "where s.exerciseKind = 'master' and s.isCorrect = true group by s.exerciseId")
  List<Object[]> aggregateMasterSolvedUsers();
}
