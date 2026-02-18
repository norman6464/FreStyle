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
import lombok.extern.slf4j.Slf4j;

@RestController
@RequestMapping("/api/user-profile")
@RequiredArgsConstructor
@Slf4j
public class UserProfileController {

    private final GetUserProfileUseCase getUserProfileUseCase;
    private final CreateUserProfileUseCase createUserProfileUseCase;
    private final UpdateUserProfileUseCase updateUserProfileUseCase;
    private final UpsertUserProfileUseCase upsertUserProfileUseCase;
    private final DeleteUserProfileUseCase deleteUserProfileUseCase;

    @GetMapping("/me")
    public ResponseEntity<?> getMyProfile(@AuthenticationPrincipal Jwt jwt) {
        String sub = jwt.getSubject();
        log.info("プロファイル取得: sub={}", sub);
        UserProfileDto profileDto = getUserProfileUseCase.execute(sub);
        if (profileDto == null) {
            return ResponseEntity.ok(Map.of("message", "プロファイルが設定されていません。"));
        }
        return ResponseEntity.ok(profileDto);
    }

    @PostMapping("/me/create")
    public ResponseEntity<?> createMyProfile(
            @AuthenticationPrincipal Jwt jwt,
            @Validated @RequestBody UserProfileForm form) {
        String sub = jwt.getSubject();
        log.info("プロファイル作成: sub={}", sub);
        UserProfileDto profileDto = createUserProfileUseCase.execute(sub, form);
        return ResponseEntity.status(HttpStatus.CREATED).body(profileDto);
    }

    @PostMapping("/me/update")
    public ResponseEntity<?> updateMyProfile(
            @AuthenticationPrincipal Jwt jwt,
            @Validated @RequestBody UserProfileForm form) {
        String sub = jwt.getSubject();
        log.info("プロファイル更新: sub={}", sub);
        UserProfileDto profileDto = updateUserProfileUseCase.execute(sub, form);
        return ResponseEntity.ok(profileDto);
    }

    @PostMapping("/me/upsert")
    public ResponseEntity<?> upsertMyProfile(
            @AuthenticationPrincipal Jwt jwt,
            @Validated @RequestBody UserProfileForm form) {
        String sub = jwt.getSubject();
        log.info("プロファイルupsert: sub={}", sub);
        UserProfileDto profileDto = upsertUserProfileUseCase.execute(sub, form);
        return ResponseEntity.ok(profileDto);
    }

    @PostMapping("/me/delete")
    public ResponseEntity<?> deleteMyProfile(@AuthenticationPrincipal Jwt jwt) {
        String sub = jwt.getSubject();
        log.info("プロファイル削除: sub={}", sub);
        deleteUserProfileUseCase.execute(sub);
        return ResponseEntity.ok(Map.of("message", "プロファイルを削除しました。"));
    }
}
