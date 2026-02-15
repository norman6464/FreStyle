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
            AiChatSessionDto dto = new AiChatSessionDto();
            dto.setId(1);
            when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
            when(getAiChatSessionByIdUseCase.execute(1, 10)).thenReturn(dto);

            ResponseEntity<AiChatSessionDto> response = aiChatController.getSession(jwt, 1);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(response.getBody().getId()).isEqualTo(1);
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
}
