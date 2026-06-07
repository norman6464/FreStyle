package com.normanblog.frestyle.service;

import com.normanblog.frestyle.dto.MasterExerciseDetailResponse;
import com.normanblog.frestyle.dto.MasterExerciseExampleResponse;
import com.normanblog.frestyle.dto.MasterExerciseResponse;
import com.normanblog.frestyle.dto.MasterExerciseWithStatusResponse;
import com.normanblog.frestyle.entity.ExerciseKind;
import com.normanblog.frestyle.entity.MasterExercise;
import com.normanblog.frestyle.repository.ExerciseSubmissionRepository;
import com.normanblog.frestyle.repository.MasterExerciseExampleRepository;
import com.normanblog.frestyle.repository.MasterExerciseRepository;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.StringUtils;
import org.springframework.web.server.ResponseStatusException;

/**
 * 演習の一覧・詳細・提出履歴を提供する読み取り系サービス(採点・提出は {@link SubmitExerciseUseCase})。
 *
 * <p>一覧の status / stats は batch クエリで取得し N+1 を避ける。
 */
@Service
@Transactional(readOnly = true)
public class ExerciseService {

  private final MasterExerciseRepository exercises;
  private final MasterExerciseExampleRepository examples;
  private final ExerciseSubmissionRepository submissions;

  public ExerciseService(
      MasterExerciseRepository exercises,
      MasterExerciseExampleRepository examples,
      ExerciseSubmissionRepository submissions) {
    this.exercises = exercises;
    this.examples = examples;
    this.submissions = submissions;
  }

  /** 公開演習を language で絞り込んで一覧する。userId が null なら status は全て未着手("")。 */
  public List<MasterExerciseWithStatusResponse> listWithStatus(Long userId, String language) {
    List<MasterExercise> list =
        StringUtils.hasText(language)
            ? exercises.findByIsPublishedTrueAndLanguageOrderByOrderIndexAsc(language)
            : exercises.findByIsPublishedTrueOrderByOrderIndexAsc();

    Set<Long> solvedIds = new HashSet<>();
    Set<Long> attemptedIds = new HashSet<>();
    if (userId != null) {
      solvedIds.addAll(submissions.findSolvedMasterExerciseIds(userId));
      attemptedIds.addAll(submissions.findAttemptedMasterExerciseIds(userId));
    }

    Map<Long, Long> totalMap = toCountMap(submissions.aggregateMasterTotalSubmissions());
    Map<Long, Long> solvedUsersMap = toCountMap(submissions.aggregateMasterSolvedUsers());

    return list.stream()
        .map(
            e -> {
              String status =
                  solvedIds.contains(e.getId())
                      ? "solved"
                      : attemptedIds.contains(e.getId()) ? "in_progress" : "";
              long total = totalMap.getOrDefault(e.getId(), 0L);
              long solvedUsers = solvedUsersMap.getOrDefault(e.getId(), 0L);
              return MasterExerciseWithStatusResponse.of(e, status, total, solvedUsers);
            })
        .toList();
  }

  /** slug の演習詳細(本体 + 例)を返す。未公開・不存在は 404。 */
  public MasterExerciseDetailResponse getDetail(String slug) {
    MasterExercise exercise =
        exercises
            .findBySlug(slug)
            .filter(MasterExercise::isPublished)
            .orElseThrow(
                () -> new ResponseStatusException(HttpStatus.NOT_FOUND, "演習問題が見つかりません"));

    List<MasterExerciseExampleResponse> exampleResponses =
        examples.findByExerciseIdOrderByOrderIndexAsc(exercise.getId()).stream()
            .map(MasterExerciseExampleResponse::from)
            .toList();

    return new MasterExerciseDetailResponse(
        MasterExerciseResponse.from(exercise), exampleResponses);
  }

  /** ユーザーの該当演習への提出履歴(新しい順)。 */
  public List<com.normanblog.frestyle.dto.ExerciseSubmissionResponse> listSubmissions(
      Long userId, String slug) {
    MasterExercise exercise =
        exercises
            .findBySlug(slug)
            .orElseThrow(
                () -> new ResponseStatusException(HttpStatus.NOT_FOUND, "演習問題が見つかりません"));

    return submissions
        .findByUserIdAndExerciseKindAndExerciseIdOrderBySubmittedAtDesc(
            userId, ExerciseKind.MASTER, exercise.getId())
        .stream()
        .map(com.normanblog.frestyle.dto.ExerciseSubmissionResponse::from)
        .toList();
  }

  // [exerciseId, count] の Object[] 列を Map<Long,Long> に畳む。
  private static Map<Long, Long> toCountMap(List<Object[]> rows) {
    Map<Long, Long> map = new HashMap<>();
    for (Object[] row : rows) {
      map.put(((Number) row[0]).longValue(), ((Number) row[1]).longValue());
    }
    return map;
  }
}
