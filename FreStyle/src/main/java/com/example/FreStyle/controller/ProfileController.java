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

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/profile")
@Slf4j
public class ProfileController {

    private final CognitoAuthService cognitoAuthService;
    private final UserService userService;
    private final UserIdentityService userIdentityService;

    // -----------------------
    // GET /me
    // -----------------------
    @GetMapping("/me")
    public ResponseEntity<?> getProfile(@AuthenticationPrincipal Jwt jwt) {
        log.info("[ProfileController /me] Endpoint called");
        
        if (jwt == null) {
            log.warn("認証エラー: JWTがnull");
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "認証に失敗しました。"));
        }
        
        log.info("[ProfileController /me] JWT Principal: {}", jwt);

        String sub = jwt.getSubject();
        log.info("[ProfileController /me] JWT Subject (sub): {}", sub);

        if (sub == null || sub.isEmpty()) {
            // JWT が無効 → 認証されていない
            log.warn("認証エラー: JWTのsubがnullまたは空");
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "認証に失敗しました。"));
        }

        try {
            log.info("[ProfileController /me] Finding user with sub: {}", sub);
            User user = userIdentityService.findUserBySub(sub);

            if (user == null) {
                // sub に紐づくユーザーが DB に存在しない
                log.warn("ユーザー未存在: sub={}", sub);
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                        .body(Map.of("error", "ユーザーが存在しません。"));
            }
            
            log.info("[ProfileController /me] User found - ID: {}, Name: {}", user.getId(), user.getName());

            ProfileDto profileDto = new ProfileDto(
                    user.getName(),
                    user.getBio());

            log.info("[ProfileController /me] Returning profile data");
            return ResponseEntity.ok(profileDto);

        } catch (Exception e) {
            // 想定外のエラー
            log.error("プロフィール取得エラー: {}", e.getMessage(), e);
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

        log.info("[ProfileController /me/update] Endpoint called");
        
        if (jwt == null) {
            log.warn("認証エラー: JWTがnull");
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "認証に失敗しました。"));
        }
        
        log.info("[ProfileController /me/update] JWT Principal: {}", jwt);

        String sub = jwt.getSubject();
        log.info("[ProfileController /me/update] JWT Subject (sub): {}", sub);

        if (sub == null || sub.isEmpty()) {
            log.warn("認証エラー: JWTのsubがnullまたは空");
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "認証に失敗しました。"));
        }

        log.info("[ProfileController /me/update] Update request - name: {}, bio: {}", form.getName(),
                          form.getBio() != null ? form.getBio().substring(0, Math.min(30, form.getBio().length())) + "..." : "null");

        try {
            boolean isOidcUser = jwt.hasClaim("cognito:groups");
            log.info("[ProfileController /me/update] User type - isOidcUser: {}", isOidcUser);

            if (isOidcUser) {
                // OIDCユーザー → DBのみ更新
                log.info("[ProfileController /me/update] OIDC user detected - updating DB only");
                userService.updateUser(form, sub);
                log.info("[ProfileController /me/update] DB update successful");

            } else {
                // Cognitoユーザー → Cognito + DB 更新
                log.info("[ProfileController /me/update] Cognito user detected - updating Cognito and DB");
                String accessToken = jwt.getTokenValue();
                log.info("[ProfileController /me/update] Updating Cognito profile");
                cognitoAuthService.updateUserProfile(accessToken, form.getName());
                log.info("[ProfileController /me/update] Cognito update successful - updating DB");
                userService.updateUser(form, sub);
                log.info("[ProfileController /me/update] DB update successful");
            }

        } catch (IllegalArgumentException e) {
            // フォームの入力値が不正等
            log.warn("プロフィール更新エラー(入力値不正): {}", e.getMessage());
            return ResponseEntity.badRequest()
                    .body(Map.of("error", e.getMessage()));

        } catch (Exception e) {
            // 想定外エラー
            log.error("プロフィール更新エラー: {}", e.getMessage(), e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body(Map.of("error", "サーバーエラーが発生しました。"));
        }

        log.info("[ProfileController /me/update] Profile update completed successfully");
        return ResponseEntity.ok(Map.of("message", "プロフィールを更新しました。"));
    }
}
