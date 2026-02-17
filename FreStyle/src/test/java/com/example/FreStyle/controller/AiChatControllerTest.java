package com.example.FreStyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.*;

import java.util.List;
import java.util.Map;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.oauth2.jwt.Jwt;

import com.example.FreStyle.dto.AiChatMessageDto;
import com.example.FreStyle.dto.AiChatMessageResponseDto;
import com.example.FreStyle.dto.AiChatSessionDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.AiChatService;
import com.example.FreStyle.service.BedrockService;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.CreateAiChatSessionUseCase;
import com.example.FreStyle.usecase.DeleteAiChatSessionUseCase;
import com.example.FreStyle.usecase.GetAiChatSessionByIdUseCase;
import com.example.FreStyle.usecase.GetAiChatSessionsByUserIdUseCase;
import com.example.FreStyle.usecase.UpdateAiChatSessionTitleUseCase;
import com.example.FreStyle.usecase.GetAiChatMessagesBySessionIdUseCase;
import com.example.FreStyle.usecase.AddAiChatMessageUseCase;

@ExtendWith(MockitoExtension.class)
@DisplayName("AiChatController")
class AiChatControllerTest {

    @Mock private AiChatService aiChatService;
    @Mock private UserIdentityService userIdentityService;
    @Mock private BedrockService bedrockService;
    @Mock private GetAiChatSessionsByUserIdUseCase getAiChatSessionsByUserIdUseCase;
    @Mock private CreateAiChatSessionUseCase createAiChatSessionUseCase;
    @Mock private GetAiChatSessionByIdUseCase getAiChatSessionByIdUseCase;
    @Mock private UpdateAiChatSessionTitleUseCase updateAiChatSessionTitleUseCase;
    @Mock private DeleteAiChatSessionUseCase deleteAiChatSessionUseCase;
    @Mock private GetAiChatMessagesBySessionIdUseCase getAiChatMessagesBySessionIdUseCase;
    @Mock private AddAiChatMessageUseCase addAiChatMessageUseCase;

    @InjectMocks
    private AiChatController aiChatController;

    private Jwt mockJwt(String sub) {
        Jwt jwt = mock(Jwt.class);
        when(jwt.getSubject()).thenReturn(sub);
        return jwt;
    }

    private User createUser(Integer id) {
        User user = new User();
        user.setId(id);
        user.setName("テストユーザー");
        return user;
    }

    @Nested
    @DisplayName("getSessions")
    class GetSessions {

        @Test
        @DisplayName("セッション一覧を返す")
        void returnsSessions() {
            Jwt jwt = mockJwt("sub-123");
            User user = createUser(10);
            when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
            when(getAiChatSessionsByUserIdUseCase.execute(10)).thenReturn(List.of());

            ResponseEntity<List<AiChatSessionDto>> response = aiChatController.getSessions(jwt);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(response.getBody()).isEmpty();
        }
    }

    @Nested
    @DisplayName("getSession")
    class GetSession {

        @Test
        @DisplayName("セッション詳細を返す")
        void returnsSession() {
            Jwt jwt = mockJwt("sub-123");
            User user = createUser(10);
            AiChatSessionDto dto = new AiChatSessionDto(1, null, null, null, null, null, null, null, null);
            when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
            when(getAiChatSessionByIdUseCase.execute(1, 10)).thenReturn(dto);

            ResponseEntity<AiChatSessionDto> response = aiChatController.getSession(jwt, 1);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(response.getBody().id()).isEqualTo(1);
        }
    }

    @Nested
    @DisplayName("deleteSession")
    class DeleteSession {

        @Test
        @DisplayName("セッションを削除する")
        void deletesSession() {
            Jwt jwt = mockJwt("sub-123");
            User user = createUser(10);
            when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);

            ResponseEntity<Void> response = aiChatController.deleteSession(jwt, 1);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.NO_CONTENT);
            verify(deleteAiChatSessionUseCase).execute(1, 10);
        }
    }

    @Nested
    @DisplayName("rephrase")
    class Rephrase {

        @Test
        @DisplayName("言い換え結果を返す")
        void returnsRephraseResult() {
            Jwt jwt = mockJwt("sub-123");
            User user = createUser(10);
            when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
            when(bedrockService.rephrase("テスト", "会議")).thenReturn("言い換え結果");

            // Use reflection-free approach: create record via constructor
            var request = new AiChatController.RephraseRequest("テスト", "会議");
            ResponseEntity<Map<String, String>> response = aiChatController.rephrase(jwt, request);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(response.getBody().get("result")).isEqualTo("言い換え結果");
        }
    }

    @Nested
    @DisplayName("getChatHistory")
    class GetChatHistory {

        @Test
        @DisplayName("AI履歴を返す")
        void returnsChatHistory() {
            Jwt jwt = mockJwt("sub-123");
            User user = createUser(10);
            when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
            List<AiChatMessageDto> history = List.of(
                    new AiChatMessageDto("こんにちは", true, System.currentTimeMillis()));
            when(aiChatService.getChatHistory(10)).thenReturn(history);

            ResponseEntity<?> response = aiChatController.getChatHistory(jwt);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(response.getBody()).isEqualTo(history);
        }
    }

    @Nested
    @DisplayName("createSession")
    class CreateSession {

        @Test
        @DisplayName("新規セッションを作成する")
        void createsSession() {
            Jwt jwt = mockJwt("sub-123");
            User user = createUser(10);
            AiChatSessionDto dto = new AiChatSessionDto(5, null, null, null, null, null, null, null, null);
            when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
            when(createAiChatSessionUseCase.execute(10, "テストセッション", null)).thenReturn(dto);

            var request = new AiChatController.CreateSessionRequest("テストセッション", null);
            ResponseEntity<AiChatSessionDto> response = aiChatController.createSession(jwt, request);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(response.getBody().id()).isEqualTo(5);
        }
    }

    @Nested
    @DisplayName("updateSessionTitle")
    class UpdateSessionTitle {

        @Test
        @DisplayName("セッションタイトルを更新する")
        void updatesTitle() {
            Jwt jwt = mockJwt("sub-123");
            User user = createUser(10);
            AiChatSessionDto dto = new AiChatSessionDto(1, null, null, null, null, null, null, null, null);
            when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
            when(updateAiChatSessionTitleUseCase.execute(1, 10, "新しいタイトル")).thenReturn(dto);

            var request = new AiChatController.UpdateSessionRequest("新しいタイトル");
            ResponseEntity<AiChatSessionDto> response = aiChatController.updateSessionTitle(jwt, 1, request);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            verify(updateAiChatSessionTitleUseCase).execute(1, 10, "新しいタイトル");
        }
    }

    @Nested
    @DisplayName("getMessages")
    class GetMessages {

        @Test
        @DisplayName("セッション内のメッセージ一覧を返す")
        void returnsMessages() {
            Jwt jwt = mockJwt("sub-123");
            User user = createUser(10);
            AiChatSessionDto sessionDto = new AiChatSessionDto(1, null, null, null, null, null, null, null, null);
            when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
            when(getAiChatSessionByIdUseCase.execute(1, 10)).thenReturn(sessionDto);
            when(getAiChatMessagesBySessionIdUseCase.execute(1)).thenReturn(List.of());

            ResponseEntity<List<AiChatMessageResponseDto>> response = aiChatController.getMessages(jwt, 1);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(response.getBody()).isEmpty();
            verify(getAiChatSessionByIdUseCase).execute(1, 10);
        }
    }

    @Nested
    @DisplayName("addMessage")
    class AddMessage {

        @Test
        @DisplayName("ユーザーメッセージを追加する")
        void addsUserMessage() {
            Jwt jwt = mockJwt("sub-123");
            User user = createUser(10);
            AiChatSessionDto sessionDto = new AiChatSessionDto(1, null, null, null, null, null, null, null, null);
            AiChatMessageResponseDto messageDto = new AiChatMessageResponseDto(100, null, null, null, null, null);
            when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
            when(getAiChatSessionByIdUseCase.execute(1, 10)).thenReturn(sessionDto);
            when(addAiChatMessageUseCase.executeUserMessage(1, 10, "テスト")).thenReturn(messageDto);

            var request = new AiChatController.AddMessageRequest("テスト", "user");
            ResponseEntity<AiChatMessageResponseDto> response = aiChatController.addMessage(jwt, 1, request);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(response.getBody().id()).isEqualTo(100);
        }

        @Test
        @DisplayName("アシスタントメッセージを追加する")
        void addsAssistantMessage() {
            Jwt jwt = mockJwt("sub-123");
            User user = createUser(10);
            AiChatSessionDto sessionDto = new AiChatSessionDto(1, null, null, null, null, null, null, null, null);
            AiChatMessageResponseDto messageDto = new AiChatMessageResponseDto(101, null, null, null, null, null);
            when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
            when(getAiChatSessionByIdUseCase.execute(1, 10)).thenReturn(sessionDto);
            when(addAiChatMessageUseCase.executeAssistantMessage(1, 10, "AI応答")).thenReturn(messageDto);

            var request = new AiChatController.AddMessageRequest("AI応答", "assistant");
            ResponseEntity<AiChatMessageResponseDto> response = aiChatController.addMessage(jwt, 1, request);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(response.getBody().id()).isEqualTo(101);
        }
    }
}
