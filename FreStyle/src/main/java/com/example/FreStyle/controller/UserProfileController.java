package com.example.FreStyle.controller;

import java.util.Map;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.dto.UserProfileDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.form.UserProfileForm;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.service.UserProfileService;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@RestController
@RequestMapping("/api/user-profile")
@RequiredArgsConstructor
@Slf4j
public class UserProfileController {

    private final UserProfileService userProfileService;
    private final UserIdentityService userIdentityService;

    // -----------------------
    // GET /api/user-profile/me - 自分のプロファイル取得
    // -----------------------
    @GetMapping("/me")
    public ResponseEntity<?> getMyProfile(@AuthenticationPrincipal Jwt jwt) {
        log.info("[UserProfileController /me] Endpoint called");
        
        if (jwt == null) {
            log.info("[UserProfileController /me] ERROR: JWT is null");
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "認証に失敗しました。"));
        }

        String sub = jwt.getSubject();
        if (sub == null || sub.isEmpty()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "認証に失敗しました。"));
        }

        try {
            User user = userIdentityService.findUserBySub(sub);
            if (user == null) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                        .body(Map.of("error", "ユーザーが存在しません。"));
            }

            UserProfileDto profileDto = userProfileService.getProfileByUserId(user.getId());
            
            if (profileDto == null) {
                // プロファイルが未作成の場合は空のレスポンス
                return ResponseEntity.ok(Map.of("message", "プロファイルが設定されていません。"));
            }

            return ResponseEntity.ok(profileDto);

        } catch (Exception e) {
            log.info("[UserProfileController /me] ERROR: " + e.getMessage());
            e.printStackTrace();
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body(Map.of("error", "サーバーエラーが発生しました。"));
        }
    }

    // -----------------------
    // POST /api/user-profile/me/create - プロファイル作成
    // -----------------------
    @PostMapping("/me/create")
    public ResponseEntity<?> createMyProfile(
            @AuthenticationPrincipal Jwt jwt,
            @Validated @RequestBody UserProfileForm form) {
        
        log.info("[UserProfileController POST /me] Endpoint called");
        
        if (jwt == null) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "認証に失敗しました。"));
        }

        String sub = jwt.getSubject();
        if (sub == null || sub.isEmpty()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "認証に失敗しました。"));
        }

        try {
            User user = userIdentityService.findUserBySub(sub);
            if (user == null) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                        .body(Map.of("error", "ユーザーが存在しません。"));
            }

            UserProfileDto profileDto = userProfileService.createProfile(user, form);
            return ResponseEntity.status(HttpStatus.CREATED).body(profileDto);

        } catch (RuntimeException e) {
            log.info("[UserProfileController POST /me] ERROR: " + e.getMessage());
            return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                    .body(Map.of("error", e.getMessage()));
        } catch (Exception e) {
            log.info("[UserProfileController POST /me] ERROR: " + e.getMessage());
            e.printStackTrace();
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body(Map.of("error", "サーバーエラーが発生しました。"));
        }
    }

    // -----------------------
    // PUT /api/user-profile/me/update - プロファイル更新
    // -----------------------
    @PostMapping("/me/update")
    public ResponseEntity<?> updateMyProfile(
            @AuthenticationPrincipal Jwt jwt,
            @Validated @RequestBody UserProfileForm form) {
        
        log.info("[UserProfileController PUT /me] Endpoint called");
        
        if (jwt == null) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "認証に失敗しました。"));
        }

        String sub = jwt.getSubject();
        if (sub == null || sub.isEmpty()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "認証に失敗しました。"));
        }

        try {
            User user = userIdentityService.findUserBySub(sub);
            if (user == null) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                        .body(Map.of("error", "ユーザーが存在しません。"));
            }

            UserProfileDto profileDto = userProfileService.updateProfile(user.getId(), form);
            return ResponseEntity.ok(profileDto);

        } catch (RuntimeException e) {
            log.info("[UserProfileController PUT /me] ERROR: " + e.getMessage());
            return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                    .body(Map.of("error", e.getMessage()));
        } catch (Exception e) {
            log.info("[UserProfileController PUT /me] ERROR: " + e.getMessage());
            e.printStackTrace();
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body(Map.of("error", "サーバーエラーが発生しました。"));
        }
    }

    // -----------------------
    // PUT /api/user-profile/me/upsert - プロファイル作成または更新
    // -----------------------
    @PostMapping("/me/upsert")
    public ResponseEntity<?> upsertMyProfile(
            @AuthenticationPrincipal Jwt jwt,
            @Validated @RequestBody UserProfileForm form) {
        
        log.info("[UserProfileController PUT /me/upsert] Endpoint called");
        
        if (jwt == null) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "認証に失敗しました。"));
        }

        String sub = jwt.getSubject();
        if (sub == null || sub.isEmpty()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "認証に失敗しました。"));
        }

        try {
            User user = userIdentityService.findUserBySub(sub);
            if (user == null) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                        .body(Map.of("error", "ユーザーが存在しません。"));
            }

            UserProfileDto profileDto = userProfileService.createOrUpdateProfile(user, form);
            return ResponseEntity.ok(profileDto);

        } catch (RuntimeException e) {
            log.info("[UserProfileController PUT /me/upsert] ERROR: " + e.getMessage());
            return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                    .body(Map.of("error", e.getMessage()));
        } catch (Exception e) {
            log.info("[UserProfileController PUT /me/upsert] ERROR: " + e.getMessage());
            e.printStackTrace();
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body(Map.of("error", "サーバーエラーが発生しました。"));
        }
    }

    // -----------------------
    // DELETE /api/user-profile/me - プロファイル削除
    // -----------------------
    @PostMapping("/me/delete")
    public ResponseEntity<?> deleteMyProfile(@AuthenticationPrincipal Jwt jwt) {
        log.info("[UserProfileController DELETE /me] Endpoint called");
        
        if (jwt == null) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "認証に失敗しました。"));
        }

        String sub = jwt.getSubject();
        if (sub == null || sub.isEmpty()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "認証に失敗しました。"));
        }

        try {
            User user = userIdentityService.findUserBySub(sub);
            if (user == null) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                        .body(Map.of("error", "ユーザーが存在しません。"));
            }

            userProfileService.deleteProfile(user.getId());
            return ResponseEntity.ok(Map.of("message", "プロファイルを削除しました。"));

        } catch (RuntimeException e) {
            log.info("[UserProfileController DELETE /me] ERROR: " + e.getMessage());
            return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                    .body(Map.of("error", e.getMessage()));
        } catch (Exception e) {
            log.info("[UserProfileController DELETE /me] ERROR: " + e.getMessage());
            e.printStackTrace();
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body(Map.of("error", "サーバーエラーが発生しました。"));
        }
    }
}
