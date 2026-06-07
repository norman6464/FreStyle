package com.normanblog.frestyle.controller;

import static org.springframework.security.test.web.servlet.request.SecurityMockMvcRequestPostProcessors.jwt;
import static org.springframework.security.test.web.servlet.setup.SecurityMockMvcConfigurers.springSecurity;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import com.normanblog.frestyle.entity.AiChatSession;
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

/**
 * AI チャット SSE の事前検証(同期で返る 401 / 400 / 403)を検証する。SSE 本体(200)は
 * {@link com.normanblog.frestyle.service.SendAiMessageUseCaseTest} でカバーする。
 */
@SpringBootTest
class AiChatStreamControllerTest {

  private static final String SUB = "stream-sub";
  private static final String URL = "/api/v2/ai-chat/stream";

  @Autowired private WebApplicationContext context;
  @Autowired private UserRepository users;
  @Autowired private AiChatSessionRepository sessions;

  private MockMvc mvc;
  private Long otherId;

  @BeforeEach
  void setUp() {
    sessions.deleteAll();
    users.deleteAll();
    users.save(User.builder().cognitoSub(SUB).email("me@example.com").role(Role.TRAINEE).build());
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

  @Test
  void stream_withoutAuth_returns401() throws Exception {
    mvc.perform(
            post(URL)
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"sessionId\":0,\"content\":\"hi\"}"))
        .andExpect(status().isUnauthorized());
  }

  @Test
  void stream_emptyContentAndNoAttachments_returns400() throws Exception {
    mvc.perform(
            post(URL)
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"sessionId\":0,\"content\":\"\"}"))
        .andExpect(status().isBadRequest());
  }

  @Test
  void stream_othersSession_returns403() throws Exception {
    AiChatSession others =
        sessions.save(
            AiChatSession.builder()
                .userId(otherId)
                .title("他人")
                .sessionType("free")
                .createdAt(Instant.now())
                .updatedAt(Instant.now())
                .build());

    mvc.perform(
            post(URL)
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"sessionId\":" + others.getId() + ",\"content\":\"hi\"}"))
        .andExpect(status().isForbidden());
  }

  @Test
  void stream_attachmentWrongKeyPrefix_returns400() throws Exception {
    // 他ユーザーの prefix を指定 → サーバ側で読まされる前に弾く。
    String body =
        "{\"sessionId\":0,\"content\":\"hi\",\"attachments\":[{\"key\":\"ai-chat/"
            + otherId
            + "/x.png\",\"filename\":\"x.png\",\"contentType\":\"image/png\",\"sizeBytes\":100}]}";
    mvc.perform(
            post(URL)
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content(body))
        .andExpect(status().isBadRequest());
  }
}
