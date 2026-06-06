package com.normanblog.frestyle.controller;

import com.normanblog.frestyle.dto.MeResponse;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.security.CurrentUserProvider;
import com.normanblog.frestyle.service.AuthService;
import java.util.List;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/** 認証ユーザーの自己情報を返すコントローラ。 */
@RestController
@RequestMapping("/api/v2/auth")
public class AuthController {

  private final CurrentUserProvider currentUser;
  private final AuthService authService;

  public AuthController(CurrentUserProvider currentUser, AuthService authService) {
    this.currentUser = currentUser;
    this.authService = authService;
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
}
