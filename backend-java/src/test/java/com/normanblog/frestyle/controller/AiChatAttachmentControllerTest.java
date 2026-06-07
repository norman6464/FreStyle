package com.normanblog.frestyle.controller;

import static org.springframework.security.test.web.servlet.request.SecurityMockMvcRequestPostProcessors.jwt;
import static org.springframework.security.test.web.servlet.setup.SecurityMockMvcConfigurers.springSecurity;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import com.normanblog.frestyle.entity.Role;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.repository.UserRepository;
import org.hamcrest.Matchers;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.context.WebApplicationContext;

/**
 * AI チャット添付の署名 URL 発行を検証する。テストは bucket 未設定のため stub presigner が使われる。
 */
@SpringBootTest
class AiChatAttachmentControllerTest {

  private static final String SUB = "att-sub";
  private static final String URL = "/api/v2/ai-chat/attachments/upload-url";

  @Autowired private WebApplicationContext context;
  @Autowired private UserRepository users;

  private MockMvc mvc;
  private Long myId;

  @BeforeEach
  void setUp() {
    users.deleteAll();
    myId =
        users
            .save(User.builder().cognitoSub(SUB).email("me@example.com").role(Role.TRAINEE).build())
            .getId();
    mvc = MockMvcBuilders.webAppContextSetup(context).apply(springSecurity()).build();
  }

  private String body(String filename, String contentType, long sizeBytes) {
    return "{\"filename\":\"%s\",\"contentType\":\"%s\",\"sizeBytes\":%d}"
        .formatted(filename, contentType, sizeBytes);
  }

  @Test
  void issue_withoutAuth_returns401() throws Exception {
    mvc.perform(post(URL).contentType(MediaType.APPLICATION_JSON).content(body("a.png", "image/png", 100)))
        .andExpect(status().isUnauthorized());
  }

  @Test
  void issue_validImage_returnsUploadUrlWithOwnKeyspace() throws Exception {
    mvc.perform(
            post(URL)
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content(body("icon.png", "image/png", 1024)))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.uploadUrl").isNotEmpty())
        .andExpect(jsonPath("$.expiresIn").isNumber())
        .andExpect(jsonPath("$.key").value(Matchers.startsWith("ai-chat/" + myId + "/")))
        .andExpect(jsonPath("$.key").value(Matchers.endsWith(".png")));
  }

  @Test
  void issue_unsupportedType_returns415() throws Exception {
    mvc.perform(
            post(URL)
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content(body("doc.pdf", "application/pdf", 1024)))
        .andExpect(status().isUnsupportedMediaType());
  }

  @Test
  void issue_tooLarge_returns413() throws Exception {
    long sixMB = 6L * 1024 * 1024;
    mvc.perform(
            post(URL)
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content(body("big.png", "image/png", sixMB)))
        .andExpect(status().is(413));
  }

  @Test
  void issue_missingFields_returns400() throws Exception {
    mvc.perform(
            post(URL)
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"filename\":\"\",\"contentType\":\"\",\"sizeBytes\":0}"))
        .andExpect(status().isBadRequest());
  }
}
