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
    public ResponseEntity<PresignedUrlResponse> getPresignedUrl(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable String noteId,
            @Valid @RequestBody PresignedUrlRequest request
    ) {
        String sub = jwt.getSubject();
        User user = userIdentityService.findUserBySub(sub);

        PresignedUrlResponse response = noteImageService.generatePresignedUrl(
                user.getId(), noteId, request.fileName(), request.contentType()
        );
        logger.info("Presigned URL生成成功 - noteId: {}, fileName: {}", noteId, request.fileName());

        return ResponseEntity.ok(response);
    }
}
