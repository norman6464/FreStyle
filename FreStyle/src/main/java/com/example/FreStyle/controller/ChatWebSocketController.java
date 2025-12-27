package com.example.FreStyle.controller;

import java.util.Map;

import org.springframework.messaging.handler.annotation.MessageMapping;
import org.springframework.messaging.handler.annotation.Payload;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.stereotype.Controller;

import com.example.FreStyle.dto.ChatMessageDto;
import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.ChatMessageService;
import com.example.FreStyle.service.ChatRoomService;
import com.example.FreStyle.service.UserIdentityService;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@Controller
@RequiredArgsConstructor
@Slf4j
public class ChatWebSocketController {

    private final ChatRoomService chatRoomService;
    private final ChatMessageService chatMessageService;
    private final UserIdentityService userIdentityService;
    private final SimpMessagingTemplate messagingTemplate;

    @MessageMapping("/chat/send")
    public void sendMessage(
            @Payload Map<String, Object> payload,
            @AuthenticationPrincipal Jwt jwt
    ) {
        try {
            log.info("📨 Received message: {}", payload);
            
            // JWT → User
            String sub = jwt.getSubject();
            User sender = userIdentityService.findUserBySub(sub);
            log.info("Sender: {}", sender.getName());

            Integer roomId = ((Number) payload.get("roomId")).intValue();
            String content = (String) payload.get("content");
            
            ChatRoom room = chatRoomService.findChatRoomById(roomId);
            log.info("Room: {}", room.getId());

            ChatMessageDto saved = chatMessageService.addMessage(room, sender, content);
            log.info("Message saved: {}", saved.getId());

            messagingTemplate.convertAndSend(
                    "/topic/chat/" + room.getId(),
                    saved
            );
            log.info("✅ Message sent to topic");
        } catch (Exception e) {
            log.error("Error in sendMessage", e);
        }
    }

    @MessageMapping("/chat/delete")
    public void deleteMessage(
            @Payload Map<String, Object> payload,
            @AuthenticationPrincipal Jwt jwt
    ) {
        try {
            log.info("🗑️ Delete request: {}", payload);
            
            Integer messageId = ((Number) payload.get("messageId")).intValue();
            Integer roomId = ((Number) payload.get("roomId")).intValue();
            
            chatMessageService.deleteMessage(messageId);
            
            messagingTemplate.convertAndSend(
                    "/topic/chat/" + roomId,
                    Map.of(
                        "type", "delete",
                        "messageId", messageId
                    )
            );
            log.info("✅ Delete message sent to topic");
        } catch (Exception e) {
            log.error("Error in deleteMessage", e);
        }
    }
}

