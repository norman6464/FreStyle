package com.example.FreStyle.controller;

import java.util.List;
import java.util.Map;

import jakarta.validation.Valid;
import jakarta.validation.constraints.NotBlank;

import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.dto.AiChatMessageDto;
import com.example.FreStyle.dto.AiChatMessageResponseDto;
import com.example.FreStyle.dto.AiChatSessionDto;
import com.example.FreStyle.dto.AiChatSessionStatsDto;
import com.example.FreStyle.dto.PracticeSessionSummaryDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.AiChatService;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.AddAiChatMessageUseCase;
import com.example.FreStyle.usecase.CreateAiChatSessionUseCase;
import com.example.FreStyle.usecase.DeleteAiChatSessionUseCase;
import com.example.FreStyle.usecase.GetAiChatMessagesBySessionIdUseCase;
import com.example.FreStyle.usecase.GetAiChatSessionByIdUseCase;
import com.example.FreStyle.usecase.GetAiChatSessionsByUserIdUseCase;
import com.example.FreStyle.usecase.GetAiChatSessionStatsUseCase;
import com.example.FreStyle.usecase.GetPracticeSessionSummaryUseCase;
import com.example.FreStyle.usecase.RephraseMessageUseCase;
import com.example.FreStyle.usecase.UpdateAiChatSessionTitleUseCase;

import io.swagger.v3.oas.annotations.tags.Tag;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;


@RestController
@RequiredArgsConstructor
@RequestMapping("/api/chat/ai")
@Slf4j
@Tag(name = "AI Chat", description = "AI 対話セッションとメッセージの管理 API")
public class AiChatController {
    private final AiChatService aiChatService;
    private final UserIdentityService userIdentityService;

    // UseCases (クリーンアーキテクチャー)
    private final GetAiChatSessionsByUserIdUseCase getAiChatSessionsByUserIdUseCase;
    private final CreateAiChatSessionUseCase createAiChatSessionUseCase;
    private final GetAiChatSessionByIdUseCase getAiChatSessionByIdUseCase;
    private final UpdateAiChatSessionTitleUseCase updateAiChatSessionTitleUseCase;
    private final DeleteAiChatSessionUseCase deleteAiChatSessionUseCase;
    private final GetAiChatMessagesBySessionIdUseCase getAiChatMessagesBySessionIdUseCase;
    private final AddAiChatMessageUseCase addAiChatMessageUseCase;
    private final GetPracticeSessionSummaryUseCase getPracticeSessionSummaryUseCase;
    private final GetAiChatSessionStatsUseCase getAiChatSessionStatsUseCase;
    private final RephraseMessageUseCase rephraseMessageUseCase;


    // =============================================
    // 既存のDynamoDB履歴取得API（後方互換性のため維持）
    // =============================================

    @GetMapping("/history")
    public ResponseEntity<?> getChatHistory(@AuthenticationPrincipal Jwt jwt) {
        User user = resolveUser(jwt);
        log.info("📥 [AiChatController] AI履歴取得リクエスト開始 - senderId: {}", user.getId());

        log.debug("🔍 [AiChatController] AiChatService.getChatHistory() を呼び出し");
        List<AiChatMessageDto> history = aiChatService.getChatHistory(user.getId());

        log.info("✅ [AiChatController] AI履歴取得成功 - メッセージ数: {}", history.size());
        log.debug("📋 [AiChatController] 取得履歴: {}", history);

        return ResponseEntity.ok(history);
    }

    // =============================================
    // 新規RDBベースのセッション管理API
    // =============================================

    /**
     * セッション一覧を取得
     */
    @GetMapping("/sessions")
    public ResponseEntity<List<AiChatSessionDto>> getSessions(@AuthenticationPrincipal Jwt jwt) {
        log.info("========== GET /api/chat/ai/sessions ==========");

        User user = resolveUser(jwt);
        List<AiChatSessionDto> sessions = getAiChatSessionsByUserIdUseCase.execute(user.getId());
        log.info("✅ セッション一覧取得成功 - 件数: {}", sessions.size());

        return ResponseEntity.ok(sessions);
    }

    /**
     * 新規セッションを作成
     */
    @PostMapping("/sessions")
    public ResponseEntity<AiChatSessionDto> createSession(
            @AuthenticationPrincipal Jwt jwt,
            @Valid @RequestBody CreateSessionRequest request
    ) {
        log.info("========== POST /api/chat/ai/sessions ==========");
        log.info("📝 リクエスト: {}", request);

        User user = resolveUser(jwt);
        AiChatSessionDto session = createAiChatSessionUseCase.execute(
                user.getId(),
                request.title(),
                request.relatedRoomId()
        );
        log.info("✅ セッション作成成功 - sessionId: {}", session.id());

        return ResponseEntity.ok(session);
    }

    /**
     * セッション詳細を取得
     */
    @GetMapping("/sessions/{sessionId}")
    public ResponseEntity<AiChatSessionDto> getSession(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer sessionId
    ) {
        log.info("========== GET /api/chat/ai/sessions/{} ==========", sessionId);

        User user = resolveUser(jwt);
        AiChatSessionDto session = getAiChatSessionByIdUseCase.execute(sessionId, user.getId());
        log.info("✅ セッション取得成功");

        return ResponseEntity.ok(session);
    }

    /**
     * セッションタイトルを更新
     */
    @PutMapping("/sessions/{sessionId}")
    public ResponseEntity<AiChatSessionDto> updateSessionTitle(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer sessionId,
            @Valid @RequestBody UpdateSessionRequest request
    ) {
        log.info("========== PUT /api/chat/ai/sessions/{} ==========", sessionId);

        User user = resolveUser(jwt);
        AiChatSessionDto session = updateAiChatSessionTitleUseCase.execute(
                sessionId,
                user.getId(),
                request.title()
        );
        log.info("✅ セッションタイトル更新成功");

        return ResponseEntity.ok(session);
    }

    /**
     * セッションを削除
     */
    @DeleteMapping("/sessions/{sessionId}")
    public ResponseEntity<Void> deleteSession(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer sessionId
    ) {
        log.info("========== DELETE /api/chat/ai/sessions/{} ==========", sessionId);

        User user = resolveUser(jwt);
        deleteAiChatSessionUseCase.execute(sessionId, user.getId());
        log.info("✅ セッション削除成功");

        return ResponseEntity.noContent().build();
    }

    /**
     * セッション内のメッセージ一覧を取得
     */
    @GetMapping("/sessions/{sessionId}/messages")
    public ResponseEntity<List<AiChatMessageResponseDto>> getMessages(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer sessionId
    ) {
        log.info("========== GET /api/chat/ai/sessions/{}/messages ==========", sessionId);

        User user = resolveUser(jwt);
        // 権限チェック（セッションがユーザーのものか確認）
        getAiChatSessionByIdUseCase.execute(sessionId, user.getId());

        List<AiChatMessageResponseDto> messages = getAiChatMessagesBySessionIdUseCase.execute(sessionId);
        log.info("✅ メッセージ一覧取得成功 - 件数: {}", messages.size());

        return ResponseEntity.ok(messages);
    }

    /**
     * メッセージを追加（REST API経由）
     */
    @PostMapping("/sessions/{sessionId}/messages")
    public ResponseEntity<AiChatMessageResponseDto> addMessage(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer sessionId,
            @Valid @RequestBody AddMessageRequest request
    ) {
        log.info("========== POST /api/chat/ai/sessions/{}/messages ==========", sessionId);
        log.info("📝 リクエスト: {}", request);

        User user = resolveUser(jwt);
        // 権限チェック
        getAiChatSessionByIdUseCase.execute(sessionId, user.getId());

        AiChatMessageResponseDto message;
        if ("assistant".equalsIgnoreCase(request.role())) {
            message = addAiChatMessageUseCase.executeAssistantMessage(sessionId, user.getId(), request.content());
        } else {
            message = addAiChatMessageUseCase.executeUserMessage(sessionId, user.getId(), request.content());
        }

        log.info("✅ メッセージ追加成功 - messageId: {}", message.id());

        return ResponseEntity.ok(message);
    }

    /**
     * セッションサマリーを取得
     */
    @GetMapping("/sessions/{sessionId}/summary")
    public ResponseEntity<PracticeSessionSummaryDto> getSessionSummary(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer sessionId
    ) {
        log.info("========== GET /api/chat/ai/sessions/{}/summary ==========", sessionId);

        User user = resolveUser(jwt);
        PracticeSessionSummaryDto summary = getPracticeSessionSummaryUseCase.execute(sessionId, user.getId());
        log.info("✅ セッションサマリー取得成功");

        return ResponseEntity.ok(summary);
    }

    /**
     * メッセージの言い換え提案を取得（REST API）
     */
    @PostMapping("/rephrase")
    public ResponseEntity<Map<String, String>> rephrase(
            @AuthenticationPrincipal Jwt jwt,
            @Valid @RequestBody RephraseRequest request
    ) {
        log.info("========== POST /api/chat/ai/rephrase ==========");

        resolveUser(jwt); // 認証チェック

        String result = rephraseMessageUseCase.execute(request.originalMessage(), request.scene());
        log.info("✅ 言い換え提案取得成功");

        return ResponseEntity.ok(Map.of("result", result));
    }

    /**
     * セッション統計を取得
     */
    @GetMapping("/session-stats")
    public ResponseEntity<AiChatSessionStatsDto> getSessionStats(
            @AuthenticationPrincipal Jwt jwt) {
        log.info("========== GET /api/chat/ai/session-stats ==========");

        User user = resolveUser(jwt);
        AiChatSessionStatsDto result = getAiChatSessionStatsUseCase.execute(user.getId());
        log.info("✅ セッション統計取得成功 - userId={}", user.getId());

        return ResponseEntity.ok(result);
    }

    private User resolveUser(Jwt jwt) {
        return userIdentityService.findUserBySub(jwt.getSubject());
    }

    // リクエスト用のRecord
    record CreateSessionRequest(String title, Integer relatedRoomId) {}
    record UpdateSessionRequest(@NotBlank String title) {}
    record AddMessageRequest(@NotBlank String content, @NotBlank String role) {}
    record RephraseRequest(@NotBlank String originalMessage, @NotBlank String scene) {}
}
