package com.normanblog.frestyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.springframework.security.test.web.servlet.request.SecurityMockMvcRequestPostProcessors.jwt;
import static org.springframework.security.test.web.servlet.setup.SecurityMockMvcConfigurers.springSecurity;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.put;
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

/** プロフィール API の自己参照(IDOR)・更新・onboarding を検証する。 */
@SpringBootTest
class ProfileControllerTest {

  private static final String SUB = "prof-sub";

  @Autowired private WebApplicationContext context;
  @Autowired private UserRepository users;

  private MockMvc mvc;
  private Long myId;

  @BeforeEach
  void setUp() {
    users.deleteAll();
    User me =
        users.save(
            User.builder()
                .cognitoSub(SUB)
                .email("me@example.com")
                .displayName("Me")
                .role(Role.TRAINEE)
                .build());
    myId = me.getId();
    mvc = MockMvcBuilders.webAppContextSetup(context).apply(springSecurity()).build();
  }

  @Test
  void get_withoutAuth_returns401() throws Exception {
    mvc.perform(get("/api/v2/profile/me")).andExpect(status().isUnauthorized());
  }

  @Test
  void get_me_returnsOwnProfile() throws Exception {
    mvc.perform(get("/api/v2/profile/me").with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.userId").value(myId))
        .andExpect(jsonPath("$.displayName").value("Me"))
        .andExpect(jsonPath("$.email").value("me@example.com"));
  }

  @Test
  void get_otherUserId_returns403() throws Exception {
    mvc.perform(get("/api/v2/profile/" + (myId + 999)).with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isForbidden());
  }

  @Test
  void update_savesDisplayNameAndProfileFields() throws Exception {
    mvc.perform(
            put("/api/v2/profile/me")
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"displayName\":\"新しい名前\",\"bio\":\"自己紹介\",\"status\":\"学習中\"}"))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.displayName").value("新しい名前"))
        .andExpect(jsonPath("$.bio").value("自己紹介"))
        .andExpect(jsonPath("$.status").value("学習中"));

    assertThat(users.findById(myId).orElseThrow().getDisplayName()).isEqualTo("新しい名前");
  }

  @Test
  void update_omittedDisplayName_isPreserved() throws Exception {
    // displayName を省略 → 既存の "Me" を保持し、bio だけ更新する。
    mvc.perform(
            put("/api/v2/profile/me/update")
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"bio\":\"bioのみ\"}"))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.displayName").value("Me"))
        .andExpect(jsonPath("$.bio").value("bioのみ"));
  }

  @Test
  void onboarding_setsOnboardedAt_idempotent() throws Exception {
    mvc.perform(post("/api/v2/profile/me/onboarding/complete").with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isNoContent());
    assertThat(users.findById(myId).orElseThrow().getOnboardedAt()).isNotNull();

    var first = users.findById(myId).orElseThrow().getOnboardedAt();
    mvc.perform(post("/api/v2/profile/me/onboarding/complete").with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isNoContent());
    // 二度押しでも初回日時を保持する(冪等)。
    assertThat(users.findById(myId).orElseThrow().getOnboardedAt()).isEqualTo(first);
  }
}
