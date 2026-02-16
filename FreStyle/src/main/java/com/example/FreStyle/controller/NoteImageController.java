package com.example.FreStyle.controller;

import com.example.FreStyle.dto.PresignedUrlRequest;
import com.example.FreStyle.dto.PresignedUrlResponse;
import com.example.FreStyle.usecase.GeneratePresignedUrlUseCase;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.*;

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/notes")
@Slf4j
public class NoteImageController {

    private final GeneratePresignedUrlUseCase generatePresignedUrlUseCase;

    @PostMapping("/{noteId}/images/presigned-url")
    public ResponseEntity<?> getPresignedUrl(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable String noteId,
            @Valid @RequestBody PresignedUrlRequest request
    ) {
        try {
            PresignedUrlResponse response = generatePresignedUrlUseCase.execute(
                    jwt.getSubject(), noteId, request.fileName(), request.contentType());
            return ResponseEntity.ok(response);
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body(java.util.Map.of("error", e.getMessage()));
        } catch (Exception e) {
            log.error("Presigned URL生成失敗 - noteId: {}, fileName: {}", noteId, request.fileName(), e);
            return ResponseEntity.internalServerError().body(java.util.Map.of("error", "画像アップロードの準備に失敗しました"));
        }
    }
}
