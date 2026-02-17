package com.example.FreStyle.controller;

import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.dto.SessionNoteDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.GetSessionNoteUseCase;
import com.example.FreStyle.usecase.SaveSessionNoteUseCase;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/session-notes")
@Slf4j
public class SessionNoteController {

    private final GetSessionNoteUseCase getSessionNoteUseCase;
    private final SaveSessionNoteUseCase saveSessionNoteUseCase;
    private final UserIdentityService userIdentityService;

    @GetMapping("/{sessionId}")
    public ResponseEntity<SessionNoteDto> getNote(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer sessionId) {
        User user = resolveUser(jwt);
        log.info("セッションノート取得: userId={}, sessionId={}", user.getId(), sessionId);
        SessionNoteDto dto = getSessionNoteUseCase.execute(user.getId(), sessionId);
        return ResponseEntity.ok(dto);
    }

    @PutMapping("/{sessionId}")
    public ResponseEntity<Void> saveNote(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer sessionId,
            @RequestBody SaveNoteRequest request) {
        User user = resolveUser(jwt);
        log.info("セッションノート保存: userId={}, sessionId={}", user.getId(), sessionId);
        saveSessionNoteUseCase.execute(user, sessionId, request.note());
        return ResponseEntity.ok().build();
    }

    private User resolveUser(Jwt jwt) {
        return userIdentityService.findUserBySub(jwt.getSubject());
    }

    public record SaveNoteRequest(String note) {}
}
