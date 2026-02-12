package com.example.FreStyle.controller;

import java.util.List;
import java.util.Map;

import org.springframework.messaging.handler.annotation.MessageMapping;
import org.springframework.messaging.handler.annotation.Payload;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.stereotype.Controller;

import com.example.FreStyle.dto.AiChatMessageResponseDto;
import com.example.FreStyle.dto.AiChatSessionDto;
import com.example.FreStyle.dto.ScoreCardDto;
import com.example.FreStyle.dto.UserProfileDto;
import com.example.FreStyle.entity.AiChatMessage.Role;
import com.example.FreStyle.entity.PracticeScenario;
import com.example.FreStyle.service.AiChatMessageService;
import com.example.FreStyle.service.AiChatSessionService;
import com.example.FreStyle.service.BedrockService;
import com.example.FreStyle.service.PracticeScenarioService;
import com.example.FreStyle.service.ScoreCardService;
import com.example.FreStyle.service.SystemPromptBuilder;
import com.example.FreStyle.service.UserProfileService;

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
    private final UserProfileService userProfileService;
    private final ScoreCardService scoreCardService;
    private final PracticeScenarioService practiceScenarioService;
    private final SystemPromptBuilder systemPromptBuilder;

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
            Object fromChatFeedbackObj = payload.get("fromChatFeedback"); // チャットフィードバックモードフラグ
            Object sceneObj = payload.get("scene"); // フィードバックシーン
            Object sessionTypeObj = payload.get("sessionType"); // セッション種別（normal, practice）
            Object scenarioIdObj = payload.get("scenarioId"); // 練習シナリオID

            System.out.println("   - userId タイプ: " + (userIdObj != null ? userIdObj.getClass().getSimpleName() : "null"));
            System.out.println("   - userId 値: " + userIdObj);
            System.out.println("   - sessionId タイプ: " + (sessionIdObj != null ? sessionIdObj.getClass().getSimpleName() : "null"));
            System.out.println("   - sessionId 値: " + sessionIdObj);
            System.out.println("   - content: " + contentObj);
            System.out.println("   - role: " + roleObj);
            System.out.println("   - fromChatFeedback: " + fromChatFeedbackObj);
            System.out.println("   - scene: " + sceneObj);

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

            System.out.println("✅ パラメータ抽出成功");
            System.out.println("   - userId (最終): " + userId);
            System.out.println("   - sessionId (最終): " + sessionId);
            System.out.println("   - content: " + content);
            System.out.println("   - role: " + role);
            System.out.println("   - fromChatFeedback (最終): " + fromChatFeedback);
            System.out.println("   - scene (最終): " + scene);

            // セッションが存在しない場合は新規作成
            if (sessionId == null) {
                System.out.println("🆕 新規セッション作成中...");
                // フィードバックモードの場合はタイトルを変更
                String title = fromChatFeedback ? "チャットフィードバック" : "新しいチャット";
                // シーンが指定されている場合はタイトルにシーン名を含める
                if (scene != null && fromChatFeedback) {
                    title = getSceneDisplayName(scene) + "フィードバック";
                }
                AiChatSessionDto newSession = aiChatSessionService.createSession(userId, title, null, scene);
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
            String aiReply;
            if (isPracticeMode && scenarioId != null) {
                // 練習モード: シナリオに基づいたロールプレイ
                System.out.println("🎭 練習モード: scenarioId=" + scenarioId);
                PracticeScenario scenario = practiceScenarioService.getScenarioEntityById(scenarioId);
                String practicePrompt = systemPromptBuilder.buildPracticePrompt(
                        scenario.getName(), scenario.getRoleName(),
                        scenario.getDifficulty(), scenario.getSystemPrompt());
                aiReply = bedrockService.chatInPracticeMode(content, practicePrompt);
            } else if (fromChatFeedback) {
                // チャットフィードバックモード: バックエンドでUserProfileを取得
                System.out.println("🤖 フィードバックモード: UserProfileをバックエンドで取得中... scene=" + scene);
                UserProfileDto userProfile = userProfileService.getProfileByUserId(userId);

                if (userProfile != null) {
                    System.out.println("✅ UserProfile取得成功");
                    System.out.println("   - UserProfile情報:");
                    System.out.println("     - displayName: " + userProfile.getDisplayName());
                    System.out.println("     - goals: " + userProfile.getGoals());
                    System.out.println("     - concerns: " + userProfile.getConcerns());
                    System.out.println("     - preferredFeedbackStyle: " + userProfile.getPreferredFeedbackStyle());

                    String personalityTraits = userProfile.getPersonalityTraits() != null
                        ? String.join(", ", userProfile.getPersonalityTraits())
                        : null;

                    aiReply = bedrockService.chatWithUserProfileAndScene(
                        content,
                        scene,
                        userProfile.getDisplayName(),
                        userProfile.getSelfIntroduction(),
                        userProfile.getCommunicationStyle(),
                        personalityTraits,
                        userProfile.getGoals(),
                        userProfile.getConcerns(),
                        userProfile.getPreferredFeedbackStyle()
                    );
                } else {
                    // UserProfileが存在しない場合は通常モードで処理
                    System.out.println("⚠️ UserProfileが見つかりません。通常モードで処理します。");
                    aiReply = bedrockService.chat(content);
                }
            } else {
                // 通常モード
                System.out.println("🤖 Bedrock にメッセージを送信中...");
                aiReply = bedrockService.chat(content);
            }
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

            // フィードバックモードの場合、AI応答からスコアを抽出・保存・通知
            if (fromChatFeedback) {
                List<ScoreCardService.AxisScore> scores = scoreCardService.parseScoresFromResponse(aiReply);
                if (!scores.isEmpty()) {
                    scoreCardService.saveScores(sessionId, userId, scores, scene);
                    double overallScore = scoreCardService.calculateOverallScore(scores);

                    List<ScoreCardDto.AxisScoreDto> scoreDtos = scores.stream()
                            .map(s -> new ScoreCardDto.AxisScoreDto(s.getAxis(), s.getScore(), s.getComment()))
                            .toList();

                    ScoreCardDto scoreCard = new ScoreCardDto(sessionId, scoreDtos, overallScore);

                    messagingTemplate.convertAndSend(
                            "/topic/ai-chat/user/" + userId + "/scorecard",
                            scoreCard
                    );
                    System.out.println("✅ スコアカード送信完了 - 総合スコア: " + overallScore);
                } else {
                    System.out.println("⚠️ AI応答からスコアを抽出できませんでした");
                }
            }

            // 練習モードで「練習終了」の場合、スコアを抽出・保存・通知
            if (isPracticeMode && aiReply.contains("練習終了")) {
                System.out.println("🎓 練習終了を検知 - スコア抽出中...");
                List<ScoreCardService.AxisScore> scores = scoreCardService.parseScoresFromResponse(aiReply);
                if (!scores.isEmpty()) {
                    scoreCardService.saveScores(sessionId, userId, scores, null);
                    double overallScore = scoreCardService.calculateOverallScore(scores);

                    List<ScoreCardDto.AxisScoreDto> scoreDtos = scores.stream()
                            .map(s -> new ScoreCardDto.AxisScoreDto(s.getAxis(), s.getScore(), s.getComment()))
                            .toList();

                    ScoreCardDto scoreCard = new ScoreCardDto(sessionId, scoreDtos, overallScore);

                    messagingTemplate.convertAndSend(
                            "/topic/ai-chat/user/" + userId + "/scorecard",
                            scoreCard
                    );
                    System.out.println("✅ 練習スコアカード送信完了 - 総合スコア: " + overallScore);
                } else {
                    System.out.println("⚠️ 練習AI応答からスコアを抽出できませんでした");
                }
            }

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
     * メッセージの言い換え提案
     * クライアントから /app/ai-chat/rephrase へリクエストを送信
     */
    @MessageMapping("/ai-chat/rephrase")
    public void rephraseMessage(@Payload Map<String, Object> payload) {
        System.out.println("\n========== WebSocket /ai-chat/rephrase リクエスト受信 ==========");

        try {
            Integer userId = convertToInteger(payload.get("userId"));
            String originalMessage = (String) payload.get("originalMessage");
            Object sceneObj = payload.get("scene");
            String scene = sceneObj != null ? String.valueOf(sceneObj) : null;

            System.out.println("   - userId: " + userId);
            System.out.println("   - originalMessage: " + originalMessage);
            System.out.println("   - scene: " + scene);

            // Bedrockに言い換えリクエスト
            String rephraseResult = bedrockService.rephrase(originalMessage, scene);
            System.out.println("✅ 言い換え結果取得: " + rephraseResult);

            // WebSocket トピックへ言い換え結果を送信
            messagingTemplate.convertAndSend(
                    "/topic/ai-chat/user/" + userId + "/rephrase",
                    Map.of(
                            "originalMessage", originalMessage,
                            "result", rephraseResult
                    )
            );
            System.out.println("✅ 言い換え結果送信完了");
            System.out.println("========== /ai-chat/rephrase 処理完了 ==========\n");

        } catch (Exception e) {
            System.out.println("❌ 言い換え処理エラー: " + e.getMessage());
            e.printStackTrace();
            System.out.println("========== /ai-chat/rephrase 処理失敗 ==========\n");
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
     * シーン識別子から表示名を取得
     */
    private String getSceneDisplayName(String scene) {
        if (scene == null) return "";
        switch (scene) {
            case "meeting": return "会議";
            case "one_on_one": return "1on1";
            case "email": return "メール";
            case "presentation": return "プレゼン";
            case "negotiation": return "商談";
            case "code_review": return "コードレビュー";
            case "incident": return "障害対応";
            case "daily_report": return "日報・週報";
            default: return "";
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
