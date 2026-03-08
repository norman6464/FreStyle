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

import com.example.FreStyle.dto.RankingDto;
import com.example.FreStyle.dto.RankingDto.RankingEntryDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.GetRankingUseCase;

@ExtendWith(MockitoExtension.class)
@DisplayName("RankingController")
class RankingControllerTest {

    @Mock
    private UserIdentityService userIdentityService;

    @Mock
    private GetRankingUseCase getRankingUseCase;

    @InjectMocks
    private RankingController controller;

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
    @DisplayName("GET /api/ranking")
    class GetRanking {

        @Test
        @DisplayName("デフォルト期間（weekly）でランキングを取得する")
        void shouldReturnRankingWithDefaultPeriod() {
            RankingEntryDto entry = new RankingEntryDto(1, 1, "testUser", null, 8.5, 3);
            RankingDto ranking = new RankingDto(List.of(entry), entry);
            when(getRankingUseCase.execute("weekly", 1)).thenReturn(ranking);

            ResponseEntity<RankingDto> response = controller.getRanking(jwt, "weekly");

            assertThat(response.getStatusCode().value()).isEqualTo(200);
            assertThat(response.getBody()).isNotNull();
            assertThat(response.getBody().entries()).hasSize(1);
        }

        @Test
        @DisplayName("指定した期間でランキングを取得する")
        void shouldReturnRankingWithSpecifiedPeriod() {
            RankingEntryDto entry = new RankingEntryDto(1, 2, "user2", null, 9.0, 5);
            RankingDto ranking = new RankingDto(List.of(entry), null);
            when(getRankingUseCase.execute("monthly", 1)).thenReturn(ranking);

            ResponseEntity<RankingDto> response = controller.getRanking(jwt, "monthly");

            assertThat(response.getStatusCode().value()).isEqualTo(200);
            assertThat(response.getBody()).isNotNull();
            assertThat(response.getBody().entries()).hasSize(1);
        }

        @Test
        @DisplayName("UseCaseにユーザーIDと期間を渡す")
        void shouldPassUserIdAndPeriodToUseCase() {
            RankingDto ranking = new RankingDto(List.of(), null);
            when(getRankingUseCase.execute("weekly", 1)).thenReturn(ranking);

            controller.getRanking(jwt, "weekly");

            verify(getRankingUseCase).execute("weekly", 1);
        }
    }
}
