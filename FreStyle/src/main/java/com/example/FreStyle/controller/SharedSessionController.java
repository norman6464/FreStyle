package com.example.FreStyle.controller;

import java.util.List;

import jakarta.validation.Valid;

import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.dto.SharedSessionDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.form.ShareSessionForm;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.GetPublicSessionsUseCase;
import com.example.FreStyle.usecase.ShareSessionUseCase;
import com.example.FreStyle.usecase.UnshareSessionUseCase;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/shared-sessions")
@Slf4j
public class SharedSessionController {

    private final GetPublicSessionsUseCase getPublicSessionsUseCase;
    private final ShareSessionUseCase shareSessionUseCase;
    private final UnshareSessionUseCase unshareSessionUseCase;
    private final UserIdentityService userIdentityService;

    @GetMapping
    public ResponseEntity<List<SharedSessionDto>> getPublicSessions() {
        log.info("========== GET /api/shared-sessions ==========");
        List<SharedSessionDto> sessions = getPublicSessionsUseCase.execute();
        log.info("公開セッション一覧取得成功 - 件数: {}", sessions.size());
        return ResponseEntity.ok(sessions);
    }

    @PostMapping
    public ResponseEntity<SharedSessionDto> shareSession(
            @AuthenticationPrincipal Jwt jwt,
            @Valid @RequestBody ShareSessionForm form) {
        User user = resolveUser(jwt);
        log.info("========== POST /api/shared-sessions ========== userId={}, sessionId={}", user.getId(), form.sessionId());
        SharedSessionDto dto = shareSessionUseCase.execute(user.getId(), form);
        log.info("セッション共有成功 - sharedSessionId: {}", dto.id());
        return ResponseEntity.ok(dto);
    }

    @DeleteMapping("/{sessionId}")
    public ResponseEntity<Void> unshareSession(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer sessionId) {
        User user = resolveUser(jwt);
        log.info("========== DELETE /api/shared-sessions/{} ========== userId={}", sessionId, user.getId());
        unshareSessionUseCase.execute(user.getId(), sessionId);
        log.info("セッション共有解除成功 - sessionId: {}", sessionId);
        return ResponseEntity.ok().build();
    }

    private User resolveUser(Jwt jwt) {
        return userIdentityService.findUserBySub(jwt.getSubject());
    }
}
