package com.example.FreStyle.controller;

import java.util.Map;

import org.springframework.messaging.handler.annotation.MessageMapping;
import org.springframework.messaging.handler.annotation.Payload;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.stereotype.Controller;

import com.example.FreStyle.constant.SceneDisplayName;
import com.example.FreStyle.dto.AiChatMessageResponseDto;
import com.example.FreStyle.dto.AiChatSessionDto;
import com.example.FreStyle.dto.ScoreCardDto;
import com.example.FreStyle.entity.AiChatMessage.Role;
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
    public void sendMessage(@Payload Map<String, Object> payload) {
        log.info("\n========== WebSocket /ai-chat/send リクエスト受信 ==========");
        log.info("📨 ペイロード全体: {}", payload);

        try {
            // パラメータの取得と検証
            log.info("🔍 パラメータを抽出中...");
            Object userIdObj = payload.get("userId");
            Object sessionIdObj = payload.get("sessionId");
            Object contentObj = payload.get("content");
            Object roleObj = payload.get("role"); // "user" または "assistant"
            Object fromChatFeedbackObj = payload.get("fromChatFeedback"); // チャットフィードバックモードフラグ
            Object sceneObj = payload.get("scene"); // フィードバックシーン
            Object sessionTypeObj = payload.get("sessionType"); // セッション種別（normal, practice）
            Object scenarioIdObj = payload.get("scenarioId"); // 練習シナリオID

            log.debug("   - userId タイプ: {}", userIdObj != null ? userIdObj.getClass().getSimpleName() : "null");
            log.debug("   - userId 値: {}", userIdObj);
            log.debug("   - sessionId タイプ: {}", sessionIdObj != null ? sessionIdObj.getClass().getSimpleName() : "null");
            log.debug("   - sessionId 値: {}", sessionIdObj);
            log.debug("   - content: {}", contentObj);
            log.debug("   - role: {}", roleObj);
            log.debug("   - fromChatFeedback: {}", fromChatFeedbackObj);
            log.debug("   - scene: {}", sceneObj);

            // userId の変換
            Integer userId = convertToInteger(userIdObj);

            // sessionId の変換（新規セッションの場合はnull）
            Integer sessionId = sessionIdObj != null ? convertToInteger(sessionIdObj) : null;

            String content = (String) contentObj;
            String roleStr = roleObj != null ? (String) roleObj : "user";
            Role role = "assistant".equalsIgnoreCase(roleStr) ? Role.assistant : Role.user;

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

            log.info("✅ パラメータ抽出成功");
            log.debug("   - userId (最終): {}", userId);
            log.debug("   - sessionId (最終): {}", sessionId);
            log.debug("   - content: {}", content);
            log.debug("   - role: {}", role);
            log.debug("   - fromChatFeedback (最終): {}", fromChatFeedback);
            log.debug("   - scene (最終): {}", scene);

            // セッションが存在しない場合は新規作成
            if (sessionId == null) {
                log.info("🆕 新規セッション作成中...");
                // フィードバックモードの場合はタイトルを変更
                String title = fromChatFeedback ? "チャットフィードバック" : "新しいチャット";
                // シーンが指定されている場合はタイトルにシーン名を含める
                if (scene != null && fromChatFeedback) {
                    title = SceneDisplayName.of(scene) + "フィードバック";
                }
                AiChatSessionDto newSession = createAiChatSessionUseCase.execute(userId, title, null, scene);
                sessionId = newSession.getId();
                log.info("✅ 新規セッション作成完了 - sessionId: {}", sessionId);

                // 新しいセッション情報をクライアントに通知
                messagingTemplate.convertAndSend(
                        "/topic/ai-chat/user/" + userId + "/session",
                        newSession
                );
            }

            // メッセージ保存（ユーザーメッセージ）
            log.info("💾 ユーザーメッセージをデータベースに保存中...");
            AiChatMessageResponseDto savedUserMessage = addAiChatMessageUseCase.execute(sessionId, userId, role, content);
            log.info("✅ ユーザーメッセージ保存成功");
            log.debug("   - messageId: {}", savedUserMessage.getId());
            log.debug("   - sessionId: {}", savedUserMessage.getSessionId());
            log.debug("   - role: {}", savedUserMessage.getRole());

            // WebSocket トピックへユーザーメッセージを送信
            log.info("📤 WebSocket トピック /topic/ai-chat/session/{} へユーザーメッセージを送信中...", sessionId);
            messagingTemplate.convertAndSend(
                    "/topic/ai-chat/session/" + sessionId,
                    savedUserMessage
            );
            log.info("✅ ユーザーメッセージ WebSocket 送信完了");

            // Bedrockにメッセージを送信してAI応答を取得
            String aiReply = getAiReplyUseCase.execute(content, isPracticeMode, scenarioId, fromChatFeedback, scene, userId);

            // AI応答をデータベースに保存（role: assistant）
            log.info("💾 AI応答をデータベースに保存中...");
            AiChatMessageResponseDto savedAiMessage = addAiChatMessageUseCase.execute(sessionId, userId, Role.assistant, aiReply);
            log.info("✅ AI応答保存成功");
            log.debug("   - messageId: {}", savedAiMessage.getId());
            log.debug("   - role: {}", savedAiMessage.getRole());

            // WebSocket トピックへAI応答を送信
            log.info("📤 WebSocket トピック /topic/ai-chat/session/{} へAI応答を送信中...", sessionId);
            messagingTemplate.convertAndSend(
                    "/topic/ai-chat/session/" + sessionId,
                    savedAiMessage
            );
            log.info("✅ AI応答 WebSocket 送信完了");

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
    public void receiveAiResponse(@Payload Map<String, Object> payload) {
        log.info("\n========== WebSocket /ai-chat/response リクエスト受信 ==========");
        log.info("🤖 AIレスポンス ペイロード: {}", payload);

        try {
            Integer sessionId = convertToInteger(payload.get("sessionId"));
            Integer userId = convertToInteger(payload.get("userId"));
            String content = (String) payload.get("content");

            // AIからのレスポンスを保存
            AiChatMessageResponseDto saved = addAiChatMessageUseCase.executeAssistantMessage(sessionId, userId, content);

            // WebSocket トピックへ送信
            messagingTemplate.convertAndSend(
                    "/topic/ai-chat/session/" + sessionId,
                    saved
            );
            log.info("✅ AIレスポンス送信完了");
            log.info("========== /ai-chat/response 処理完了 ==========\n");

        } catch (Exception e) {
            log.error("AIレスポンス処理エラー: {}", e.getMessage(), e);
        }
    }

    /**
     * メッセージの言い換え提案
     * クライアントから /app/ai-chat/rephrase へリクエストを送信
     */
    @MessageMapping("/ai-chat/rephrase")
    public void rephraseMessage(@Payload Map<String, Object> payload) {
        log.info("\n========== WebSocket /ai-chat/rephrase リクエスト受信 ==========");

        try {
            Integer userId = convertToInteger(payload.get("userId"));
            String originalMessage = (String) payload.get("originalMessage");
            Object sceneObj = payload.get("scene");
            String scene = sceneObj != null ? String.valueOf(sceneObj) : null;

            log.debug("   - userId: {}", userId);
            log.debug("   - originalMessage: {}", originalMessage);
            log.debug("   - scene: {}", scene);

            // Bedrockに言い換えリクエスト
            String rephraseResult = bedrockService.rephrase(originalMessage, scene);
            log.info("✅ 言い換え結果取得: {}", rephraseResult);

            // WebSocket トピックへ言い換え結果を送信
            messagingTemplate.convertAndSend(
                    "/topic/ai-chat/user/" + userId + "/rephrase",
                    Map.of(
                            "originalMessage", originalMessage,
                            "result", rephraseResult
                    )
            );
            log.info("✅ 言い換え結果送信完了");
            log.info("========== /ai-chat/rephrase 処理完了 ==========\n");

        } catch (Exception e) {
            log.error("言い換え処理エラー: {}", e.getMessage(), e);
        }
    }

    /**
     * セッション削除
     */
    @MessageMapping("/ai-chat/delete-session")
    public void deleteSession(@Payload Map<String, Object> payload) {
        log.info("\n========== WebSocket /ai-chat/delete-session リクエスト受信 ==========");

        try {
            Integer sessionId = convertToInteger(payload.get("sessionId"));
            Integer userId = convertToInteger(payload.get("userId"));

            deleteAiChatSessionUseCase.execute(sessionId, userId);

            // 削除完了通知
            messagingTemplate.convertAndSend(
                    "/topic/ai-chat/user/" + userId + "/session-deleted",
                    Map.of("sessionId", sessionId, "deleted", true)
            );
            log.info("✅ セッション削除完了");
            log.info("========== /ai-chat/delete-session 処理完了 ==========\n");

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
            log.info("✅ {}送信完了 - 総合スコア: {}", logLabel, scoreCard.getOverallScore());
        } else {
            log.warn("⚠️ AI応答からスコアを抽出できませんでした");
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
