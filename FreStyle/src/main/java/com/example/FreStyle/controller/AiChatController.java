package com.example.FreStyle.controller;

import java.util.List;
import java.util.Map;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.dto.AiChatMessageDto;
import com.example.FreStyle.service.AiChatService;

@RestController 
@RequestMapping("/api/chat/ai")
public class AiChatController {

    private final AiChatService aiChatService;

    public AiChatController(AiChatService aiChatService) {
        this.aiChatService = aiChatService;
    }

    @GetMapping("/history")
    public ResponseEntity<?> getChatHistory(@AuthenticationPrincipal Jwt jwt) {
        try {
            // Jwt から senderId(sub) を取得
            String senderId = jwt.getSubject();

            // ロジックは変更しない
            List<AiChatMessageDto> history = aiChatService.getChatHistory(senderId);

            return ResponseEntity.ok(history);

        } catch (RuntimeException e) {
            // 予期しないアプリケーションエラー → 500
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body(Map.of("error", "サーバーのエラーです。"));
        }
    }
}
