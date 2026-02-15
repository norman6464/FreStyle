package com.example.FreStyle.controller;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.lang.reflect.Method;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.messaging.simp.SimpMessagingTemplate;

import com.example.FreStyle.dto.AiChatMessageResponseDto;
import com.example.FreStyle.service.BedrockService;
import com.example.FreStyle.service.SystemPromptBuilder;
import com.example.FreStyle.service.UserProfileService;
import com.example.FreStyle.usecase.AddAiChatMessageUseCase;
import com.example.FreStyle.usecase.CreateAiChatSessionUseCase;
import com.example.FreStyle.usecase.DeleteAiChatSessionUseCase;
import com.example.FreStyle.usecase.GetPracticeScenarioByIdUseCase;
import com.example.FreStyle.usecase.SaveScoreCardUseCase;

@ExtendWith(MockitoExtension.class)
@DisplayName("AiChatWebSocketController")
class AiChatWebSocketControllerTest {

    @Mock private BedrockService bedrockService;
    @Mock private SimpMessagingTemplate messagingTemplate;
    @Mock private UserProfileService userProfileService;
    @Mock private SystemPromptBuilder systemPromptBuilder;
    @Mock private CreateAiChatSessionUseCase createAiChatSessionUseCase;
    @Mock private AddAiChatMessageUseCase addAiChatMessageUseCase;
    @Mock private DeleteAiChatSessionUseCase deleteAiChatSessionUseCase;
    @Mock private GetPracticeScenarioByIdUseCase getPracticeScenarioByIdUseCase;
    @Mock private SaveScoreCardUseCase saveScoreCardUseCase;

    @InjectMocks
    private AiChatWebSocketController controller;

    @Nested
    @DisplayName("receiveAiResponse")
    class ReceiveAiResponse {

        @Test
        @DisplayName("AI応答を保存してWebSocket送信する")
        void savesAndBroadcasts() {
            AiChatMessageResponseDto saved = new AiChatMessageResponseDto();
            saved.setId(1);
            saved.setSessionId(10);
            when(addAiChatMessageUseCase.executeAssistantMessage(10, 5, "AI応答")).thenReturn(saved);

            Map<String, Object> payload = Map.of(
                "sessionId", 10,
                "userId", 5,
                "content", "AI応答"
            );

            controller.receiveAiResponse(payload);

            verify(addAiChatMessageUseCase).executeAssistantMessage(10, 5, "AI応答");
            verify(messagingTemplate).convertAndSend("/topic/ai-chat/session/10", saved);
        }
    }

    @Nested
    @DisplayName("rephraseMessage")
    class RephraseMessage {

        @Test
        @DisplayName("言い換え結果をWebSocket送信する")
        void sendsRephraseResult() {
            when(bedrockService.rephrase("元のメッセージ", "meeting")).thenReturn("言い換え結果");

            Map<String, Object> payload = Map.of(
                "userId", 5,
                "originalMessage", "元のメッセージ",
                "scene", "meeting"
            );

            controller.rephraseMessage(payload);

            verify(bedrockService).rephrase("元のメッセージ", "meeting");
            verify(messagingTemplate).convertAndSend(
                eq("/topic/ai-chat/user/5/rephrase"),
                any(Map.class)
            );
        }

        @Test
        @DisplayName("シーンなしで言い換えする")
        void rephraseWithoutScene() {
            when(bedrockService.rephrase("テスト", null)).thenReturn("結果");

            Map<String, Object> payload = Map.of(
                "userId", 5,
                "originalMessage", "テスト"
            );

            controller.rephraseMessage(payload);

            verify(bedrockService).rephrase("テスト", null);
        }
    }

    @Nested
    @DisplayName("deleteSession")
    class DeleteSession {

        @Test
        @DisplayName("セッションを削除して通知する")
        void deletesAndNotifies() {
            Map<String, Object> payload = Map.of(
                "sessionId", 10,
                "userId", 5
            );

            controller.deleteSession(payload);

            verify(deleteAiChatSessionUseCase).execute(10, 5);
            verify(messagingTemplate).convertAndSend(
                eq("/topic/ai-chat/user/5/session-deleted"),
                any(Map.class)
            );
        }
    }

    @Nested
    @DisplayName("getSceneDisplayName")
    class GetSceneDisplayName {

        private String invokeGetSceneDisplayName(String scene) throws Exception {
            Method method = AiChatWebSocketController.class.getDeclaredMethod("getSceneDisplayName", String.class);
            method.setAccessible(true);
            return (String) method.invoke(controller, scene);
        }

        @Test
        @DisplayName("会議シーンの表示名を返す")
        void returnsMeetingDisplayName() throws Exception {
            assertThat(invokeGetSceneDisplayName("meeting")).isEqualTo("会議");
        }

        @Test
        @DisplayName("1on1シーンの表示名を返す")
        void returnsOneOnOneDisplayName() throws Exception {
            assertThat(invokeGetSceneDisplayName("one_on_one")).isEqualTo("1on1");
        }

        @Test
        @DisplayName("nullの場合は空文字を返す")
        void returnsEmptyForNull() throws Exception {
            assertThat(invokeGetSceneDisplayName(null)).isEmpty();
        }

        @Test
        @DisplayName("不明なシーンの場合は空文字を返す")
        void returnsEmptyForUnknown() throws Exception {
            assertThat(invokeGetSceneDisplayName("unknown")).isEmpty();
        }
    }
}
