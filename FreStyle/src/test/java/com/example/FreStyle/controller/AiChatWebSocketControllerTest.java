package com.example.FreStyle.controller;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyInt;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.HashMap;
import java.util.Map;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.messaging.simp.SimpMessageHeaderAccessor;
import org.springframework.messaging.simp.SimpMessagingTemplate;

import com.example.FreStyle.config.WebSocketAuthHandshakeInterceptor;
import com.example.FreStyle.dto.AiChatMessageResponseDto;
import com.example.FreStyle.service.BedrockService;
import com.example.FreStyle.usecase.AddAiChatMessageUseCase;
import com.example.FreStyle.usecase.CreateAiChatSessionUseCase;
import com.example.FreStyle.usecase.DeleteAiChatSessionUseCase;
import com.example.FreStyle.usecase.GetAiReplyUseCase;
import com.example.FreStyle.usecase.SaveScoreCardUseCase;

@ExtendWith(MockitoExtension.class)
@DisplayName("AiChatWebSocketController")
class AiChatWebSocketControllerTest {

    @Mock private BedrockService bedrockService;
    @Mock private SimpMessagingTemplate messagingTemplate;
    @Mock private CreateAiChatSessionUseCase createAiChatSessionUseCase;
    @Mock private AddAiChatMessageUseCase addAiChatMessageUseCase;
    @Mock private DeleteAiChatSessionUseCase deleteAiChatSessionUseCase;
    @Mock private GetAiReplyUseCase getAiReplyUseCase;
    @Mock private SaveScoreCardUseCase saveScoreCardUseCase;

    @InjectMocks
    private AiChatWebSocketController controller;

    private SimpMessageHeaderAccessor headerAccessor;

    @BeforeEach
    void setUp() {
        headerAccessor = SimpMessageHeaderAccessor.create();
        Map<String, Object> sessionAttributes = new HashMap<>();
        sessionAttributes.put(WebSocketAuthHandshakeInterceptor.AUTHENTICATED_USER_ID, 5);
        headerAccessor.setSessionAttributes(sessionAttributes);
    }

    private SimpMessageHeaderAccessor createUnauthenticatedHeader() {
        SimpMessageHeaderAccessor header = SimpMessageHeaderAccessor.create();
        header.setSessionAttributes(new HashMap<>());
        return header;
    }

    @Nested
    @DisplayName("receiveAiResponse")
    class ReceiveAiResponse {

        @Test
        @DisplayName("AI応答を保存してWebSocket送信する")
        void savesAndBroadcasts() {
            AiChatMessageResponseDto saved = new AiChatMessageResponseDto("msg-1", 10, null, null, null, null);
            when(addAiChatMessageUseCase.executeAssistantMessage(10, 5, "AI応答")).thenReturn(saved);

            Map<String, Object> payload = Map.of(
                "sessionId", 10,
                "content", "AI応答"
            );

            controller.receiveAiResponse(payload, headerAccessor);

            verify(addAiChatMessageUseCase).executeAssistantMessage(10, 5, "AI応答");
            verify(messagingTemplate).convertAndSend("/topic/ai-chat/session/10", saved);
        }

        @Test
        @DisplayName("未認証ユーザーの場合はAI応答を保存しない")
        void unauthenticated_doesNotSave() {
            Map<String, Object> payload = Map.of(
                "sessionId", 10,
                "content", "AI応答"
            );

            controller.receiveAiResponse(payload, createUnauthenticatedHeader());

            verify(addAiChatMessageUseCase, never()).executeAssistantMessage(anyInt(), anyInt(), anyString());
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
                "originalMessage", "元のメッセージ",
                "scene", "meeting"
            );

            controller.rephraseMessage(payload, headerAccessor);

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
                "originalMessage", "テスト"
            );

            controller.rephraseMessage(payload, headerAccessor);

            verify(bedrockService).rephrase("テスト", null);
        }

        @Test
        @DisplayName("未認証ユーザーの場合は言い換えしない")
        void unauthenticated_doesNotRephrase() {
            Map<String, Object> payload = Map.of(
                "originalMessage", "テスト"
            );

            controller.rephraseMessage(payload, createUnauthenticatedHeader());

            verify(bedrockService, never()).rephrase(anyString(), anyString());
        }
    }

    @Nested
    @DisplayName("deleteSession")
    class DeleteSession {

        @Test
        @DisplayName("セッションを削除して通知する")
        void deletesAndNotifies() {
            Map<String, Object> payload = Map.of(
                "sessionId", 10
            );

            controller.deleteSession(payload, headerAccessor);

            verify(deleteAiChatSessionUseCase).execute(10, 5);
            verify(messagingTemplate).convertAndSend(
                eq("/topic/ai-chat/user/5/session-deleted"),
                any(Map.class)
            );
        }

        @Test
        @DisplayName("未認証ユーザーの場合はセッション削除しない")
        void unauthenticated_doesNotDelete() {
            Map<String, Object> payload = Map.of(
                "sessionId", 10
            );

            controller.deleteSession(payload, createUnauthenticatedHeader());

            verify(deleteAiChatSessionUseCase, never()).execute(anyInt(), anyInt());
        }
    }

}
