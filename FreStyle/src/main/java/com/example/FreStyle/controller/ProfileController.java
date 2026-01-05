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

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/profile")
public class ProfileController {

    private final CognitoAuthService cognitoAuthService;
    private final UserService userService;
    private final UserIdentityService userIdentityService;

    // -----------------------
    // GET /me
    // -----------------------
    @GetMapping("/me")
    public ResponseEntity<?> getProfile(@AuthenticationPrincipal Jwt jwt) {
        System.out.println("[ProfileController /me] Endpoint called");
        
        if (jwt == null) {
            System.out.println("[ProfileController /me] ERROR: JWT is null - user not authenticated");
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "認証に失敗しました。"));
        }
        
        System.out.println("[ProfileController /me] JWT Principal: " + jwt.toString());
        
        String sub = jwt.getSubject();
        System.out.println("[ProfileController /me] JWT Subject (sub): " + sub);

        if (sub == null || sub.isEmpty()) {
            // JWT が無効 → 認証されていない
            System.out.println("[ProfileController /me] ERROR: JWT subject is null or empty");
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "認証に失敗しました。"));
        }

        try {
            System.out.println("[ProfileController /me] Finding user with sub: " + sub);
            User user = userIdentityService.findUserBySub(sub);

            if (user == null) {
                // sub に紐づくユーザーが DB に存在しない
                System.out.println("[ProfileController /me] ERROR: User not found for sub: " + sub);
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                        .body(Map.of("error", "ユーザーが存在しません。"));
            }
            
            System.out.println("[ProfileController /me] User found - ID: " + user.getId() + ", Name: " + user.getName());

            ProfileDto profileDto = new ProfileDto(
                    user.getName(),
                    user.getBio());

            System.out.println("[ProfileController /me] Returning profile data");
            return ResponseEntity.ok(profileDto);

        } catch (Exception e) {
            // 想定外のエラー
            System.out.println("[ProfileController /me] ERROR: " + e.getClass().getSimpleName() + " - " + e.getMessage());
            e.printStackTrace();
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

        System.out.println("[ProfileController /me/update] Endpoint called");
        
        if (jwt == null) {
            System.out.println("[ProfileController /me/update] ERROR: JWT is null - user not authenticated");
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "認証に失敗しました。"));
        }
        
        System.out.println("[ProfileController /me/update] JWT Principal: " + jwt.toString());

        String sub = jwt.getSubject();
        System.out.println("[ProfileController /me/update] JWT Subject (sub): " + sub);

        if (sub == null || sub.isEmpty()) {
            System.out.println("[ProfileController /me/update] ERROR: JWT subject is null or empty");
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "認証に失敗しました。"));
        }

        System.out.println("[ProfileController /me/update] Update request - name: " + form.getName() + 
                          ", bio: " + (form.getBio() != null ? form.getBio().substring(0, Math.min(30, form.getBio().length())) + "..." : "null"));

        try {
            boolean isOidcUser = jwt.hasClaim("cognito:groups");
            System.out.println("[ProfileController /me/update] User type - isOidcUser: " + isOidcUser);

            if (isOidcUser) {
                // OIDCユーザー → DBのみ更新
                System.out.println("[ProfileController /me/update] OIDC user detected - updating DB only");
                userService.updateUser(form, sub);
                System.out.println("[ProfileController /me/update] DB update successful");

            } else {
                // Cognitoユーザー → Cognito + DB 更新
                System.out.println("[ProfileController /me/update] Cognito user detected - updating Cognito and DB");
                String accessToken = jwt.getTokenValue();
                System.out.println("[ProfileController /me/update] Updating Cognito profile");
                cognitoAuthService.updateUserProfile(accessToken, form.getName());
                System.out.println("[ProfileController /me/update] Cognito update successful - updating DB");
                userService.updateUser(form, sub);
                System.out.println("[ProfileController /me/update] DB update successful");
            }

        } catch (IllegalArgumentException e) {
            // フォームの入力値が不正等
            System.out.println("[ProfileController /me/update] ERROR: IllegalArgumentException - " + e.getMessage());
            return ResponseEntity.badRequest()
                    .body(Map.of("error", e.getMessage()));

        } catch (Exception e) {
            // 想定外エラー
            System.out.println("[ProfileController /me/update] ERROR: " + e.getClass().getSimpleName() + " - " + e.getMessage());
            e.printStackTrace();
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body(Map.of("error", "サーバーエラーが発生しました。"));
        }

        System.out.println("[ProfileController /me/update] Profile update completed successfully");
        return ResponseEntity.ok(Map.of("message", "プロフィールを更新しました。"));
    }
}
