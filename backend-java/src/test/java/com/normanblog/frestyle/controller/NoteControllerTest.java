package com.normanblog.frestyle.controller;

import static org.hamcrest.Matchers.hasItem;
import static org.springframework.security.test.web.servlet.request.SecurityMockMvcRequestPostProcessors.jwt;
import static org.springframework.security.test.web.servlet.setup.SecurityMockMvcConfigurers.springSecurity;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.repository.UserRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.context.WebApplicationContext;

/** ノート API のHTTP振る舞いとヘルスチェックを、認証込みで検証する。 */
@SpringBootTest
class NoteControllerTest {

  private static final String SUB = "note-test-sub";

  @Autowired private WebApplicationContext context;
  @Autowired private UserRepository users;

  private MockMvc mvc;

  @BeforeEach
  void setUp() {
    // @SpringBootTest は H2 をテスト間で共有するため、毎回クリアして固定ユーザーを seed する。
    users.deleteAll();
    users.save(User.builder().cognitoSub(SUB).email("u@example.com").role("trainee").build());
    mvc = MockMvcBuilders.webAppContextSetup(context).apply(springSecurity()).build();
  }

  @Test
  void health_isPublic_returnsUp() throws Exception {
    mvc.perform(get("/api/v2/health"))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.status").value("UP"));
  }

  @Test
  void notes_withoutAuth_returns401() throws Exception {
    mvc.perform(get("/api/v2/notes")).andExpect(status().isUnauthorized());
  }

  @Test
  void createThenList_asAuthenticatedUser() throws Exception {
    mvc.perform(
            post("/api/v2/notes")
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"title\":\"テストノート\",\"content\":\"本文\"}"))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.title").value("テストノート"))
        .andExpect(jsonPath("$.isPublic").value(false))
        .andExpect(jsonPath("$.isPinned").value(false));

    mvc.perform(get("/api/v2/notes").with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$[*].title", hasItem("テストノート")));
  }

  @Test
  void create_blankTitle_returns400() throws Exception {
    mvc.perform(
            post("/api/v2/notes")
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"title\":\"\",\"content\":\"本文\"}"))
        .andExpect(status().isBadRequest());
  }
}
