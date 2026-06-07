package com.normanblog.frestyle.controller;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;
import static org.springframework.security.test.web.servlet.setup.SecurityMockMvcConfigurers.springSecurity;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.cookie;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.normanblog.frestyle.infra.cognito.CognitoTokenClient;
import com.normanblog.frestyle.infra.cognito.CognitoTokens;
import com.normanblog.frestyle.infra.cognito.TokenExchangeException;
import com.normanblog.frestyle.entity.AdminInvitation;
import com.normanblog.frestyle.entity.InvitationStatus;
import com.normanblog.frestyle.entity.Role;
import com.normanblog.frestyle.repository.AdminInvitationRepository;
import com.normanblog.frestyle.repository.UserRepository;
import java.time.Instant;
import java.time.temporal.ChronoUnit;
import java.util.Base64;
import java.util.List;
import java.util.Map;
import jakarta.servlet.http.Cookie;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.MediaType;
import org.springframework.test.context.bean.override.mockito.MockitoBean;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.context.WebApplicationContext;

/** Cognito ログインフロー(callback / logout / refresh)を、token client をモックして検証する。 */
@SpringBootTest
class AuthCognitoFlowControllerTest {

  private static final ObjectMapper MAPPER = new ObjectMapper();

  @Autowired private WebApplicationContext context;
  @Autowired private UserRepository users;
  @Autowired private AdminInvitationRepository invitations;
  @MockitoBean private CognitoTokenClient tokenClient;

  private MockMvc mvc;

  @BeforeEach
  void setUp() {
    users.deleteAll();
    invitations.deleteAll();
    mvc = MockMvcBuilders.webAppContextSetup(context).apply(springSecurity()).build();
  }

  // 署名検証しない id_token を模した JWT(header.payload.)を組み立てる。
  private String fakeIdToken(String sub, String email, List<String> groups) throws Exception {
    String header = b64(Map.of("alg", "none"));
    String payload = b64(Map.of("sub", sub, "email", email, "cognito:groups", groups));
    return header + "." + payload + ".";
  }

  private String b64(Map<String, ?> claims) throws Exception {
    return Base64.getUrlEncoder()
        .withoutPadding()
        .encodeToString(MAPPER.writeValueAsBytes(claims));
  }

  @Test
  void callback_withPendingInvitation_setsCookiesAnd200() throws Exception {
    invitations.save(
        AdminInvitation.builder()
            .email("inv@x.com")
            .role(Role.TRAINEE)
            .companyId(7L)
            .status(InvitationStatus.PENDING)
            .expiresAt(Instant.now().plus(1, ChronoUnit.DAYS))
            .createdAt(Instant.now())
            .build());
    when(tokenClient.exchangeAuthorizationCode(any()))
        .thenReturn(
            new CognitoTokens("access-jwt", fakeIdToken("sub1", "inv@x.com", List.of()), "refresh-jwt", 3600));

    mvc.perform(
            post("/api/v2/auth/login")
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"code\":\"abc\"}"))
        .andExpect(status().isOk())
        .andExpect(cookie().exists("access_token"))
        .andExpect(cookie().value("access_token", "access-jwt"))
        .andExpect(cookie().httpOnly("access_token", true))
        .andExpect(cookie().exists("refresh_token"));
  }

  @Test
  void callback_withoutInvitation_returns403AndNoAuthCookie() throws Exception {
    when(tokenClient.exchangeAuthorizationCode(any()))
        .thenReturn(
            new CognitoTokens("access-jwt", fakeIdToken("sub2", "nobody@x.com", List.of()), "refresh-jwt", 3600));

    mvc.perform(
            post("/api/v2/auth/login")
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"code\":\"abc\"}"))
        .andExpect(status().isForbidden())
        // クリア用に maxAge=0 の Set-Cookie が出る(値は空)。
        .andExpect(cookie().maxAge("access_token", 0));
  }

  @Test
  void callback_tokenExchangeFails_returns401() throws Exception {
    when(tokenClient.exchangeAuthorizationCode(any()))
        .thenThrow(new TokenExchangeException(400, "invalid_grant"));

    mvc.perform(
            post("/api/v2/auth/login")
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"code\":\"bad\"}"))
        .andExpect(status().isUnauthorized());
  }

  @Test
  void logout_clearsCookies() throws Exception {
    mvc.perform(post("/api/v2/auth/logout"))
        .andExpect(status().isOk())
        .andExpect(cookie().maxAge("access_token", 0))
        .andExpect(cookie().maxAge("refresh_token", 0));
  }

  @Test
  void refresh_withCookie_reissuesAccessToken() throws Exception {
    when(tokenClient.refreshAccessToken(any()))
        .thenReturn(new CognitoTokens("new-access", "", "", 3600));

    mvc.perform(post("/api/v2/auth/refresh").cookie(new Cookie("refresh_token", "rt")))
        .andExpect(status().isOk())
        .andExpect(cookie().value("access_token", "new-access"));
  }

  @Test
  void refresh_withoutCookie_returns401() throws Exception {
    mvc.perform(post("/api/v2/auth/refresh")).andExpect(status().isUnauthorized());
  }

  @Test
  void refresh_withExpiredAccessTokenCookie_stillReachesController() throws Exception {
    // refresh は access_token が期限切れの状態で叩かれる。壊れた access_token Cookie が
    // 付いていても JWT 検証で 401 にならず、controller に到達して再発行できること。
    when(tokenClient.refreshAccessToken(any()))
        .thenReturn(new CognitoTokens("new-access", "", "", 3600));

    mvc.perform(
            post("/api/v2/auth/refresh")
                .cookie(new Cookie("refresh_token", "rt"))
                .cookie(new Cookie("access_token", "expired-or-garbage")))
        .andExpect(status().isOk())
        .andExpect(cookie().value("access_token", "new-access"));
  }
}
