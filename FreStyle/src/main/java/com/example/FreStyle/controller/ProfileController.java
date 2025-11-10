package com.example.FreStyle.controller;

import java.util.Map;

import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.dto.ProfileDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.form.ProfileForm;
import com.example.FreStyle.service.CognitoAuthService;
import com.example.FreStyle.service.UserService;

// ユーザー情報
// Cognito/DB両方を使う
@RestController
@RequestMapping("/api/profile")
public class ProfileController {
  
  private final CognitoAuthService cognitoAuthService;
  private final UserService userService;
  
  public ProfileController(CognitoAuthService cognitoAuthService, UserService userService) {
    this.cognitoAuthService = cognitoAuthService;
    this.userService = userService;
  }
  
  @GetMapping("/me")
  public ResponseEntity<?> getProfile(@AuthenticationPrincipal Jwt jwt) {
    
    String sub = jwt.getSubject();
    
    if (sub == null || sub.isEmpty()) {
      return ResponseEntity.badRequest().body(Map.of("error", "無効なリクエストです。"));
    }
    
    try {
    
    User user = userService.findUser(sub);
    
    ProfileDto profileDto = new ProfileDto(
      user.getUsername(),
      user.getBio());
    
    return ResponseEntity.ok().body(profileDto);
    } catch (Exception e) {
      return ResponseEntity.badRequest().body(Map.of("error", "無効なリクエストです。"));
    }
    
  }
  
@PutMapping("/me/update")
public ResponseEntity<?> updateProfile(
    @AuthenticationPrincipal Jwt jwt,
    @RequestBody ProfileForm form) {

  System.out.println("PUT /api/profile/me/update");
  String sub = jwt.getSubject();

  if (sub == null || sub.isEmpty()) {
    return ResponseEntity.badRequest().body(Map.of("error", "無効なリクエストです。"));
  }

  try {
    // ==== 分岐ポイント ====
    // 外部IDプロバイダでログインしたユーザーには "cognito:groups" が含まれる
    boolean isOidcUser = jwt.hasClaim("cognito:groups");

    if (isOidcUser) {
      // OIDC（Google、Appleなど外部IDP）経由 → DBのみ更新
      System.out.println("request user is OIDC user no cognito update db only update");
      userService.updateUser(form, sub);

    } else {
      // 通常のCognitoユーザー → Cognito + DBを更新
      String accessToken = jwt.getTokenValue();
      cognitoAuthService.updateUserProfile(accessToken, form.getUsername());
      userService.updateUser(form, sub);
    }

  } catch (Exception e) {
    System.out.println("profile update error" + e.getMessage());
    return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
  }

  return ResponseEntity.ok().body(Map.of("message", "プロフィールを更新しました。"));
}

  
  
  
}
