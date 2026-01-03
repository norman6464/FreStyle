package com.example.FreStyle.controller;

import java.util.Map;

import org.springframework.messaging.handler.annotation.MessageMapping;
import org.springframework.messaging.handler.annotation.Payload;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.stereotype.Controller;

import com.example.FreStyle.dto.ChatMessageDto;
import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.service.ChatMessageService;
import com.example.FreStyle.service.ChatRoomService;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@Controller
@RequiredArgsConstructor
@Slf4j
public class ChatWebSocketController {

    private final ChatRoomService chatRoomService;
    private final ChatMessageService chatMessageService;
    private final SimpMessagingTemplate messagingTemplate;

    @MessageMapping("/chat/send")
    public void sendMessage(
            @Payload Map<String, Object> payload
    ) {
        System.out.println("\n========== WebSocket /chat/send リクエスト受信 ==========");
        System.out.println("📨 ペイロード全体: " + payload);
        
        try {
            // パラメータの取得と検証
            System.out.println("🔍 パラメータを抽出中...");
            Object senderIdObj = payload.get("senderId");
            Object roomIdObj = payload.get("roomId");
            Object contentObj = payload.get("content");
            
            System.out.println("   - senderId タイプ: " + (senderIdObj != null ? senderIdObj.getClass().getSimpleName() : "null"));
            System.out.println("   - senderId 値: " + senderIdObj);
            System.out.println("   - roomId タイプ: " + (roomIdObj != null ? roomIdObj.getClass().getSimpleName() : "null"));
            System.out.println("   - roomId 値: " + roomIdObj);
            System.out.println("   - content タイプ: " + (contentObj != null ? contentObj.getClass().getSimpleName() : "null"));
            System.out.println("   - content 値: " + contentObj);
            
            // senderId は String または Integer で来る可能性がある
            // Integer に変換して扱う
            Integer senderId;
            if (senderIdObj instanceof Integer) {
                // Object型をIntegerに変換をしてから格納をする
                senderId = (Integer) senderIdObj;
                System.out.println("   💡 senderId を Integer から String に変換");
            } else {
                senderId = (Integer) senderIdObj;
            }
            
            // roomId は String または Integer で来る可能性がある
            Integer roomId;
            if (roomIdObj instanceof Integer) {
                roomId = (Integer) roomIdObj;
                System.out.println("   💡 roomId を Integer として取得");
            } else {
                roomId = Integer.parseInt((String) roomIdObj);
                System.out.println("   💡 roomId を String から Integer に変換");
            }
            
            String content = (String) payload.get("content");
            
            System.out.println("✅ パラメータ抽出成功");
            System.out.println("   - senderId (最終): " + senderId + " (タイプ: String)");
            System.out.println("   - roomId (最終): " + roomId + " (タイプ: Integer)");
            System.out.println("   - content: " + content);
            
            // ChatRoom 取得
            System.out.println("🔍 ChatRoom を roomId=" + roomId + " で取得中...");
            ChatRoom room = chatRoomService.findChatRoomById(roomId);
            System.out.println("✅ ChatRoom 取得成功: " + room.getId());
            
            // メッセージ保存
            System.out.println("💾 メッセージをデータベースに保存中...");
            ChatMessageDto saved = chatMessageService.addMessage(room, senderId, content);
            System.out.println("✅ メッセージ保存成功");
            System.out.println("   - messageId: " + saved.getId());
            System.out.println("   - roomId: " + saved.getRoomId());
            System.out.println("   - senderId: " + saved.getSenderId());
            System.out.println("   - content: " + saved.getContent());
            System.out.println("   - createdAt: " + saved.getCreatedAt());

            // WebSocket トピックへ送信
            System.out.println("📤 WebSocket トピック /topic/chat/" + room.getId() + " へメッセージを送信中...");
            messagingTemplate.convertAndSend(
                    "/topic/chat/" + room.getId(),
                    saved
            );
            System.out.println("✅ WebSocket 送信完了");
            System.out.println("========== /chat/send 処理完了 ==========\n");
            
        } catch (NumberFormatException e) {
            System.out.println("❌ 型変換エラー発生");
            System.out.println("   エラーメッセージ: " + e.getMessage());
            System.out.println("   roomId を Integer に変換できませんでした");
            e.printStackTrace();
            System.out.println("========== /chat/send 処理失敗 ==========\n");
        } catch (NullPointerException e) {
            System.out.println("❌ NullPointerException 発生");
            System.out.println("   ペイロードに必須パラメータが不足しています");
            System.out.println("   必須: senderId, roomId, content");
            System.out.println("   エラーメッセージ: " + e.getMessage());
            e.printStackTrace();
            System.out.println("========== /chat/send 処理失敗 ==========\n");
        } catch (Exception e) {
            System.out.println("❌ 予期しないエラー発生");
            System.out.println("   エラータイプ: " + e.getClass().getName());
            System.out.println("   エラーメッセージ: " + e.getMessage());
            e.printStackTrace();
            System.out.println("========== /chat/send 処理失敗 ==========\n");
        }
    }

    @MessageMapping("/chat/delete")
    public void deleteMessage(
            @Payload Map<String, Object> payload
    ) {
        System.out.println("\n========== WebSocket /chat/delete リクエスト受信 ==========");
        System.out.println("🗑️ ペイロード全体: " + payload);
        
        try {
            // パラメータの取得と検証
            System.out.println("🔍 パラメータを抽出中...");
            Object messageIdObj = payload.get("messageId");
            Object roomIdObj = payload.get("roomId");
            
            System.out.println("   - messageId タイプ: " + (messageIdObj != null ? messageIdObj.getClass().getSimpleName() : "null"));
            System.out.println("   - messageId 値: " + messageIdObj);
            System.out.println("   - roomId タイプ: " + (roomIdObj != null ? roomIdObj.getClass().getSimpleName() : "null"));
            System.out.println("   - roomId 値: " + roomIdObj);
            
            Integer messageId = ((Number) payload.get("messageId")).intValue();
            Integer roomId = ((Number) payload.get("roomId")).intValue();
            
            System.out.println("✅ パラメータ抽出成功");
            System.out.println("   - messageId (最終): " + messageId);
            System.out.println("   - roomId (最終): " + roomId);
            
            // メッセージ削除
            System.out.println("🔍 messageId=" + messageId + " を削除中...");
            chatMessageService.deleteMessage(messageId);
            System.out.println("✅ メッセージ削除成功");
            
            // WebSocket トピックへ削除通知を送信
            System.out.println("📤 WebSocket トピック /topic/chat/" + roomId + " へ削除通知を送信中...");
            messagingTemplate.convertAndSend(
                    "/topic/chat/" + roomId,
                    Map.of(
                        "type", "delete",
                        "messageId", messageId
                    )
            );
            System.out.println("✅ WebSocket 送信完了");
            System.out.println("========== /chat/delete 処理完了 ==========\n");
            
        } catch (NumberFormatException e) {
            System.out.println("❌ 型変換エラー発生");
            System.out.println("   エラーメッセージ: " + e.getMessage());
            System.out.println("   messageId または roomId を Integer に変換できませんでした");
            e.printStackTrace();
            System.out.println("========== /chat/delete 処理失敗 ==========\n");
        } catch (NullPointerException e) {
            System.out.println("❌ NullPointerException 発生");
            System.out.println("   ペイロードに必須パラメータが不足しています");
            System.out.println("   必須: messageId, roomId");
            System.out.println("   エラーメッセージ: " + e.getMessage());
            e.printStackTrace();
            System.out.println("========== /chat/delete 処理失敗 ==========\n");
        } catch (Exception e) {
            System.out.println("❌ 予期しないエラー発生");
            System.out.println("   エラータイプ: " + e.getClass().getName());
            System.out.println("   エラーメッセージ: " + e.getMessage());
            e.printStackTrace();
            System.out.println("========== /chat/delete 処理失敗 ==========\n");
        }
    }
}