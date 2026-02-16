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
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.form.ProfileForm;
import com.example.FreStyle.usecase.GetProfileUseCase;
import com.example.FreStyle.usecase.UpdateProfileUseCase;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/profile")
@Slf4j
public class ProfileController {

    private final GetProfileUseCase getProfileUseCase;
    private final UpdateProfileUseCase updateProfileUseCase;

    @GetMapping("/me")
    public ResponseEntity<?> getProfile(@AuthenticationPrincipal Jwt jwt) {
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
            ProfileDto profileDto = getProfileUseCase.execute(sub);
            return ResponseEntity.ok(profileDto);
        } catch (ResourceNotFoundException e) {
            return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "ユーザーが存在しません。"));
        } catch (Exception e) {
            log.error("プロフィール取得エラー: {}", e.getMessage(), e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body(Map.of("error", "サーバーエラーが発生しました。"));
        }
    }

    @PutMapping("/me/update")
    public ResponseEntity<?> updateProfile(
            @AuthenticationPrincipal Jwt jwt,
            @RequestBody ProfileForm form) {

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
            updateProfileUseCase.execute(jwt, form);
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest()
                    .body(Map.of("error", e.getMessage()));
        } catch (Exception e) {
            log.error("プロフィール更新エラー: {}", e.getMessage(), e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body(Map.of("error", "サーバーエラーが発生しました。"));
        }

        return ResponseEntity.ok(Map.of("message", "プロフィールを更新しました。"));
    }
}
