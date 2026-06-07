package com.normanblog.frestyle.controller;

import com.normanblog.frestyle.dto.ExerciseSubmissionResponse;
import com.normanblog.frestyle.dto.ExerciseSubmitResult;
import com.normanblog.frestyle.dto.MasterExerciseDetailResponse;
import com.normanblog.frestyle.dto.MasterExerciseWithStatusResponse;
import com.normanblog.frestyle.dto.SubmitExerciseRequest;
import com.normanblog.frestyle.security.CurrentUserProvider;
import com.normanblog.frestyle.service.ExerciseService;
import com.normanblog.frestyle.service.SubmitExerciseUseCase;
import jakarta.validation.Valid;
import java.util.List;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

/**
 * master 演習 API。一覧(current user の status 付き)・詳細・提出(サーバ側採点)・提出履歴を公開する。
 *
 * <p>採点ロジックは {@link SubmitExerciseUseCase} に置きサーバ側で実行する(testable)。
 */
@RestController
@RequestMapping("/api/v2/exercises")
public class ExerciseController {

  private final ExerciseService exerciseService;
  private final SubmitExerciseUseCase submitExercise;
  private final CurrentUserProvider currentUser;

  public ExerciseController(
      ExerciseService exerciseService,
      SubmitExerciseUseCase submitExercise,
      CurrentUserProvider currentUser) {
    this.exerciseService = exerciseService;
    this.submitExercise = submitExercise;
    this.currentUser = currentUser;
  }

  @GetMapping
  public List<MasterExerciseWithStatusResponse> list(
      @RequestParam(value = "language", required = false) String language) {
    Long userId = currentUser.require().getId();
    return exerciseService.listWithStatus(userId, language);
  }

  @GetMapping("/{slug}")
  public MasterExerciseDetailResponse detail(@PathVariable String slug) {
    currentUser.require();
    return exerciseService.getDetail(slug);
  }

  @PostMapping("/{slug}/submit")
  public ExerciseSubmitResult submit(
      @PathVariable String slug, @Valid @RequestBody SubmitExerciseRequest request) {
    Long userId = currentUser.require().getId();
    return submitExercise.execute(userId, slug, request.code());
  }

  @GetMapping("/{slug}/submissions")
  public List<ExerciseSubmissionResponse> submissions(@PathVariable String slug) {
    Long userId = currentUser.require().getId();
    return exerciseService.listSubmissions(userId, slug);
  }
}
