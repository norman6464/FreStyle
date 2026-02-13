package com.example.FreStyle.controller;

import java.util.Map;
import java.util.Optional;

import org.springframework.messaging.handler.annotation.MessageMapping;
import org.springframework.messaging.handler.annotation.Payload;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.stereotype.Controller;

import com.example.FreStyle.dto.ChatMessageDto;
import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.RoomMemberRepository;
import com.example.FreStyle.service.ChatMessageService;
import com.example.FreStyle.service.ChatRoomService;
import com.example.FreStyle.service.UnreadCountService;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@Controller
@RequiredArgsConstructor
@Slf4j
public class ChatWebSocketController {

    private final ChatRoomService chatRoomService;
    private final ChatMessageService chatMessageService;
    private final SimpMessagingTemplate messagingTemplate;
    private final UnreadCountService unreadCountService;
    private final RoomMemberRepository roomMemberRepository;

    @MessageMapping("/chat/send")
    public void sendMessage(
            @Payload Map<String, Object> payload
    ) {
        log.info("\n========== WebSocket /chat/send リクエスト受信 ==========");
        log.info("📨 ペイロード全体: " + payload);
        
        try {
            // パラメータの取得と検証
            log.info("🔍 パラメータを抽出中...");
            Object senderIdObj = payload.get("senderId");
            Object roomIdObj = payload.get("roomId");
            Object contentObj = payload.get("content");
            
            log.debug("   - senderId タイプ: " + (senderIdObj != null ? senderIdObj.getClass().getSimpleName() : "null"));
            log.debug("   - senderId 値: " + senderIdObj);
            log.debug("   - roomId タイプ: " + (roomIdObj != null ? roomIdObj.getClass().getSimpleName() : "null"));
            log.debug("   - roomId 値: " + roomIdObj);
            log.debug("   - content タイプ: " + (contentObj != null ? contentObj.getClass().getSimpleName() : "null"));
            log.debug("   - content 値: " + contentObj);
            
            // senderId は String または Integer で来る可能性がある
            // Integer に変換して扱う
            Integer senderId;
            if (senderIdObj instanceof Integer) {
                // Object型をIntegerに変換をしてから格納をする
                senderId = (Integer) senderIdObj;
                log.info("   💡 senderId を Integer から String に変換");
            } else {
                senderId = (Integer) senderIdObj;
            }
            
            // roomId は String または Integer で来る可能性がある
            Integer roomId;
            if (roomIdObj instanceof Integer) {
                roomId = (Integer) roomIdObj;
                log.info("   💡 roomId を Integer として取得");
            } else {
                roomId = Integer.parseInt((String) roomIdObj);
                log.info("   💡 roomId を String から Integer に変換");
            }
            
            String content = (String) payload.get("content");
            
            log.info("✅ パラメータ抽出成功");
            log.debug("   - senderId (最終): " + senderId + " (タイプ: String)");
            log.debug("   - roomId (最終): " + roomId + " (タイプ: Integer)");
            log.debug("   - content: " + content);
            
            // ChatRoom 取得
            log.info("🔍 ChatRoom を roomId=" + roomId + " で取得中...");
            ChatRoom room = chatRoomService.findChatRoomById(roomId);
            log.info("✅ ChatRoom 取得成功: " + room.getId());
            
            // メッセージ保存
            log.info("💾 メッセージをデータベースに保存中...");
            ChatMessageDto saved = chatMessageService.addMessage(room, senderId, content);
            log.info("✅ メッセージ保存成功");
            log.debug("   - messageId: " + saved.getId());
            log.debug("   - roomId: " + saved.getRoomId());
            log.debug("   - senderId: " + saved.getSenderId());
            log.debug("   - content: " + saved.getContent());
            log.debug("   - createdAt: " + saved.getCreatedAt());

            // WebSocket トピックへ送信
            log.info("📤 WebSocket トピック /topic/chat/" + room.getId() + " へメッセージを送信中...");
            messagingTemplate.convertAndSend(
                    "/topic/chat/" + room.getId(),
                    saved
            );
            log.info("✅ WebSocket 送信完了");

            // 相手ユーザーの未読数をインクリメントし、WebSocketで通知
            Optional<User> partnerOpt = roomMemberRepository.findPartnerByRoomIdAndUserId(roomId, senderId);
            if (partnerOpt.isPresent()) {
                User partner = partnerOpt.get();
                unreadCountService.incrementUnreadCount(partner.getId(), roomId);
                messagingTemplate.convertAndSend(
                        "/topic/unread/" + partner.getId(),
                        Map.of(
                                "type", "unread_update",
                                "roomId", roomId,
                                "increment", 1
                        )
                );
                log.info("📤 未読数通知を /topic/unread/" + partner.getId() + " へ送信");
            }

            log.info("========== /chat/send 処理完了 ==========\n");
            
        } catch (NumberFormatException e) {
            log.error("❌ 型変換エラー発生");
            log.info("   エラーメッセージ: " + e.getMessage());
            log.info("   roomId を Integer に変換できませんでした");
            log.info("========== /chat/send 処理失敗 ==========\n");
        } catch (NullPointerException e) {
            log.error("❌ NullPointerException 発生");
            log.info("   ペイロードに必須パラメータが不足しています");
            log.info("   必須: senderId, roomId, content");
            log.info("   エラーメッセージ: " + e.getMessage());
            log.info("========== /chat/send 処理失敗 ==========\n");
        } catch (Exception e) {
            log.error("❌ 予期しないエラー発生");
            log.info("   エラータイプ: " + e.getClass().getName());
            log.info("   エラーメッセージ: " + e.getMessage());
            log.info("========== /chat/send 処理失敗 ==========\n");
        }
    }

    @MessageMapping("/chat/delete")
    public void deleteMessage(
            @Payload Map<String, Object> payload
    ) {
        log.info("\n========== WebSocket /chat/delete リクエスト受信 ==========");
        log.info("🗑️ ペイロード全体: " + payload);
        
        try {
            // パラメータの取得と検証
            log.info("🔍 パラメータを抽出中...");
            Object messageIdObj = payload.get("messageId");
            Object roomIdObj = payload.get("roomId");
            
            log.debug("   - messageId タイプ: " + (messageIdObj != null ? messageIdObj.getClass().getSimpleName() : "null"));
            log.debug("   - messageId 値: " + messageIdObj);
            log.debug("   - roomId タイプ: " + (roomIdObj != null ? roomIdObj.getClass().getSimpleName() : "null"));
            log.debug("   - roomId 値: " + roomIdObj);
            
            Integer messageId = ((Number) payload.get("messageId")).intValue();
            Integer roomId = ((Number) payload.get("roomId")).intValue();
            
            log.info("✅ パラメータ抽出成功");
            log.debug("   - messageId (最終): " + messageId);
            log.debug("   - roomId (最終): " + roomId);
            
            // メッセージ削除
            log.info("🔍 messageId=" + messageId + " を削除中...");
            chatMessageService.deleteMessage(messageId);
            log.info("✅ メッセージ削除成功");
            
            // WebSocket トピックへ削除通知を送信
            log.info("📤 WebSocket トピック /topic/chat/" + roomId + " へ削除通知を送信中...");
            messagingTemplate.convertAndSend(
                    "/topic/chat/" + roomId,
                    Map.of(
                        "type", "delete",
                        "messageId", messageId
                    )
            );
            log.info("✅ WebSocket 送信完了");
            log.info("========== /chat/delete 処理完了 ==========\n");
            
        } catch (NumberFormatException e) {
            log.error("❌ 型変換エラー発生");
            log.info("   エラーメッセージ: " + e.getMessage());
            log.info("   messageId または roomId を Integer に変換できませんでした");
            log.info("========== /chat/delete 処理失敗 ==========\n");
        } catch (NullPointerException e) {
            log.error("❌ NullPointerException 発生");
            log.info("   ペイロードに必須パラメータが不足しています");
            log.info("   必須: messageId, roomId");
            log.info("   エラーメッセージ: " + e.getMessage());
            log.info("========== /chat/delete 処理失敗 ==========\n");
        } catch (Exception e) {
            log.error("❌ 予期しないエラー発生");
            log.info("   エラータイプ: " + e.getClass().getName());
            log.info("   エラーメッセージ: " + e.getMessage());
            log.info("========== /chat/delete 処理失敗 ==========\n");
        }
    }
}