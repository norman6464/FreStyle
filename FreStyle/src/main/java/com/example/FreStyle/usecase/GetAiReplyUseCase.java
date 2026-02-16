package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;

import com.example.FreStyle.dto.PracticeScenarioDto;
import com.example.FreStyle.dto.UserProfileDto;
import com.example.FreStyle.service.BedrockService;
import com.example.FreStyle.service.SystemPromptBuilder;
import com.example.FreStyle.service.UserProfileService;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@Service
@RequiredArgsConstructor
@Slf4j
public class GetAiReplyUseCase {

    private final BedrockService bedrockService;
    private final UserProfileService userProfileService;
    private final SystemPromptBuilder systemPromptBuilder;
    private final GetPracticeScenarioByIdUseCase getPracticeScenarioByIdUseCase;

    /**
     * AI応答を取得（モードに応じて適切なBedrockメソッドを呼び出す）
     */
    public String execute(String content, boolean isPracticeMode, Integer scenarioId,
                          boolean fromChatFeedback, String scene, Integer userId) {
        if (isPracticeMode && scenarioId != null) {
            return handlePracticeMode(content, scenarioId);
        }

        if (fromChatFeedback) {
            return handleFeedbackMode(content, scene, userId);
        }

        log.info("🤖 Bedrock にメッセージを送信中...");
        return bedrockService.chat(content);
    }

    private String handlePracticeMode(String content, Integer scenarioId) {
        log.info("🎭 練習モード: scenarioId={}", scenarioId);
        PracticeScenarioDto scenario = getPracticeScenarioByIdUseCase.execute(scenarioId);
        String practicePrompt = systemPromptBuilder.buildPracticePrompt(
                scenario.getName(), scenario.getRoleName(),
                scenario.getDifficulty(), scenario.getSystemPrompt());

        if ("練習開始".equals(content)) {
            String startPrompt = practicePrompt +
                "\n\nこれから練習が始まります。あなたは相手役として、シナリオに基づいた最初の発言をしてください。" +
                "ユーザーに対して、シナリオの状況を反映した自然な会話で話しかけてください。";
            return bedrockService.chatInPracticeMode("", startPrompt);
        }
        return bedrockService.chatInPracticeMode(content, practicePrompt);
    }

    private String handleFeedbackMode(String content, String scene, Integer userId) {
        log.info("🤖 フィードバックモード: UserProfileをバックエンドで取得中... scene={}", scene);
        UserProfileDto userProfile = userProfileService.getProfileByUserId(userId);

        if (userProfile != null) {
            log.info("✅ UserProfile取得成功");
            String personalityTraits = userProfile.getPersonalityTraits() != null
                ? String.join(", ", userProfile.getPersonalityTraits())
                : null;

            return bedrockService.chatWithUserProfileAndScene(
                content, scene,
                userProfile.getDisplayName(),
                userProfile.getSelfIntroduction(),
                userProfile.getCommunicationStyle(),
                personalityTraits,
                userProfile.getGoals(),
                userProfile.getConcerns(),
                userProfile.getPreferredFeedbackStyle()
            );
        }

        log.warn("⚠️ UserProfileが見つかりません。通常モードで処理します。");
        return bedrockService.chat(content);
    }
}
