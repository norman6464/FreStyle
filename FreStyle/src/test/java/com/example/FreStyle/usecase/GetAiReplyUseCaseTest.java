package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.PracticeScenarioDto;
import com.example.FreStyle.dto.UserProfileDto;
import com.example.FreStyle.service.BedrockService;
import com.example.FreStyle.service.SystemPromptBuilder;
import com.example.FreStyle.service.UserProfileService;

import java.util.List;

@ExtendWith(MockitoExtension.class)
class GetAiReplyUseCaseTest {

    @Mock
    private BedrockService bedrockService;

    @Mock
    private UserProfileService userProfileService;

    @Mock
    private SystemPromptBuilder systemPromptBuilder;

    @Mock
    private GetPracticeScenarioByIdUseCase getPracticeScenarioByIdUseCase;

    @InjectMocks
    private GetAiReplyUseCase getAiReplyUseCase;

    @Nested
    class 通常モード {

        @Test
        void 通常モードでBedrockのchatを呼ぶ() {
            when(bedrockService.chat("こんにちは")).thenReturn("AI応答です");

            String result = getAiReplyUseCase.execute("こんにちは", false, null, false, null, 1);

            assertThat(result).isEqualTo("AI応答です");
            verify(bedrockService).chat("こんにちは");
        }
    }

    @Nested
    class 練習モード {

        @Test
        void 練習モードでシナリオベースのプロンプトを構築してchatInPracticeModeを呼ぶ() {
            PracticeScenarioDto scenario = new PracticeScenarioDto(
                    1, "クレーム対応", "説明", "customer", "顧客", "intermediate", "システムプロンプト");
            when(getPracticeScenarioByIdUseCase.execute(1)).thenReturn(scenario);
            when(systemPromptBuilder.buildPracticePrompt("クレーム対応", "顧客", "intermediate", "システムプロンプト"))
                    .thenReturn("構築されたプロンプト");
            when(bedrockService.chatInPracticeMode("ユーザーメッセージ", "構築されたプロンプト"))
                    .thenReturn("練習AI応答");

            String result = getAiReplyUseCase.execute("ユーザーメッセージ", true, 1, false, null, 1);

            assertThat(result).isEqualTo("練習AI応答");
        }

        @Test
        void 練習開始時は開始プロンプトを付与する() {
            PracticeScenarioDto scenario = new PracticeScenarioDto(
                    1, "クレーム対応", "説明", "customer", "顧客", "intermediate", "システムプロンプト");
            when(getPracticeScenarioByIdUseCase.execute(1)).thenReturn(scenario);
            when(systemPromptBuilder.buildPracticePrompt("クレーム対応", "顧客", "intermediate", "システムプロンプト"))
                    .thenReturn("構築されたプロンプト");
            when(bedrockService.chatInPracticeMode(eq(""), any(String.class)))
                    .thenReturn("練習開始AI応答");

            String result = getAiReplyUseCase.execute("練習開始", true, 1, false, null, 1);

            assertThat(result).isEqualTo("練習開始AI応答");
            verify(bedrockService).chatInPracticeMode(eq(""), any(String.class));
        }
    }

    @Nested
    class フィードバックモード {

        @Test
        void UserProfileありでchatWithUserProfileAndSceneを呼ぶ() {
            UserProfileDto profile = new UserProfileDto(
                    1, 1, "テストユーザー", "自己紹介",
                    "フレンドリー", List.of("明るい", "積極的"),
                    "目標", "課題", "優しく");
            when(userProfileService.getProfileByUserId(1)).thenReturn(profile);
            when(bedrockService.chatWithUserProfileAndScene(
                    "メッセージ", "meeting", "テストユーザー", "自己紹介",
                    "フレンドリー", "明るい, 積極的", "目標", "課題", "優しく"))
                    .thenReturn("フィードバック応答");

            String result = getAiReplyUseCase.execute("メッセージ", false, null, true, "meeting", 1);

            assertThat(result).isEqualTo("フィードバック応答");
        }

        @Test
        void UserProfileがnullの場合は通常モードにフォールバック() {
            when(userProfileService.getProfileByUserId(1)).thenReturn(null);
            when(bedrockService.chat("メッセージ")).thenReturn("通常応答");

            String result = getAiReplyUseCase.execute("メッセージ", false, null, true, "meeting", 1);

            assertThat(result).isEqualTo("通常応答");
            verify(bedrockService).chat("メッセージ");
        }

        @Test
        void personalityTraitsがnullの場合はnullを渡す() {
            UserProfileDto profile = new UserProfileDto(
                    1, 1, "テストユーザー", "自己紹介",
                    "フレンドリー", null,
                    "目標", "課題", "優しく");
            when(userProfileService.getProfileByUserId(1)).thenReturn(profile);
            when(bedrockService.chatWithUserProfileAndScene(
                    "メッセージ", "meeting", "テストユーザー", "自己紹介",
                    "フレンドリー", null, "目標", "課題", "優しく"))
                    .thenReturn("フィードバック応答");

            String result = getAiReplyUseCase.execute("メッセージ", false, null, true, "meeting", 1);

            assertThat(result).isEqualTo("フィードバック応答");
        }
    }
}
