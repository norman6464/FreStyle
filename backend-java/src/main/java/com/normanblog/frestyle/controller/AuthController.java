package com.normanblog.frestyle.controller;

import com.normanblog.frestyle.cognito.CognitoTokens;
import com.normanblog.frestyle.cognito.TokenExchangeException;
import com.normanblog.frestyle.dto.CallbackRequest;
import com.normanblog.frestyle.dto.MeResponse;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.security.AuthCookies;
import com.normanblog.frestyle.security.CurrentUserProvider;
import com.normanblog.frestyle.service.AuthService;
import com.normanblog.frestyle.service.LoginService;
import jakarta.servlet.http.HttpServletResponse;
import jakarta.validation.Valid;
import java.util.List;
import java.util.Map;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.CookieValue;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/** 認証(自己情報取得 + Cognito ログインフロー)のエンドポイント。 */
@RestController
@RequestMapping("/api/v2/auth")
public class AuthController {

  private final CurrentUserProvider currentUser;
  private final AuthService authService;
  private final LoginService loginService;
  private final AuthCookies cookies;

  public AuthController(
      CurrentUserProvider currentUser,
      AuthService authService,
      LoginService loginService,
      AuthCookies cookies) {
    this.currentUser = currentUser;
    this.authService = authService;
    this.loginService = loginService;
    this.cookies = cookies;
  }

  /** 現在ログイン中のユーザー情報(+ groups / isAdmin / onboarded)を返す。 */
  @GetMapping("/me")
  public MeResponse me() {
    Jwt jwt = currentUser.currentJwt();
    User user = currentUser.require();
    List<String> groups = currentUser.groups(jwt);
    boolean isAdmin = currentUser.isAdmin(groups);
    user = authService.syncAdminRole(user, isAdmin);
    return MeResponse.of(user, groups, isAdmin);
  }

  /** 認可コードを token に交換し、HttpOnly Cookie を発行する。招待が無い新規ユーザーは 403。 */
  @PostMapping("/login")
  public ResponseEntity<Map<String, String>> login(
      @Valid @RequestBody CallbackRequest request, HttpServletResponse response) {
    LoginService.LoginResult result;
    try {
      result = loginService.callback(request.code(), request.invitationToken());
    } catch (TokenExchangeException e) {
      // 交換失敗時は古いセッションを残さないよう既存 Cookie もクリアする。
      cookies.clear(response);
      return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
          .body(Map.of("error", "login_failed"));
    }

    if (!result.allowed()) {
      // 招待ゲートで弾かれた新規ユーザー。発行しかけた Cookie を残さない。
      cookies.clear(response);
      return ResponseEntity.status(HttpStatus.FORBIDDEN)
          .body(
              Map.of(
                  "error",
                  "invitation_required",
                  "message",
                  "ご利用には企業申請、または所属企業からの招待が必要です。招待メールに記載されたリンクからログインしてください。"));
    }

    CognitoTokens tokens = result.tokens();
    cookies.setAccessToken(response, tokens.accessToken(), tokens.expiresIn());
    cookies.setRefreshToken(response, tokens.refreshToken());
    return ResponseEntity.ok(Map.of("message", "ログインしました。"));
  }

  /** access/refresh Cookie を破棄する。 */
  @PostMapping("/logout")
  public ResponseEntity<Map<String, String>> logout(HttpServletResponse response) {
    cookies.clear(response);
    return ResponseEntity.ok(Map.of("message", "ログアウトしました。"));
  }

  /** refresh_token Cookie で access_token を再発行する。失効していれば Cookie を破棄して 401。 */
  @PostMapping("/refresh")
  public ResponseEntity<Map<String, String>> refresh(
      @CookieValue(name = "refresh_token", required = false) String refreshToken,
      HttpServletResponse response) {
    if (refreshToken == null || refreshToken.isBlank()) {
      return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
          .body(Map.of("error", "refresh_token_missing"));
    }
    try {
      LoginService.LoginResult result = loginService.refresh(refreshToken);
      if (!result.allowed()) {
        // 招待取り消し / ユーザー削除等で許可が外れたらセッションを破棄する。
        cookies.clear(response);
        return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(Map.of("error", "not_allowed"));
      }
      cookies.setAccessToken(response, result.tokens().accessToken(), result.tokens().expiresIn());
      return ResponseEntity.ok(Map.of("message", "refreshed"));
    } catch (TokenExchangeException e) {
      cookies.clear(response);
      return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(Map.of("error", "refresh_failed"));
    }
  }
}
