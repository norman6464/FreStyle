package com.normanblog.frestyle.controller;

import static org.springframework.security.test.web.servlet.request.SecurityMockMvcRequestPostProcessors.jwt;
import static org.springframework.security.test.web.servlet.setup.SecurityMockMvcConfigurers.springSecurity;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import com.normanblog.frestyle.entity.Role;
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

/**
 * プロフィール画像アップロード URL 発行の自己参照(IDOR)と発行内容を検証する。
 *
 * <p>テスト環境は bucket 未設定のため stub presigner が使われる。
 */
@SpringBootTest
class ProfileImageControllerTest {

  private static final String SUB = "img-sub";

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

  @Test
  void issue_withoutAuth_returns401() throws Exception {
    mvc.perform(post("/api/v2/profile/me/image/presigned-url"))
        .andExpect(status().isUnauthorized());
  }

  @Test
  void issue_me_returnsUploadUrlForOwnKey() throws Exception {
    mvc.perform(
            post("/api/v2/profile/me/image/presigned-url")
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"fileName\":\"icon.png\",\"contentType\":\"image/png\"}"))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.uploadUrl").isNotEmpty())
        .andExpect(jsonPath("$.imageUrl").isNotEmpty())
        .andExpect(jsonPath("$.expiresIn").isNumber())
        // キーは自分の id 名前空間配下 + 拡張子を fileName から引き継ぐ。
        .andExpect(jsonPath("$.key").value(org.hamcrest.Matchers.startsWith("profiles/" + myId + "/")))
        .andExpect(jsonPath("$.key").value(org.hamcrest.Matchers.endsWith(".png")));
  }

  @Test
  void issue_withoutBody_stillSucceeds() throws Exception {
    // body 無しでも 400 にせず既定値で発行する。
    mvc.perform(post("/api/v2/profile/me/image/presigned-url").with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.key").value(org.hamcrest.Matchers.endsWith(".png")));
  }

  @Test
  void issue_otherUserId_returns403() throws Exception {
    mvc.perform(
            post("/api/v2/profile/" + (myId + 999) + "/image/presigned-url")
                .with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isForbidden());
  }
}
