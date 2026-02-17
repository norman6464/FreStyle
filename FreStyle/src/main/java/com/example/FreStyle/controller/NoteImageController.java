package com.example.FreStyle.controller;

import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.dto.PresignedUrlRequest;
import com.example.FreStyle.dto.PresignedUrlResponse;
import com.example.FreStyle.usecase.GeneratePresignedUrlUseCase;

import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/notes")
@Slf4j
public class NoteImageController {

    private final GeneratePresignedUrlUseCase generatePresignedUrlUseCase;

    @PostMapping("/{noteId}/images/presigned-url")
    public ResponseEntity<PresignedUrlResponse> getPresignedUrl(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable String noteId,
            @Valid @RequestBody PresignedUrlRequest request
    ) {
        log.info("Presigned URL生成: noteId={}, fileName={}", noteId, request.fileName());
        PresignedUrlResponse response = generatePresignedUrlUseCase.execute(
                jwt.getSubject(), noteId, request.fileName(), request.contentType());
        return ResponseEntity.ok(response);
    }
}
