package com.normanblog.frestyle.controller;

import static org.springframework.security.test.web.servlet.request.SecurityMockMvcRequestPostProcessors.jwt;
import static org.springframework.security.test.web.servlet.setup.SecurityMockMvcConfigurers.springSecurity;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.repository.UserRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.context.WebApplicationContext;

/** GET /api/v2/auth/me の認証・ユーザー解決・admin 同期を検証する。 */
@SpringBootTest
class AuthControllerTest {

  private static final String SUB = "me-test-sub";

  @Autowired private WebApplicationContext context;
  @Autowired private UserRepository users;

  private MockMvc mvc;

  @BeforeEach
  void setUp() {
    users.deleteAll();
    mvc = MockMvcBuilders.webAppContextSetup(context).apply(springSecurity()).build();
  }

  @Test
  void me_withoutAuth_returns401() throws Exception {
    mvc.perform(get("/api/v2/auth/me")).andExpect(status().isUnauthorized());
  }

  @Test
  void me_validJwtButNoUserRow_returns401() throws Exception {
    // JWT は正当でも users 行が無い(招待前)なら 401。
    mvc.perform(get("/api/v2/auth/me").with(jwt().jwt(j -> j.subject("unknown-sub"))))
        .andExpect(status().isUnauthorized());
  }

  @Test
  void me_returnsCurrentUser() throws Exception {
    users.save(
        User.builder().cognitoSub(SUB).email("u@example.com").displayName("U").role("trainee").build());

    mvc.perform(get("/api/v2/auth/me").with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.cognitoSub").value(SUB))
        .andExpect(jsonPath("$.email").value("u@example.com"))
        .andExpect(jsonPath("$.role").value("trainee"))
        .andExpect(jsonPath("$.isAdmin").value(false))
        .andExpect(jsonPath("$.onboarded").value(false));
  }

  @Test
  void me_adminGroup_promotesRoleToSuperAdmin() throws Exception {
    users.save(User.builder().cognitoSub(SUB).email("a@example.com").role("trainee").build());

    // cognito:groups に admin が含まれると super_admin に自動昇格する。
    mvc.perform(
            get("/api/v2/auth/me")
                .with(jwt().jwt(j -> j.subject(SUB).claim("cognito:groups", java.util.List.of("admin")))))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.isAdmin").value(true))
        .andExpect(jsonPath("$.role").value("super_admin"))
        .andExpect(jsonPath("$.groups[0]").value("admin"));
  }
}
