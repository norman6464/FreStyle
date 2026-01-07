package com.example.FreStyle.controller;

import java.util.List;
import java.util.Map;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
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
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.AiChatMessageService;
import com.example.FreStyle.service.AiChatService;
import com.example.FreStyle.service.AiChatSessionService;
import com.example.FreStyle.service.UserIdentityService;

import lombok.RequiredArgsConstructor;


@RestController 
@RequiredArgsConstructor
@RequestMapping("/api/chat/ai")
public class AiChatController {

    private static final Logger logger = LoggerFactory.getLogger(AiChatController.class);
    private final AiChatService aiChatService;
    private final AiChatSessionService aiChatSessionService;
    private final AiChatMessageService aiChatMessageService;
    private final UserIdentityService userIdentityService;


    // =============================================
    // 既存のDynamoDB履歴取得API（後方互換性のため維持）
    // =============================================

    @GetMapping("/history")
    public ResponseEntity<?> getChatHistory(@AuthenticationPrincipal Jwt jwt) {
        try {
            // Jwt から senderId(sub) を取得
            String sub = jwt.getSubject();
            
            Integer senderId = userIdentityService.findUserBySub(sub).getId();
            logger.info("📥 [AiChatController] AI履歴取得リクエスト開始 - senderId: {}", senderId);
            
            // ロジックは変更しない
            logger.debug("🔍 [AiChatController] AiChatService.getChatHistory() を呼び出し");
            List<AiChatMessageDto> history = aiChatService.getChatHistory(senderId);
            
            logger.info("✅ [AiChatController] AI履歴取得成功 - メッセージ数: {}", history.size());
            logger.debug("📋 [AiChatController] 取得履歴: {}", history);

            return ResponseEntity.ok(history);

        } catch (RuntimeException e) {
            // 予期しないアプリケーションエラー → 500
            logger.error("❌ [AiChatController] AI履歴取得エラー", e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body(Map.of("error", "サーバーのエラーです。"));
        }
    }

    // =============================================
    // 新規RDBベースのセッション管理API
    // =============================================

    /**
     * セッション一覧を取得
     */
    @GetMapping("/sessions")
    public ResponseEntity<List<AiChatSessionDto>> getSessions(@AuthenticationPrincipal Jwt jwt) {
        logger.info("========== GET /api/chat/ai/sessions ==========");
        
        try {
            String sub = jwt.getSubject();
            User user = userIdentityService.findUserBySub(sub);
            
            List<AiChatSessionDto> sessions = aiChatSessionService.getSessionsByUserId(user.getId());
            logger.info("✅ セッション一覧取得成功 - 件数: {}", sessions.size());
            
            return ResponseEntity.ok(sessions);
        } catch (Exception e) {
            logger.error("❌ セッション一覧取得エラー: {}", e.getMessage());
            return ResponseEntity.internalServerError().build();
        }
    }

    /**
     * 新規セッションを作成
     */
    @PostMapping("/sessions")
    public ResponseEntity<AiChatSessionDto> createSession(
            @AuthenticationPrincipal Jwt jwt,
            @RequestBody CreateSessionRequest request
    ) {
        logger.info("========== POST /api/chat/ai/sessions ==========");
        logger.info("📝 リクエスト: {}", request);
        
        try {
            String sub = jwt.getSubject();
            User user = userIdentityService.findUserBySub(sub);
            
            AiChatSessionDto session = aiChatSessionService.createSession(
                    user.getId(),
                    request.title(),
                    request.relatedRoomId()
            );
            logger.info("✅ セッション作成成功 - sessionId: {}", session.getId());
            
            return ResponseEntity.ok(session);
        } catch (Exception e) {
            logger.error("❌ セッション作成エラー: {}", e.getMessage());
            return ResponseEntity.internalServerError().build();
        }
    }

    /**
     * セッション詳細を取得
     */
    @GetMapping("/sessions/{sessionId}")
    public ResponseEntity<AiChatSessionDto> getSession(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer sessionId
    ) {
        logger.info("========== GET /api/chat/ai/sessions/{} ==========", sessionId);
        
        try {
            String sub = jwt.getSubject();
            User user = userIdentityService.findUserBySub(sub);
            
            AiChatSessionDto session = aiChatSessionService.getSessionByIdAndUserId(sessionId, user.getId());
            logger.info("✅ セッション取得成功");
            
            return ResponseEntity.ok(session);
        } catch (RuntimeException e) {
            logger.error("❌ セッション取得エラー: {}", e.getMessage());
            return ResponseEntity.notFound().build();
        }
    }

    /**
     * セッションタイトルを更新
     */
    @PutMapping("/sessions/{sessionId}")
    public ResponseEntity<AiChatSessionDto> updateSessionTitle(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer sessionId,
            @RequestBody UpdateSessionRequest request
    ) {
        logger.info("========== PUT /api/chat/ai/sessions/{} ==========", sessionId);
        
        try {
            String sub = jwt.getSubject();
            User user = userIdentityService.findUserBySub(sub);
            
            AiChatSessionDto session = aiChatSessionService.updateSessionTitle(
                    sessionId,
                    user.getId(),
                    request.title()
            );
            logger.info("✅ セッションタイトル更新成功");
            
            return ResponseEntity.ok(session);
        } catch (RuntimeException e) {
            logger.error("❌ セッションタイトル更新エラー: {}", e.getMessage());
            return ResponseEntity.notFound().build();
        }
    }

    /**
     * セッションを削除
     */
    @DeleteMapping("/sessions/{sessionId}")
    public ResponseEntity<Void> deleteSession(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer sessionId
    ) {
        logger.info("========== DELETE /api/chat/ai/sessions/{} ==========", sessionId);
        
        try {
            String sub = jwt.getSubject();
            User user = userIdentityService.findUserBySub(sub);
            
            aiChatSessionService.deleteSession(sessionId, user.getId());
            logger.info("✅ セッション削除成功");
            
            return ResponseEntity.noContent().build();
        } catch (RuntimeException e) {
            logger.error("❌ セッション削除エラー: {}", e.getMessage());
            return ResponseEntity.notFound().build();
        }
    }

    /**
     * セッション内のメッセージ一覧を取得
     */
    @GetMapping("/sessions/{sessionId}/messages")
    public ResponseEntity<List<AiChatMessageResponseDto>> getMessages(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer sessionId
    ) {
        logger.info("========== GET /api/chat/ai/sessions/{}/messages ==========", sessionId);
        
        try {
            String sub = jwt.getSubject();
            User user = userIdentityService.findUserBySub(sub);
            
            // 権限チェック（セッションがユーザーのものか確認）
            aiChatSessionService.getSessionByIdAndUserId(sessionId, user.getId());
            
            List<AiChatMessageResponseDto> messages = aiChatMessageService.getMessagesBySessionId(sessionId);
            logger.info("✅ メッセージ一覧取得成功 - 件数: {}", messages.size());
            
            return ResponseEntity.ok(messages);
        } catch (RuntimeException e) {
            logger.error("❌ メッセージ一覧取得エラー: {}", e.getMessage());
            return ResponseEntity.notFound().build();
        }
    }

    /**
     * メッセージを追加（REST API経由）
     */
    @PostMapping("/sessions/{sessionId}/messages")
    public ResponseEntity<AiChatMessageResponseDto> addMessage(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer sessionId,
            @RequestBody AddMessageRequest request
    ) {
        logger.info("========== POST /api/chat/ai/sessions/{}/messages ==========", sessionId);
        logger.info("📝 リクエスト: {}", request);
        
        try {
            String sub = jwt.getSubject();
            User user = userIdentityService.findUserBySub(sub);
            
            // 権限チェック
            aiChatSessionService.getSessionByIdAndUserId(sessionId, user.getId());
            
            AiChatMessageResponseDto message;
            if ("assistant".equalsIgnoreCase(request.role())) {
                message = aiChatMessageService.addAssistantMessage(sessionId, user.getId(), request.content());
            } else {
                message = aiChatMessageService.addUserMessage(sessionId, user.getId(), request.content());
            }
            
            logger.info("✅ メッセージ追加成功 - messageId: {}", message.getId());
            
            return ResponseEntity.ok(message);
        } catch (RuntimeException e) {
            logger.error("❌ メッセージ追加エラー: {}", e.getMessage());
            return ResponseEntity.badRequest().build();
        }
    }

    // リクエスト用のRecord
    record CreateSessionRequest(String title, Integer relatedRoomId) {}
    record UpdateSessionRequest(String title) {}
    record AddMessageRequest(String content, String role) {}
}
