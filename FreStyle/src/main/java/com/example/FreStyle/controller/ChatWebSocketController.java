package com.example.FreStyle.controller;

import java.util.Map;

import org.springframework.messaging.handler.annotation.MessageMapping;
import org.springframework.messaging.handler.annotation.Payload;
import org.springframework.messaging.simp.SimpMessageHeaderAccessor;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.stereotype.Controller;

import com.example.FreStyle.config.WebSocketAuthHandshakeInterceptor;
import com.example.FreStyle.usecase.DeleteChatMessageUseCase;
import com.example.FreStyle.usecase.SendChatMessageUseCase;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@Controller
@RequiredArgsConstructor
@Slf4j
public class ChatWebSocketController {

    private final SendChatMessageUseCase sendChatMessageUseCase;
    private final DeleteChatMessageUseCase deleteChatMessageUseCase;
    private final SimpMessagingTemplate messagingTemplate;

    @MessageMapping("/chat/send")
    public void sendMessage(@Payload Map<String, Object> payload, SimpMessageHeaderAccessor headerAccessor) {
        try {
            Integer senderId = getAuthenticatedUserId(headerAccessor);
            if (senderId == null) {
                log.warn("WebSocket認証エラー: 認証されていないユーザーからのメッセージ送信");
                return;
            }
            Integer roomId = ((Number) payload.get("roomId")).intValue();
            String content = (String) payload.get("content");

            SendChatMessageUseCase.Result result = sendChatMessageUseCase.execute(senderId, roomId, content);

            messagingTemplate.convertAndSend("/topic/chat/" + roomId, result.message());

            if (result.partnerId() != null) {
                messagingTemplate.convertAndSend(
                        "/topic/unread/" + result.partnerId(),
                        Map.of(
                                "type", "unread_update",
                                "roomId", roomId,
                                "increment", 1
                        )
                );
            }
        } catch (Exception e) {
            log.error("メッセージ送信エラー: {}", e.getMessage(), e);
        }
    }

    @MessageMapping("/chat/delete")
    public void deleteMessage(@Payload Map<String, Object> payload) {
        try {
            Integer roomId = ((Number) payload.get("roomId")).intValue();
            Long createdAt = ((Number) payload.get("createdAt")).longValue();

            deleteChatMessageUseCase.execute(roomId, createdAt);

            messagingTemplate.convertAndSend(
                    "/topic/chat/" + roomId,
                    Map.of(
                            "type", "delete",
                            "roomId", roomId,
                            "createdAt", createdAt
                    )
            );
        } catch (Exception e) {
            log.error("メッセージ削除エラー: {}", e.getMessage(), e);
        }
    }

    private Integer getAuthenticatedUserId(SimpMessageHeaderAccessor headerAccessor) {
        Map<String, Object> sessionAttributes = headerAccessor.getSessionAttributes();
        if (sessionAttributes == null) return null;
        return (Integer) sessionAttributes.get(WebSocketAuthHandshakeInterceptor.AUTHENTICATED_USER_ID);
    }
}
