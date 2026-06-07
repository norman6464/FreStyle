package com.normanblog.frestyle.controller;

import static org.junit.jupiter.api.Assumptions.assumeTrue;
import static org.springframework.security.test.web.servlet.request.SecurityMockMvcRequestPostProcessors.jwt;
import static org.springframework.security.test.web.servlet.setup.SecurityMockMvcConfigurers.springSecurity;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import com.normanblog.frestyle.entity.Role;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.repository.UserRepository;
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

/** コード実行(サンドボックス)の認証境界・言語検証・実 Java 実行を検証する。 */
@SpringBootTest
class CodeExecuteControllerTest {

  private static final String SUB = "exec-sub";
  private static final String URL = "/api/v2/code/execute";

  @Autowired private WebApplicationContext context;
  @Autowired private UserRepository users;

  private MockMvc mvc;

  @BeforeEach
  void setUp() {
    users.deleteAll();
    users.save(User.builder().cognitoSub(SUB).email("me@example.com").role(Role.TRAINEE).build());
    mvc = MockMvcBuilders.webAppContextSetup(context).apply(springSecurity()).build();
  }

  // CI 等で java が PATH に無い場合は実行系テストを skip する。
  private static boolean javaAvailable() {
    try {
      Process p = new ProcessBuilder("java", "--version").start();
      return p.waitFor(10, TimeUnit.SECONDS) && p.exitValue() == 0;
    } catch (Exception e) {
      return false;
    }
  }

  @Test
  void execute_withoutAuth_returns401() throws Exception {
    mvc.perform(
            post(URL)
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"code\":\"x\",\"language\":\"java\"}"))
        .andExpect(status().isUnauthorized());
  }

  @Test
  void execute_unsupportedLanguage_returns400() throws Exception {
    mvc.perform(
            post(URL)
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"code\":\"print('hi')\",\"language\":\"python\"}"))
        .andExpect(status().isBadRequest());
  }

  @Test
  void execute_validJava_runsAndReturnsStdout() throws Exception {
    assumeTrue(javaAvailable(), "java が PATH に無いため skip");
    String code =
        "public class Main { public static void main(String[] a) { System.out.println(\"Hello FreStyle\"); } }";
    mvc.perform(
            post(URL)
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"code\":" + jsonString(code) + ",\"language\":\"java\"}"))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.exitCode").value(0))
        .andExpect(jsonPath("$.stdout").value(Matchers.containsString("Hello FreStyle")));
  }

  @Test
  void execute_nonCompilingJava_returnsNonZeroExit() throws Exception {
    assumeTrue(javaAvailable(), "java が PATH に無いため skip");
    String code = "public class Main { this does not compile }";
    mvc.perform(
            post(URL)
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"code\":" + jsonString(code) + ",\"language\":\"java\"}"))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.exitCode").value(Matchers.not(0)))
        .andExpect(jsonPath("$.stderr").value(Matchers.not(Matchers.emptyString())));
  }

  // 文字列を JSON 文字列リテラルに変換する(エスケープ込み)。
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
