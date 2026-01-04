package com.example.FreStyle.controller;

import java.util.List;
import java.util.Map;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.dto.AiChatMessageDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.AiChatService;
import com.example.FreStyle.service.UserIdentityService;

import lombok.RequiredArgsConstructor;

@RestController 
@RequiredArgsConstructor
@RequestMapping("/api/chat/ai")
public class AiChatController {

    private static final Logger logger = LoggerFactory.getLogger(AiChatController.class);
    private final AiChatService aiChatService;
    private final UserIdentityService userIdentityService;

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
}
