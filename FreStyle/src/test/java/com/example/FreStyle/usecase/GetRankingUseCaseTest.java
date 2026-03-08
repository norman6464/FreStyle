package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.List;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.RankingDto;
import com.example.FreStyle.dto.RankingDto.RankingEntryDto;
import com.example.FreStyle.service.RankingService;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetRankingUseCase")
class GetRankingUseCaseTest {

    @Mock
    private RankingService rankingService;

    @InjectMocks
    private GetRankingUseCase useCase;

    @Test
    @DisplayName("RankingServiceに期間とユーザーIDを委譲する")
    void shouldDelegateToRankingService() {
        RankingEntryDto entry = new RankingEntryDto(1, 1, "testUser", null, 8.5, 3);
        RankingDto expected = new RankingDto(List.of(entry), entry);
        when(rankingService.getRanking("weekly", 1)).thenReturn(expected);

        RankingDto result = useCase.execute("weekly", 1);

        verify(rankingService).getRanking("weekly", 1);
        assertThat(result).isEqualTo(expected);
    }

    @Test
    @DisplayName("RankingServiceに正しい引数を委譲する")
    void shouldDelegateCorrectPeriodAndUserId() {
        RankingDto expected = new RankingDto(List.of(), null);
        when(rankingService.getRanking("monthly", 42)).thenReturn(expected);

        RankingDto result = useCase.execute("monthly", 42);

        verify(rankingService).getRanking("monthly", 42);
        assertThat(result).isEqualTo(expected);
    }
}
