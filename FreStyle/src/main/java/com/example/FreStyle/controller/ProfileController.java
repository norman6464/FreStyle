package com.example.FreStyle.controller;

import java.util.Map;

import org.springframework.http.HttpStatus;
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
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.service.UserService;

@RestController
@RequestMapping("/api/profile")
public class ProfileController {

    private final CognitoAuthService cognitoAuthService;
    private final UserService userService;
    private final UserIdentityService userIdentityService;

    public ProfileController(CognitoAuthService cognitoAuthService, UserService userService, UserIdentityService userIdentityService) {
        this.cognitoAuthService = cognitoAuthService;
        this.userService = userService;
        this.userIdentityService = userIdentityService;
    }

    // -----------------------
    // GET /me
    // -----------------------
    @GetMapping("/me")
    public ResponseEntity<?> getProfile(@AuthenticationPrincipal Jwt jwt) {

        String sub = jwt.getSubject();

        if (sub == null || sub.isEmpty()) {
            // JWT が無効 → 認証されていない
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "認証に失敗しました。"));
        }

        try {
            User user = userIdentityService.findUserBySub(sub);

            if (user == null) {
                // sub に紐づくユーザーが DB に存在しない
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                        .body(Map.of("error", "ユーザーが存在しません。"));
            }

            ProfileDto profileDto = new ProfileDto(
                    user.getUsername(),
                    user.getBio());

            return ResponseEntity.ok(profileDto);

        } catch (Exception e) {
            // 想定外のエラー
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body(Map.of("error", "サーバーエラーが発生しました。"));
        }
    }

    // -----------------------
    // PUT /me/update
    // -----------------------
    @PutMapping("/me/update")
    public ResponseEntity<?> updateProfile(
            @AuthenticationPrincipal Jwt jwt,
            @RequestBody ProfileForm form) {

        System.out.println("PUT /api/profile/me/update");

        String sub = jwt.getSubject();

        if (sub == null || sub.isEmpty()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "認証に失敗しました。"));
        }

        try {
            boolean isOidcUser = jwt.hasClaim("cognito:groups");

            if (isOidcUser) {
                // OIDCユーザー → DBのみ更新
                System.out.println("request user is OIDC user no cognito update db only update");
                userService.updateUser(form, sub);

            } else {
                // Cognitoユーザー → Cognito + DB 更新
                String accessToken = jwt.getTokenValue();
                cognitoAuthService.updateUserProfile(accessToken, form.getUsername());
                userService.updateUser(form, sub);
            }

        } catch (IllegalArgumentException e) {
            // フォームの入力値が不正等
            return ResponseEntity.badRequest()
                    .body(Map.of("error", e.getMessage()));

        } catch (Exception e) {
            // 想定外エラー
            System.out.println("profile update error" + e.getMessage());
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body(Map.of("error", "サーバーエラーが発生しました。"));
        }

        return ResponseEntity.ok(Map.of("message", "プロフィールを更新しました。"));
    }
}
