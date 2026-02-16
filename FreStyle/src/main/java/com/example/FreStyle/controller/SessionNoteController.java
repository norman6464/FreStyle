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

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/session-notes")
public class SessionNoteController {

    private final GetSessionNoteUseCase getSessionNoteUseCase;
    private final SaveSessionNoteUseCase saveSessionNoteUseCase;
    private final UserIdentityService userIdentityService;

    @GetMapping("/{sessionId}")
    public ResponseEntity<SessionNoteDto> getNote(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer sessionId) {
        String sub = jwt.getSubject();
        User user = userIdentityService.findUserBySub(sub);
        SessionNoteDto dto = getSessionNoteUseCase.execute(user.getId(), sessionId);
        return ResponseEntity.ok(dto);
    }

    @PutMapping("/{sessionId}")
    public ResponseEntity<Void> saveNote(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer sessionId,
            @RequestBody SaveNoteRequest request) {
        String sub = jwt.getSubject();
        User user = userIdentityService.findUserBySub(sub);
        saveSessionNoteUseCase.execute(user, sessionId, request.note());
        return ResponseEntity.ok().build();
    }

    public record SaveNoteRequest(String note) {}
}
