package com.normanblog.frestyle.service;

import com.normanblog.frestyle.dto.CodeExecuteResponse;
import com.normanblog.frestyle.dto.ExerciseSubmitResult;
import com.normanblog.frestyle.dto.ExerciseSubmitResult.TestCaseResult;
import com.normanblog.frestyle.entity.ExerciseKind;
import com.normanblog.frestyle.entity.ExerciseMode;
import com.normanblog.frestyle.entity.ExerciseSubmission;
import com.normanblog.frestyle.entity.MasterExercise;
import com.normanblog.frestyle.entity.MasterExerciseExample;
import com.normanblog.frestyle.infra.exec.CodeExecutor;
import com.normanblog.frestyle.repository.ExerciseSubmissionRepository;
import com.normanblog.frestyle.repository.MasterExerciseExampleRepository;
import com.normanblog.frestyle.repository.MasterExerciseRepository;
import java.time.Instant;
import java.util.ArrayList;
import java.util.List;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Component;
import org.springframework.web.server.ResponseStatusException;

/**
 * master 演習の提出を採点し履歴に保存するユースケース。採点ロジックをサーバ側に置く(testable)。
 *
 * <p>mode で分岐: qa はコード実行せず提出文字列を比較、execute は examples を全件実行して stdout 比較。
 * コード実行・採点・保存と複数の関心を跨ぐため専用 UseCase に切り出す。
 */
@Component
public class SubmitExerciseUseCase {

  private final MasterExerciseRepository exercises;
  private final MasterExerciseExampleRepository examples;
  private final ExerciseSubmissionRepository submissions;
  private final CodeExecutor executor;

  public SubmitExerciseUseCase(
      MasterExerciseRepository exercises,
      MasterExerciseExampleRepository examples,
      ExerciseSubmissionRepository submissions,
      CodeExecutor executor) {
    this.exercises = exercises;
    this.examples = examples;
    this.submissions = submissions;
    this.executor = executor;
  }

  /**
   * slug の演習に対してコードを採点し、提出履歴を保存して結果を返す。
   *
   * <p>採点はコード実行(外部プロセス / ブロッキング)を伴うため、ここに DB トランザクションを張らない。
   * 各リポジトリ呼び出しが自前の短いトランザクションで完結し、実行中に接続を握り続けないようにする。
   */
  public ExerciseSubmitResult execute(Long userId, String slug, String code) {
    MasterExercise exercise =
        exercises
            .findBySlug(slug)
            .filter(MasterExercise::isPublished)
            .orElseThrow(
                () -> new ResponseStatusException(HttpStatus.NOT_FOUND, "演習問題が見つかりません"));

    if (ExerciseMode.QA.equals(exercise.getMode())) {
      return submitQa(userId, exercise, code);
    }

    return submitExecute(userId, exercise, code);
  }

  // qa: 実行せず提出文字列と expectedOutput を正規化比較する。
  private ExerciseSubmitResult submitQa(Long userId, MasterExercise exercise, String code) {
    boolean correct = ExerciseGrading.matches(code, exercise.getExpectedOutput());
    ExerciseSubmission saved =
        save(userId, exercise.getId(), code, code, "", 0, correct);

    return new ExerciseSubmitResult(
        saved.getId(),
        correct,
        List.of(
            new TestCaseResult(
                (short) 1, "", exercise.getExpectedOutput(), code, "", correct)));
  }

  // execute: examples を全件実行し stdout を比較する(全件 pass かつ exit 0 で正答)。
  private ExerciseSubmitResult submitExecute(Long userId, MasterExercise exercise, String code) {
    List<MasterExerciseExample> cases = examples.findByExerciseIdOrderByOrderIndexAsc(exercise.getId());
    if (cases.isEmpty()) {
      // examples が無い演習は exercise 自身の expectedOutput を単一の仮想ケースにする。
      MasterExerciseExample virtual =
          MasterExerciseExample.builder()
              .exerciseId(exercise.getId())
              .orderIndex((short) 1)
              .inputText("")
              .expectedOutput(exercise.getExpectedOutput() == null ? "" : exercise.getExpectedOutput())
              .build();
      cases = List.of(virtual);
    }

    List<TestCaseResult> results = new ArrayList<>(cases.size());
    boolean allPassed = true;
    String repStdout = "";
    String repStderr = "";
    int repExit = 0;

    for (MasterExerciseExample tc : cases) {
      CodeExecuteResponse out = executor.execute(exercise.getLanguage(), code, tc.getInputText());
      boolean passed = out.exitCode() == 0 && ExerciseGrading.matches(out.stdout(), tc.getExpectedOutput());
      results.add(
          new TestCaseResult(
              tc.getOrderIndex(),
              tc.getInputText(),
              tc.getExpectedOutput(),
              out.stdout(),
              out.stderr(),
              passed));

      // 代表結果: 最初の失敗で固定、全件 pass の間は最後の出力で上書き。
      if (!passed && allPassed) {
        repStdout = out.stdout();
        repStderr = out.stderr();
        repExit = out.exitCode();
        allPassed = false;
      } else if (allPassed) {
        repStdout = out.stdout();
        repStderr = out.stderr();
        repExit = out.exitCode();
      }
    }

    ExerciseSubmission saved =
        save(userId, exercise.getId(), code, repStdout, repStderr, repExit, allPassed);

    return new ExerciseSubmitResult(saved.getId(), allPassed, results);
  }

  private ExerciseSubmission save(
      Long userId, Long exerciseId, String code, String stdout, String stderr, int exit, boolean correct) {
    return submissions.save(
        ExerciseSubmission.builder()
            .userId(userId)
            .exerciseKind(ExerciseKind.MASTER)
            .exerciseId(exerciseId)
            .submittedCode(code)
            .stdout(stdout)
            .stderr(stderr)
            .exitCode(exit)
            .isCorrect(correct)
            .submittedAt(Instant.now())
            .build());
  }
}
