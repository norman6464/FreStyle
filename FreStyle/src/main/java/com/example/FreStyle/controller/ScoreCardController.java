package com.example.FreStyle.controller;

import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

import com.example.FreStyle.dto.ScoreCardDto;
import com.example.FreStyle.dto.ScoreHistoryDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.GetAiChatSessionByIdUseCase;
import com.example.FreStyle.usecase.GetScoreCardBySessionIdUseCase;
import com.example.FreStyle.usecase.GetScoreHistoryByUserIdUseCase;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/scores")
@Slf4j
public class ScoreCardController {
    private final UserIdentityService userIdentityService;
    private final GetAiChatSessionByIdUseCase getAiChatSessionByIdUseCase;
    private final GetScoreCardBySessionIdUseCase getScoreCardBySessionIdUseCase;
    private final GetScoreHistoryByUserIdUseCase getScoreHistoryByUserIdUseCase;

    /**
     * セッションのスコアカードを取得
     */
    @GetMapping("/sessions/{sessionId}")
    public ResponseEntity<ScoreCardDto> getSessionScoreCard(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer sessionId
    ) {
        log.info("========== GET /api/scores/sessions/{} ==========", sessionId);

        User user = resolveUser(jwt);

        // 権限チェック
        getAiChatSessionByIdUseCase.execute(sessionId, user.getId());

        ScoreCardDto scoreCard = getScoreCardBySessionIdUseCase.execute(sessionId);
        log.info("✅ スコアカード取得成功 - sessionId: {}", sessionId);

        return ResponseEntity.ok(scoreCard);
    }

    /**
     * ユーザーのスコア履歴を取得
     */
    @GetMapping("/history")
    public ResponseEntity<List<ScoreHistoryDto>> getScoreHistory(
            @AuthenticationPrincipal Jwt jwt
    ) {
        log.info("========== GET /api/scores/history ==========");

        User user = resolveUser(jwt);

        List<ScoreHistoryDto> history = getScoreHistoryByUserIdUseCase.execute(user.getId());
        log.info("✅ スコア履歴取得成功 - userId: {}, 件数: {}", user.getId(), history.size());

        return ResponseEntity.ok(history);
    }

    private User resolveUser(Jwt jwt) {
        return userIdentityService.findUserBySub(jwt.getSubject());
    }
}
