package com.example.FreStyle.controller;

import java.util.Map;

import org.springframework.messaging.handler.annotation.MessageMapping;
import org.springframework.messaging.handler.annotation.Payload;
import org.springframework.messaging.simp.SimpMessageHeaderAccessor;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.stereotype.Controller;

import com.example.FreStyle.config.WebSocketAuthHandshakeInterceptor;
import com.example.FreStyle.constant.SceneDisplayName;
import com.example.FreStyle.dto.AiChatMessageResponseDto;
import com.example.FreStyle.dto.AiChatSessionDto;
import com.example.FreStyle.dto.ScoreCardDto;
import com.example.FreStyle.service.BedrockService;
import com.example.FreStyle.usecase.AddAiChatMessageUseCase;
import com.example.FreStyle.usecase.CreateAiChatSessionUseCase;
import com.example.FreStyle.usecase.DeleteAiChatSessionUseCase;
import com.example.FreStyle.usecase.GetAiReplyUseCase;
import com.example.FreStyle.usecase.SaveScoreCardUseCase;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@Controller
@RequiredArgsConstructor
@Slf4j
public class AiChatWebSocketController {

    private final BedrockService bedrockService;
    private final SimpMessagingTemplate messagingTemplate;

    // UseCases (クリーンアーキテクチャー)
    private final CreateAiChatSessionUseCase createAiChatSessionUseCase;
    private final AddAiChatMessageUseCase addAiChatMessageUseCase;
    private final DeleteAiChatSessionUseCase deleteAiChatSessionUseCase;
    private final GetAiReplyUseCase getAiReplyUseCase;
    private final SaveScoreCardUseCase saveScoreCardUseCase;

    /**
     * AIチャットメッセージ送信
     * クライアントから /app/ai-chat/send へメッセージを送信
     */
    @MessageMapping("/ai-chat/send")
    public void sendMessage(@Payload Map<String, Object> payload, SimpMessageHeaderAccessor headerAccessor) {
        log.info("\n========== WebSocket /ai-chat/send リクエスト受信 ==========");

        try {
            // 認証済みユーザーIDをセッション属性から取得（クライアントのpayloadは無視）
            Integer userId = getAuthenticatedUserId(headerAccessor);
            if (userId == null) {
                log.warn("WebSocket認証エラー: 認証されていないユーザーからのAIチャット送信");
                return;
            }

            Object sessionIdObj = payload.get("sessionId");
            Object contentObj = payload.get("content");
            Object roleObj = payload.get("role");
            Object fromChatFeedbackObj = payload.get("fromChatFeedback");
            Object sceneObj = payload.get("scene");
            Object sessionTypeObj = payload.get("sessionType");
            Object scenarioIdObj = payload.get("scenarioId");

            // sessionId の変換（新規セッションの場合はnull）
            Integer sessionId = sessionIdObj != null ? convertToInteger(sessionIdObj) : null;

            String content = (String) contentObj;
            String role = roleObj != null ? (String) roleObj : "user";

            // チャットフィードバックモードの判定
            boolean fromChatFeedback = fromChatFeedbackObj != null &&
                (fromChatFeedbackObj instanceof Boolean ? (Boolean) fromChatFeedbackObj :
                 "true".equalsIgnoreCase(String.valueOf(fromChatFeedbackObj)));

            // シーンの取得
            String scene = sceneObj != null ? String.valueOf(sceneObj) : null;

            // セッション種別・シナリオIDの取得
            String sessionType = sessionTypeObj != null ? String.valueOf(sessionTypeObj) : "normal";
            Integer scenarioId = scenarioIdObj != null ? convertToInteger(scenarioIdObj) : null;
            boolean isPracticeMode = "practice".equals(sessionType);

            log.info("✅ パラメータ抽出成功 - userId: {}, sessionId: {}", userId, sessionId);

            // セッションが存在しない場合は新規作成
            if (sessionId == null) {
                String title = fromChatFeedback ? "チャットフィードバック" : "新しいチャット";
                if (scene != null && fromChatFeedback) {
                    title = SceneDisplayName.of(scene) + "フィードバック";
                }
                AiChatSessionDto newSession = createAiChatSessionUseCase.execute(userId, title, null, scene);
                sessionId = newSession.id();

                messagingTemplate.convertAndSend(
                        "/topic/ai-chat/user/" + userId + "/session",
                        newSession
                );
            }

            // メッセージ保存（ユーザーメッセージ）
            AiChatMessageResponseDto savedUserMessage = addAiChatMessageUseCase.execute(sessionId, userId, role, content);

            // WebSocket トピックへユーザーメッセージを送信
            messagingTemplate.convertAndSend(
                    "/topic/ai-chat/session/" + sessionId,
                    savedUserMessage
            );

            // Bedrockにメッセージを送信してAI応答を取得
            var aiReplyCommand = new GetAiReplyUseCase.Command(content, isPracticeMode, scenarioId, fromChatFeedback, scene, userId);
            String aiReply = getAiReplyUseCase.execute(aiReplyCommand);

            // AI応答をデータベースに保存（role: assistant）
            AiChatMessageResponseDto savedAiMessage = addAiChatMessageUseCase.execute(sessionId, userId, "assistant", aiReply);

            // WebSocket トピックへAI応答を送信
            messagingTemplate.convertAndSend(
                    "/topic/ai-chat/session/" + sessionId,
                    savedAiMessage
            );

            // スコア抽出・保存・通知
            notifyScoreCardIfNeeded(sessionId, userId, aiReply, scene, fromChatFeedback, isPracticeMode);

            log.info("========== /ai-chat/send 処理完了 ==========\n");

        } catch (NumberFormatException e) {
            log.error("AIチャット送信エラー(型変換): {}", e.getMessage(), e);
        } catch (NullPointerException e) {
            log.error("AIチャット送信エラー(パラメータ不足): {}", e.getMessage(), e);
        } catch (Exception e) {
            log.error("AIチャット送信エラー: {}", e.getMessage(), e);
        }
    }

    /**
     * AIからのレスポンスを保存してブロードキャスト
     */
    @MessageMapping("/ai-chat/response")
    public void receiveAiResponse(@Payload Map<String, Object> payload, SimpMessageHeaderAccessor headerAccessor) {
        log.info("\n========== WebSocket /ai-chat/response リクエスト受信 ==========");

        try {
            Integer userId = getAuthenticatedUserId(headerAccessor);
            if (userId == null) {
                log.warn("WebSocket認証エラー: 認証されていないユーザーからのAIレスポンス");
                return;
            }

            Integer sessionId = convertToInteger(payload.get("sessionId"));
            String content = (String) payload.get("content");

            AiChatMessageResponseDto saved = addAiChatMessageUseCase.executeAssistantMessage(sessionId, userId, content);

            messagingTemplate.convertAndSend(
                    "/topic/ai-chat/session/" + sessionId,
                    saved
            );
            log.info("✅ AIレスポンス送信完了");

        } catch (Exception e) {
            log.error("AIレスポンス処理エラー: {}", e.getMessage(), e);
        }
    }

    /**
     * メッセージの言い換え提案
     */
    @MessageMapping("/ai-chat/rephrase")
    public void rephraseMessage(@Payload Map<String, Object> payload, SimpMessageHeaderAccessor headerAccessor) {
        log.info("\n========== WebSocket /ai-chat/rephrase リクエスト受信 ==========");

        try {
            Integer userId = getAuthenticatedUserId(headerAccessor);
            if (userId == null) {
                log.warn("WebSocket認証エラー: 認証されていないユーザーからの言い換えリクエスト");
                return;
            }

            String originalMessage = (String) payload.get("originalMessage");
            Object sceneObj = payload.get("scene");
            String scene = sceneObj != null ? String.valueOf(sceneObj) : null;

            String rephraseResult = bedrockService.rephrase(originalMessage, scene);

            messagingTemplate.convertAndSend(
                    "/topic/ai-chat/user/" + userId + "/rephrase",
                    Map.of(
                            "originalMessage", originalMessage,
                            "result", rephraseResult
                    )
            );
            log.info("✅ 言い換え結果送信完了");

        } catch (Exception e) {
            log.error("言い換え処理エラー: {}", e.getMessage(), e);
        }
    }

    /**
     * セッション削除
     */
    @MessageMapping("/ai-chat/delete-session")
    public void deleteSession(@Payload Map<String, Object> payload, SimpMessageHeaderAccessor headerAccessor) {
        log.info("\n========== WebSocket /ai-chat/delete-session リクエスト受信 ==========");

        try {
            Integer userId = getAuthenticatedUserId(headerAccessor);
            if (userId == null) {
                log.warn("WebSocket認証エラー: 認証されていないユーザーからのセッション削除");
                return;
            }

            Integer sessionId = convertToInteger(payload.get("sessionId"));

            deleteAiChatSessionUseCase.execute(sessionId, userId);

            messagingTemplate.convertAndSend(
                    "/topic/ai-chat/user/" + userId + "/session-deleted",
                    Map.of("sessionId", sessionId, "deleted", true)
            );
            log.info("✅ セッション削除完了");

        } catch (Exception e) {
            log.error("セッション削除エラー: {}", e.getMessage(), e);
        }
    }

    /**
     * スコアカードを抽出・保存し、WebSocketで通知する
     */
    private void notifyScoreCardIfNeeded(Integer sessionId, Integer userId, String aiReply,
                                         String scene, boolean fromChatFeedback, boolean isPracticeMode) {
        if (fromChatFeedback) {
            sendScoreCard(sessionId, userId, aiReply, scene, "スコアカード");
        }
        if (isPracticeMode && aiReply.contains("練習終了")) {
            log.info("🎓 練習終了を検知 - スコア抽出中...");
            sendScoreCard(sessionId, userId, aiReply, null, "練習スコアカード");
        }
    }

    private void sendScoreCard(Integer sessionId, Integer userId, String aiReply,
                               String scene, String logLabel) {
        ScoreCardDto scoreCard = saveScoreCardUseCase.execute(sessionId, userId, aiReply, scene);
        if (scoreCard != null) {
            messagingTemplate.convertAndSend(
                    "/topic/ai-chat/user/" + userId + "/scorecard",
                    scoreCard
            );
            log.info("✅ {}送信完了 - 総合スコア: {}", logLabel, scoreCard.overallScore());
        } else {
            log.warn("⚠️ AI応答からスコアを抽出できませんでした");
        }
    }

    private Integer getAuthenticatedUserId(SimpMessageHeaderAccessor headerAccessor) {
        Map<String, Object> sessionAttributes = headerAccessor.getSessionAttributes();
        if (sessionAttributes == null) return null;
        return (Integer) sessionAttributes.get(WebSocketAuthHandshakeInterceptor.AUTHENTICATED_USER_ID);
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
