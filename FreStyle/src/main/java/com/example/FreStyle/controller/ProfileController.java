package com.example.FreStyle.controller;

import java.util.Map;

import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.dto.PresignedUrlRequest;
import com.example.FreStyle.dto.PresignedUrlResponse;
import com.example.FreStyle.dto.ProfileDto;
import com.example.FreStyle.form.ProfileForm;
import com.example.FreStyle.usecase.GenerateProfileImageUrlUseCase;
import com.example.FreStyle.usecase.GetProfileUseCase;
import com.example.FreStyle.usecase.UpdateProfileUseCase;

import jakarta.validation.Valid;

import lombok.RequiredArgsConstructor;

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/profile")
public class ProfileController {

    private final GetProfileUseCase getProfileUseCase;
    private final UpdateProfileUseCase updateProfileUseCase;
    private final GenerateProfileImageUrlUseCase generateProfileImageUrlUseCase;

    @GetMapping("/me")
    public ResponseEntity<ProfileDto> getProfile(@AuthenticationPrincipal Jwt jwt) {
        String sub = jwt.getSubject();
        ProfileDto profileDto = getProfileUseCase.execute(sub);
        return ResponseEntity.ok(profileDto);
    }

    @PutMapping("/me/update")
    public ResponseEntity<Map<String, String>> updateProfile(
            @AuthenticationPrincipal Jwt jwt,
            @RequestBody ProfileForm form) {
        updateProfileUseCase.execute(jwt, form);
        return ResponseEntity.ok(Map.of("message", "プロフィールを更新しました。"));
    }

    @PostMapping("/me/image/presigned-url")
    public ResponseEntity<PresignedUrlResponse> getProfileImagePresignedUrl(
            @AuthenticationPrincipal Jwt jwt,
            @Valid @RequestBody PresignedUrlRequest request) {
        PresignedUrlResponse response = generateProfileImageUrlUseCase.execute(
                jwt.getSubject(), request.fileName(), request.contentType());
        return ResponseEntity.ok(response);
    }
}
