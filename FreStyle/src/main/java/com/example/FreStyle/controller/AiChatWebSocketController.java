package com.example.FreStyle.controller;

import java.util.Map;

import org.springframework.messaging.handler.annotation.MessageMapping;
import org.springframework.messaging.handler.annotation.Payload;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.stereotype.Controller;

import com.example.FreStyle.dto.AiChatMessageResponseDto;
import com.example.FreStyle.dto.AiChatSessionDto;
import com.example.FreStyle.entity.AiChatMessage.Role;
import com.example.FreStyle.service.AiChatMessageService;
import com.example.FreStyle.service.AiChatSessionService;
import com.example.FreStyle.service.BedrockService;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@Controller
@RequiredArgsConstructor
@Slf4j
public class AiChatWebSocketController {

    private final AiChatSessionService aiChatSessionService;
    private final AiChatMessageService aiChatMessageService;
    private final BedrockService bedrockService;
    private final SimpMessagingTemplate messagingTemplate;

    /**
     * AIチャットメッセージ送信
     * クライアントから /app/ai-chat/send へメッセージを送信
     */
    @MessageMapping("/ai-chat/send")
    public void sendMessage(@Payload Map<String, Object> payload) {
        System.out.println("\n========== WebSocket /ai-chat/send リクエスト受信 ==========");
        System.out.println("📨 ペイロード全体: " + payload);

        try {
            // パラメータの取得と検証
            System.out.println("🔍 パラメータを抽出中...");
            Object userIdObj = payload.get("userId");
            Object sessionIdObj = payload.get("sessionId");
            Object contentObj = payload.get("content");
            Object roleObj = payload.get("role"); // "user" または "assistant"

            System.out.println("   - userId タイプ: " + (userIdObj != null ? userIdObj.getClass().getSimpleName() : "null"));
            System.out.println("   - userId 値: " + userIdObj);
            System.out.println("   - sessionId タイプ: " + (sessionIdObj != null ? sessionIdObj.getClass().getSimpleName() : "null"));
            System.out.println("   - sessionId 値: " + sessionIdObj);
            System.out.println("   - content: " + contentObj);
            System.out.println("   - role: " + roleObj);

            // userId の変換
            Integer userId = convertToInteger(userIdObj);

            // sessionId の変換（新規セッションの場合はnull）
            Integer sessionId = sessionIdObj != null ? convertToInteger(sessionIdObj) : null;

            String content = (String) contentObj;
            String roleStr = roleObj != null ? (String) roleObj : "user";
            Role role = "assistant".equalsIgnoreCase(roleStr) ? Role.assistant : Role.user;

            System.out.println("✅ パラメータ抽出成功");
            System.out.println("   - userId (最終): " + userId);
            System.out.println("   - sessionId (最終): " + sessionId);
            System.out.println("   - content: " + content);
            System.out.println("   - role: " + role);

            // セッションが存在しない場合は新規作成
            if (sessionId == null) {
                System.out.println("🆕 新規セッション作成中...");
                // タイトルはメッセージの最初の30文字を使用
                String title = content.length() > 30 ? content.substring(0, 30) + "..." : content;
                AiChatSessionDto newSession = aiChatSessionService.createSession(userId, title, null);
                sessionId = newSession.getId();
                System.out.println("✅ 新規セッション作成完了 - sessionId: " + sessionId);

                // 新しいセッション情報をクライアントに通知
                messagingTemplate.convertAndSend(
                        "/topic/ai-chat/user/" + userId + "/session",
                        newSession
                );
            }

            // メッセージ保存（ユーザーメッセージ）
            System.out.println("💾 ユーザーメッセージをデータベースに保存中...");
            AiChatMessageResponseDto savedUserMessage = aiChatMessageService.addMessage(sessionId, userId, role, content);
            System.out.println("✅ ユーザーメッセージ保存成功");
            System.out.println("   - messageId: " + savedUserMessage.getId());
            System.out.println("   - sessionId: " + savedUserMessage.getSessionId());
            System.out.println("   - role: " + savedUserMessage.getRole());

            // WebSocket トピックへユーザーメッセージを送信
            System.out.println("📤 WebSocket トピック /topic/ai-chat/session/" + sessionId + " へユーザーメッセージを送信中...");
            messagingTemplate.convertAndSend(
                    "/topic/ai-chat/session/" + sessionId,
                    savedUserMessage
            );
            System.out.println("✅ ユーザーメッセージ WebSocket 送信完了");

            // Bedrockにメッセージを送信してAI応答を取得
            System.out.println("🤖 Bedrock にメッセージを送信中...");
            String aiReply = bedrockService.chat(content);
            System.out.println("✅ Bedrock から応答を取得しました");
            System.out.println("   - AI Reply: " + (aiReply.length() > 100 ? aiReply.substring(0, 100) + "..." : aiReply));

            // AI応答をデータベースに保存（role: assistant）
            System.out.println("💾 AI応答をデータベースに保存中...");
            AiChatMessageResponseDto savedAiMessage = aiChatMessageService.addMessage(sessionId, userId, Role.assistant, aiReply);
            System.out.println("✅ AI応答保存成功");
            System.out.println("   - messageId: " + savedAiMessage.getId());
            System.out.println("   - role: " + savedAiMessage.getRole());

            // WebSocket トピックへAI応答を送信
            System.out.println("📤 WebSocket トピック /topic/ai-chat/session/" + sessionId + " へAI応答を送信中...");
            messagingTemplate.convertAndSend(
                    "/topic/ai-chat/session/" + sessionId,
                    savedAiMessage
            );
            System.out.println("✅ AI応答 WebSocket 送信完了");
            System.out.println("========== /ai-chat/send 処理完了 ==========\n");

        } catch (NumberFormatException e) {
            System.out.println("❌ 型変換エラー発生");
            System.out.println("   エラーメッセージ: " + e.getMessage());
            e.printStackTrace();
            System.out.println("========== /ai-chat/send 処理失敗 ==========\n");
        } catch (NullPointerException e) {
            System.out.println("❌ NullPointerException 発生");
            System.out.println("   ペイロードに必須パラメータが不足しています");
            System.out.println("   必須: userId, content");
            System.out.println("   エラーメッセージ: " + e.getMessage());
            e.printStackTrace();
            System.out.println("========== /ai-chat/send 処理失敗 ==========\n");
        } catch (Exception e) {
            System.out.println("❌ 予期しないエラー発生");
            System.out.println("   エラータイプ: " + e.getClass().getName());
            System.out.println("   エラーメッセージ: " + e.getMessage());
            e.printStackTrace();
            System.out.println("========== /ai-chat/send 処理失敗 ==========\n");
        }
    }

    /**
     * AIからのレスポンスを保存してブロードキャスト
     * Lambda等の外部サービスから呼び出される想定
     */
    @MessageMapping("/ai-chat/response")
    public void receiveAiResponse(@Payload Map<String, Object> payload) {
        System.out.println("\n========== WebSocket /ai-chat/response リクエスト受信 ==========");
        System.out.println("🤖 AIレスポンス ペイロード: " + payload);

        try {
            Integer sessionId = convertToInteger(payload.get("sessionId"));
            Integer userId = convertToInteger(payload.get("userId"));
            String content = (String) payload.get("content");

            // AIからのレスポンスを保存
            AiChatMessageResponseDto saved = aiChatMessageService.addAssistantMessage(sessionId, userId, content);

            // WebSocket トピックへ送信
            messagingTemplate.convertAndSend(
                    "/topic/ai-chat/session/" + sessionId,
                    saved
            );
            System.out.println("✅ AIレスポンス送信完了");
            System.out.println("========== /ai-chat/response 処理完了 ==========\n");

        } catch (Exception e) {
            System.out.println("❌ AIレスポンス処理エラー: " + e.getMessage());
            e.printStackTrace();
            System.out.println("========== /ai-chat/response 処理失敗 ==========\n");
        }
    }

    /**
     * セッション削除
     */
    @MessageMapping("/ai-chat/delete-session")
    public void deleteSession(@Payload Map<String, Object> payload) {
        System.out.println("\n========== WebSocket /ai-chat/delete-session リクエスト受信 ==========");

        try {
            Integer sessionId = convertToInteger(payload.get("sessionId"));
            Integer userId = convertToInteger(payload.get("userId"));

            aiChatSessionService.deleteSession(sessionId, userId);

            // 削除完了通知
            messagingTemplate.convertAndSend(
                    "/topic/ai-chat/user/" + userId + "/session-deleted",
                    Map.of("sessionId", sessionId, "deleted", true)
            );
            System.out.println("✅ セッション削除完了");
            System.out.println("========== /ai-chat/delete-session 処理完了 ==========\n");

        } catch (Exception e) {
            System.out.println("❌ セッション削除エラー: " + e.getMessage());
            e.printStackTrace();
            System.out.println("========== /ai-chat/delete-session 処理失敗 ==========\n");
        }
    }

    /**
     * Object を Integer に変換するユーティリティメソッド
     */
    private Integer convertToInteger(Object obj) {
        if (obj instanceof Integer) {
            return (Integer) obj;
        } else if (obj instanceof Number) {
            return ((Number) obj).intValue();
        } else if (obj instanceof String) {
            return Integer.parseInt((String) obj);
        }
        throw new IllegalArgumentException("Cannot convert to Integer: " + obj);
    }
}
