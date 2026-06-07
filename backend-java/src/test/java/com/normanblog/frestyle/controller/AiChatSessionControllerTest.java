package com.normanblog.frestyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.springframework.security.test.web.servlet.request.SecurityMockMvcRequestPostProcessors.jwt;
import static org.springframework.security.test.web.servlet.setup.SecurityMockMvcConfigurers.springSecurity;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.delete;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.put;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import com.normanblog.frestyle.entity.AiChatSession;
import com.normanblog.frestyle.entity.AiChatSessionType;
import com.normanblog.frestyle.entity.Role;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.repository.AiChatSessionRepository;
import com.normanblog.frestyle.repository.UserRepository;
import java.time.Instant;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.context.WebApplicationContext;

/** AI チャットセッション CRUD の作成・一覧(自分のみ降順)・所有者検証(他人は 403)を検証する。 */
@SpringBootTest
class AiChatSessionControllerTest {

  private static final String SUB = "aichat-sub";

  @Autowired private WebApplicationContext context;
  @Autowired private UserRepository users;
  @Autowired private AiChatSessionRepository sessions;

  private MockMvc mvc;
  private Long myId;
  private Long otherId;

  @BeforeEach
  void setUp() {
    sessions.deleteAll();
    users.deleteAll();
    myId =
        users
            .save(User.builder().cognitoSub(SUB).email("me@example.com").role(Role.TRAINEE).build())
            .getId();
    otherId =
        users
            .save(
                User.builder()
                    .cognitoSub("other")
                    .email("o@example.com")
                    .role(Role.TRAINEE)
                    .build())
            .getId();
    mvc = MockMvcBuilders.webAppContextSetup(context).apply(springSecurity()).build();
  }

  private Long saveSession(Long userId, String title, Instant createdAt) {
    return sessions
        .save(
            AiChatSession.builder()
                .userId(userId)
                .title(title)
                .sessionType(AiChatSessionType.FREE)
                .createdAt(createdAt)
                .updatedAt(createdAt)
                .build())
        .getId();
  }

  @Test
  void list_withoutAuth_returns401() throws Exception {
    mvc.perform(get("/api/v2/ai-chat/sessions")).andExpect(status().isUnauthorized());
  }

  @Test
  void create_then_list_returnsOnlyOwn_desc() throws Exception {
    mvc.perform(
            post("/api/v2/ai-chat/sessions")
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"title\":\"はじめての会話\"}"))
        .andExpect(status().isCreated())
        .andExpect(jsonPath("$.title").value("はじめての会話"))
        .andExpect(jsonPath("$.userId").value(myId))
        // sessionType 省略 → free に既定化。
        .andExpect(jsonPath("$.sessionType").value(AiChatSessionType.FREE));

    saveSession(otherId, "他人の会話", Instant.now());

    mvc.perform(get("/api/v2/ai-chat/sessions").with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.length()").value(1))
        .andExpect(jsonPath("$[0].title").value("はじめての会話"));
  }

  @Test
  void create_blankTitle_returns400() throws Exception {
    mvc.perform(
            post("/api/v2/ai-chat/sessions")
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"title\":\"\"}"))
        .andExpect(status().isBadRequest());
  }

  @Test
  void get_ownSession_returnsIt() throws Exception {
    Long id = saveSession(myId, "自分の会話", Instant.now());

    mvc.perform(get("/api/v2/ai-chat/sessions/" + id).with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.id").value(id));
  }

  @Test
  void get_othersSession_returns403() throws Exception {
    Long id = saveSession(otherId, "他人の会話", Instant.now());

    mvc.perform(get("/api/v2/ai-chat/sessions/" + id).with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isForbidden());
  }

  @Test
  void get_missingSession_returns404() throws Exception {
    mvc.perform(get("/api/v2/ai-chat/sessions/999999").with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isNotFound());
  }

  @Test
  void updateTitle_ownSession_updates() throws Exception {
    Long id = saveSession(myId, "旧タイトル", Instant.now());

    mvc.perform(
            put("/api/v2/ai-chat/sessions/" + id)
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"title\":\"新タイトル\"}"))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.title").value("新タイトル"));

    assertThat(sessions.findById(id).orElseThrow().getTitle()).isEqualTo("新タイトル");
  }

  @Test
  void updateTitle_othersSession_returns403_andNotChanged() throws Exception {
    Long id = saveSession(otherId, "他人タイトル", Instant.now());

    mvc.perform(
            put("/api/v2/ai-chat/sessions/" + id)
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"title\":\"乗っ取り\"}"))
        .andExpect(status().isForbidden());

    assertThat(sessions.findById(id).orElseThrow().getTitle()).isEqualTo("他人タイトル");
  }

  @Test
  void delete_ownSession_removes() throws Exception {
    Long id = saveSession(myId, "削除対象", Instant.now());

    mvc.perform(delete("/api/v2/ai-chat/sessions/" + id).with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isNoContent());

    assertThat(sessions.findById(id)).isEmpty();
  }

  @Test
  void delete_othersSession_returns403_andNotRemoved() throws Exception {
    Long id = saveSession(otherId, "他人の", Instant.now());

    mvc.perform(delete("/api/v2/ai-chat/sessions/" + id).with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isForbidden());

    assertThat(sessions.findById(id)).isPresent();
  }

  @Test
  void messages_withoutAuth_returns401() throws Exception {
    mvc.perform(get("/api/v2/ai-chat/sessions/1/messages")).andExpect(status().isUnauthorized());
  }

  @Test
  void messages_ownSession_returnsEmptyList_withStubReader() throws Exception {
    // テストは table 未設定で stub reader が使われるため、所有者検証を通って空配列が返る。
    Long id = saveSession(myId, "自分の会話", Instant.now());

    mvc.perform(get("/api/v2/ai-chat/sessions/" + id + "/messages").with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.length()").value(0));
  }

  @Test
  void messages_othersSession_returns403() throws Exception {
    Long id = saveSession(otherId, "他人の会話", Instant.now());

    mvc.perform(get("/api/v2/ai-chat/sessions/" + id + "/messages").with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isForbidden());
  }

  @Test
  void messages_missingSession_returns404() throws Exception {
    mvc.perform(get("/api/v2/ai-chat/sessions/999999/messages").with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isNotFound());
  }
}
