package com.example.FreStyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.List;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.ResponseEntity;
import org.springframework.security.oauth2.jwt.Jwt;

import com.example.FreStyle.dto.ScoreCardDto;
import com.example.FreStyle.dto.ScoreHistoryDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.GetAiChatSessionByIdUseCase;
import com.example.FreStyle.usecase.GetScoreCardBySessionIdUseCase;
import com.example.FreStyle.usecase.GetScoreHistoryByUserIdUseCase;

@ExtendWith(MockitoExtension.class)
@DisplayName("ScoreCardController")
class ScoreCardControllerTest {

    @Mock
    private UserIdentityService userIdentityService;

    @Mock
    private GetAiChatSessionByIdUseCase getAiChatSessionByIdUseCase;

    @Mock
    private GetScoreCardBySessionIdUseCase getScoreCardBySessionIdUseCase;

    @Mock
    private GetScoreHistoryByUserIdUseCase getScoreHistoryByUserIdUseCase;

    @InjectMocks
    private ScoreCardController controller;

    private Jwt jwt;
    private User user;

    @BeforeEach
    void setUp() {
        jwt = mock(Jwt.class);
        when(jwt.getSubject()).thenReturn("test-sub");

        user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("test-sub")).thenReturn(user);
    }

    @Nested
    @DisplayName("GET /api/scores/sessions/{sessionId}")
    class GetSessionScoreCard {

        @Test
        @DisplayName("スコアカードを取得する")
        void shouldReturnScoreCard() {
            ScoreCardDto scoreCard = new ScoreCardDto(10, List.of(), 8.5);
            when(getScoreCardBySessionIdUseCase.execute(10)).thenReturn(scoreCard);

            ResponseEntity<ScoreCardDto> response = controller.getSessionScoreCard(jwt, 10);

            assertThat(response.getStatusCode().value()).isEqualTo(200);
            assertThat(response.getBody()).isNotNull();
            assertThat(response.getBody().sessionId()).isEqualTo(10);
            assertThat(response.getBody().overallScore()).isEqualTo(8.5);
        }

        @Test
        @DisplayName("権限チェックが実行される")
        void shouldCheckAuthorization() {
            ScoreCardDto scoreCard = new ScoreCardDto(10, List.of(), 0.0);
            when(getScoreCardBySessionIdUseCase.execute(10)).thenReturn(scoreCard);

            controller.getSessionScoreCard(jwt, 10);

            verify(getAiChatSessionByIdUseCase).execute(10, 1);
        }
    }

    @Nested
    @DisplayName("GET /api/scores/history")
    class GetScoreHistory {

        @Test
        @DisplayName("スコア履歴を取得する")
        void shouldReturnScoreHistory() {
            ScoreHistoryDto history1 = new ScoreHistoryDto(1, null, 7.5, List.of(), null);
            ScoreHistoryDto history2 = new ScoreHistoryDto(2, null, 8.0, List.of(), null);
            when(getScoreHistoryByUserIdUseCase.execute(1))
                    .thenReturn(List.of(history1, history2));

            ResponseEntity<List<ScoreHistoryDto>> response = controller.getScoreHistory(jwt);

            assertThat(response.getStatusCode().value()).isEqualTo(200);
            assertThat(response.getBody()).hasSize(2);
        }

        @Test
        @DisplayName("UseCaseにユーザーIDを渡す")
        void shouldPassUserIdToUseCase() {
            when(getScoreHistoryByUserIdUseCase.execute(1)).thenReturn(List.of());

            controller.getScoreHistory(jwt);

            verify(getScoreHistoryByUserIdUseCase).execute(1);
        }
    }
}
