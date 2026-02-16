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
import com.example.FreStyle.form.UserProfileForm;
import com.example.FreStyle.usecase.CreateUserProfileUseCase;
import com.example.FreStyle.usecase.DeleteUserProfileUseCase;
import com.example.FreStyle.usecase.GetUserProfileUseCase;
import com.example.FreStyle.usecase.UpdateUserProfileUseCase;
import com.example.FreStyle.usecase.UpsertUserProfileUseCase;

import lombok.RequiredArgsConstructor;

@RestController
@RequestMapping("/api/user-profile")
@RequiredArgsConstructor
public class UserProfileController {

    private final GetUserProfileUseCase getUserProfileUseCase;
    private final CreateUserProfileUseCase createUserProfileUseCase;
    private final UpdateUserProfileUseCase updateUserProfileUseCase;
    private final UpsertUserProfileUseCase upsertUserProfileUseCase;
    private final DeleteUserProfileUseCase deleteUserProfileUseCase;

    @GetMapping("/me")
    public ResponseEntity<?> getMyProfile(@AuthenticationPrincipal Jwt jwt) {
        if (jwt == null || jwt.getSubject() == null || jwt.getSubject().isEmpty()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "認証に失敗しました。"));
        }
        try {
            UserProfileDto profileDto = getUserProfileUseCase.execute(jwt.getSubject());
            if (profileDto == null) {
                return ResponseEntity.ok(Map.of("message", "プロファイルが設定されていません。"));
            }
            return ResponseEntity.ok(profileDto);
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body(Map.of("error", "サーバーエラーが発生しました。"));
        }
    }

    @PostMapping("/me/create")
    public ResponseEntity<?> createMyProfile(
            @AuthenticationPrincipal Jwt jwt,
            @Validated @RequestBody UserProfileForm form) {
        if (jwt == null || jwt.getSubject() == null || jwt.getSubject().isEmpty()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "認証に失敗しました。"));
        }
        try {
            UserProfileDto profileDto = createUserProfileUseCase.execute(jwt.getSubject(), form);
            return ResponseEntity.status(HttpStatus.CREATED).body(profileDto);
        } catch (RuntimeException e) {
            return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                    .body(Map.of("error", e.getMessage()));
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body(Map.of("error", "サーバーエラーが発生しました。"));
        }
    }

    @PostMapping("/me/update")
    public ResponseEntity<?> updateMyProfile(
            @AuthenticationPrincipal Jwt jwt,
            @Validated @RequestBody UserProfileForm form) {
        if (jwt == null || jwt.getSubject() == null || jwt.getSubject().isEmpty()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "認証に失敗しました。"));
        }
        try {
            UserProfileDto profileDto = updateUserProfileUseCase.execute(jwt.getSubject(), form);
            return ResponseEntity.ok(profileDto);
        } catch (RuntimeException e) {
            return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                    .body(Map.of("error", e.getMessage()));
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body(Map.of("error", "サーバーエラーが発生しました。"));
        }
    }

    @PostMapping("/me/upsert")
    public ResponseEntity<?> upsertMyProfile(
            @AuthenticationPrincipal Jwt jwt,
            @Validated @RequestBody UserProfileForm form) {
        if (jwt == null || jwt.getSubject() == null || jwt.getSubject().isEmpty()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "認証に失敗しました。"));
        }
        try {
            UserProfileDto profileDto = upsertUserProfileUseCase.execute(jwt.getSubject(), form);
            return ResponseEntity.ok(profileDto);
        } catch (RuntimeException e) {
            return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                    .body(Map.of("error", e.getMessage()));
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body(Map.of("error", "サーバーエラーが発生しました。"));
        }
    }

    @PostMapping("/me/delete")
    public ResponseEntity<?> deleteMyProfile(@AuthenticationPrincipal Jwt jwt) {
        if (jwt == null || jwt.getSubject() == null || jwt.getSubject().isEmpty()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "認証に失敗しました。"));
        }
        try {
            deleteUserProfileUseCase.execute(jwt.getSubject());
            return ResponseEntity.ok(Map.of("message", "プロファイルを削除しました。"));
        } catch (RuntimeException e) {
            return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                    .body(Map.of("error", e.getMessage()));
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body(Map.of("error", "サーバーエラーが発生しました。"));
        }
    }
}
