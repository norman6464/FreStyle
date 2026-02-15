package com.example.FreStyle.controller;

import com.example.FreStyle.dto.PresignedUrlRequest;
import com.example.FreStyle.dto.PresignedUrlResponse;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.NoteImageService;
import com.example.FreStyle.service.UserIdentityService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.*;

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/notes")
public class NoteImageController {

    private static final Logger logger = LoggerFactory.getLogger(NoteImageController.class);
    private final NoteImageService noteImageService;
    private final UserIdentityService userIdentityService;

    @PostMapping("/{noteId}/images/presigned-url")
    public ResponseEntity<?> getPresignedUrl(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable String noteId,
            @Valid @RequestBody PresignedUrlRequest request
    ) {
        logger.info("Presigned URLリクエスト受信 - noteId: {}, fileName: {}, contentType: {}",
                noteId, request.fileName(), request.contentType());
        try {
            if (jwt == null) {
                logger.warn("JWT is null - 認証されていないリクエスト");
                return ResponseEntity.status(401).body(java.util.Map.of("error", "認証が必要です"));
            }
            String sub = jwt.getSubject();
            logger.debug("JWT subject: {}", sub);
            User user = userIdentityService.findUserBySub(sub);

            PresignedUrlResponse response = noteImageService.generatePresignedUrl(
                    user.getId(), noteId, request.fileName(), request.contentType()
            );
            logger.info("Presigned URL生成成功 - noteId: {}, fileName: {}", noteId, request.fileName());

            return ResponseEntity.ok(response);
        } catch (IllegalArgumentException e) {
            logger.warn("Presigned URL生成失敗 - 不正なリクエスト: {}", e.getMessage());
            return ResponseEntity.badRequest().body(java.util.Map.of("error", e.getMessage()));
        } catch (Exception e) {
            logger.error("Presigned URL生成失敗 - noteId: {}, fileName: {}", noteId, request.fileName(), e);
            return ResponseEntity.internalServerError().body(java.util.Map.of("error", "画像アップロードの準備に失敗しました"));
        }
    }
}
