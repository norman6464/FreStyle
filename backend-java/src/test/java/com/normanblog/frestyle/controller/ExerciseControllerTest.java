package com.normanblog.frestyle.controller;

import static org.junit.jupiter.api.Assumptions.assumeTrue;
import static org.springframework.security.test.web.servlet.request.SecurityMockMvcRequestPostProcessors.jwt;
import static org.springframework.security.test.web.servlet.setup.SecurityMockMvcConfigurers.springSecurity;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import com.normanblog.frestyle.entity.ExerciseMode;
import com.normanblog.frestyle.entity.MasterExercise;
import com.normanblog.frestyle.entity.MasterExerciseExample;
import com.normanblog.frestyle.entity.Role;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.repository.ExerciseSubmissionRepository;
import com.normanblog.frestyle.repository.MasterExerciseExampleRepository;
import com.normanblog.frestyle.repository.MasterExerciseRepository;
import com.normanblog.frestyle.repository.UserRepository;
import java.time.Instant;
import java.util.concurrent.TimeUnit;
import org.hamcrest.Matchers;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.context.WebApplicationContext;

/** 演習 API の一覧(status 付き)・詳細・採点提出(qa / execute)・提出履歴を検証する。 */
@SpringBootTest
class ExerciseControllerTest {

  private static final String SUB = "exercise-sub";

  @Autowired private WebApplicationContext context;
  @Autowired private UserRepository users;
  @Autowired private MasterExerciseRepository exercises;
  @Autowired private MasterExerciseExampleRepository examples;
  @Autowired private ExerciseSubmissionRepository submissions;

  private MockMvc mvc;

  @BeforeEach
  void setUp() {
    submissions.deleteAll();
    examples.deleteAll();
    exercises.deleteAll();
    users.deleteAll();
    users.save(User.builder().cognitoSub(SUB).email("me@example.com").role(Role.TRAINEE).build());
    mvc = MockMvcBuilders.webAppContextSetup(context).apply(springSecurity()).build();
  }

  private MasterExercise seedQaExercise() {
    return exercises.save(
        MasterExercise.builder()
            .slug("java-qa-1")
            .language("java")
            .orderIndex(1)
            .category("文法")
            .title("変数宣言")
            .description("int を宣言する")
            .starterCode("")
            .expectedOutput("int x = 1;")
            .mode(ExerciseMode.QA)
            .explanation("解説")
            .difficulty((short) 1)
            .isPublished(true)
            .createdAt(Instant.now())
            .updatedAt(Instant.now())
            .build());
  }

  private MasterExercise seedExecuteExercise() {
    MasterExercise e =
        exercises.save(
            MasterExercise.builder()
                .slug("java-exec-1")
                .language("java")
                .orderIndex(2)
                .category("入出力")
                .title("標準入力をそのまま出力")
                .description("入力された行を出力する")
                .starterCode("")
                .expectedOutput("hello")
                .mode(ExerciseMode.EXECUTE)
                .explanation("解説")
                .difficulty((short) 1)
                .isPublished(true)
                .createdAt(Instant.now())
                .updatedAt(Instant.now())
                .build());
    examples.save(
        MasterExerciseExample.builder()
            .exerciseId(e.getId())
            .orderIndex((short) 1)
            .inputText("hello\n")
            .expectedOutput("hello")
            .createdAt(Instant.now())
            .updatedAt(Instant.now())
            .build());
    return e;
  }

  @Test
  void list_withoutAuth_returns401() throws Exception {
    mvc.perform(get("/api/v2/exercises")).andExpect(status().isUnauthorized());
  }

  @Test
  void list_returnsPublishedWithStatus() throws Exception {
    seedQaExercise();
    mvc.perform(get("/api/v2/exercises").with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$[0].slug").value("java-qa-1"))
        .andExpect(jsonPath("$[0].status").value(""))
        .andExpect(jsonPath("$[0].stats.totalSubmissions").value(0));
  }

  @Test
  void list_filtersByLanguage() throws Exception {
    seedQaExercise();
    mvc.perform(
            get("/api/v2/exercises").param("language", "go").with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$").isEmpty());
  }

  @Test
  void detail_returnsExerciseAndExamples() throws Exception {
    seedExecuteExercise();
    mvc.perform(get("/api/v2/exercises/java-exec-1").with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.exercise.slug").value("java-exec-1"))
        .andExpect(jsonPath("$.examples[0].expectedOutput").value("hello"));
  }

  @Test
  void detail_unknownSlug_returns404() throws Exception {
    mvc.perform(get("/api/v2/exercises/nope").with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isNotFound());
  }

  @Test
  void submit_qaMode_correctAndIncorrect() throws Exception {
    seedQaExercise();
    // 期待値と一致 → 正答。
    mvc.perform(
            post("/api/v2/exercises/java-qa-1/submit")
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"code\":\"int x = 1;\\n\"}"))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.isCorrect").value(true));

    // 不一致 → 誤答。
    mvc.perform(
            post("/api/v2/exercises/java-qa-1/submit")
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"code\":\"int y = 2;\"}"))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.isCorrect").value(false));
  }

  @Test
  void submit_emptyCode_returns400() throws Exception {
    seedQaExercise();
    mvc.perform(
            post("/api/v2/exercises/java-qa-1/submit")
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"code\":\"\"}"))
        .andExpect(status().isBadRequest());
  }

  @Test
  void submissions_returnsHistoryNewestFirst() throws Exception {
    seedQaExercise();
    // 2 件提出し、後から送った方が先頭(新しい順)に来ることまで検証する。
    mvc.perform(
            post("/api/v2/exercises/java-qa-1/submit")
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"code\":\"int x = 1;\"}"))
        .andExpect(status().isOk());
    mvc.perform(
            post("/api/v2/exercises/java-qa-1/submit")
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"code\":\"int x = 2;\"}"))
        .andExpect(status().isOk());

    mvc.perform(get("/api/v2/exercises/java-qa-1/submissions").with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$[0].submittedCode").value("int x = 2;"))
        .andExpect(jsonPath("$[1].submittedCode").value("int x = 1;"));
  }

  @Test
  void submit_executeMode_runsJavaAndGrades() throws Exception {
    assumeTrue(available("java", "--version"), "java が PATH に無いため skip");
    seedExecuteExercise();
    // 標準入力 1 行を読んでそのまま出力する Java。example の inputText="hello" → stdout "hello"。
    String code =
        "import java.util.Scanner;"
            + "public class Main { public static void main(String[] a) {"
            + " Scanner s = new Scanner(System.in); System.out.println(s.nextLine()); } }";
    mvc.perform(
            post("/api/v2/exercises/java-exec-1/submit")
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"code\":" + jsonString(code) + "}"))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.isCorrect").value(true))
        .andExpect(jsonPath("$.results[0].passed").value(true))
        .andExpect(jsonPath("$.results[0].actualOutput").value(Matchers.containsString("hello")));
  }

  // CI 等で対象言語のランタイムが PATH に無い場合は実行系テストを skip する。
  private static boolean available(String... command) {
    try {
      Process p = new ProcessBuilder(command).start();
      if (!p.waitFor(10, TimeUnit.SECONDS)) {
        p.destroyForcibly();
        return false;
      }
      return p.exitValue() == 0;
    } catch (Exception e) {
      return false;
    }
  }

  private static String jsonString(String s) {
    StringBuilder sb = new StringBuilder("\"");
    for (char c : s.toCharArray()) {
      switch (c) {
        case '"' -> sb.append("\\\"");
        case '\\' -> sb.append("\\\\");
        case '\n' -> sb.append("\\n");
        default -> sb.append(c);
      }
    }
    return sb.append('"').toString();
  }
}
