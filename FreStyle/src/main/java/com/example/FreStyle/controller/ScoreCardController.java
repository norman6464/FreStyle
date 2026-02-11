package com.example.FreStyle.controller;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.dto.ScoreCardDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.AiChatSessionService;
import com.example.FreStyle.service.ScoreCardService;
import com.example.FreStyle.service.UserIdentityService;

import lombok.RequiredArgsConstructor;

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/scores")
public class ScoreCardController {

    private static final Logger logger = LoggerFactory.getLogger(ScoreCardController.class);
    private final ScoreCardService scoreCardService;
    private final AiChatSessionService aiChatSessionService;
    private final UserIdentityService userIdentityService;

    /**
     * セッションのスコアカードを取得
     */
    @GetMapping("/sessions/{sessionId}")
    public ResponseEntity<ScoreCardDto> getSessionScoreCard(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer sessionId
    ) {
        logger.info("========== GET /api/scores/sessions/{} ==========", sessionId);

        try {
            String sub = jwt.getSubject();
            User user = userIdentityService.findUserBySub(sub);

            // 権限チェック
            aiChatSessionService.getSessionByIdAndUserId(sessionId, user.getId());

            ScoreCardDto scoreCard = scoreCardService.getScoreCard(sessionId);
            logger.info("✅ スコアカード取得成功 - sessionId: {}", sessionId);

            return ResponseEntity.ok(scoreCard);
        } catch (RuntimeException e) {
            logger.error("❌ スコアカード取得エラー: {}", e.getMessage());
            return ResponseEntity.notFound().build();
        }
    }
}
