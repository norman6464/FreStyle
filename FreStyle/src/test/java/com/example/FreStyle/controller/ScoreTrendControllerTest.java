package com.example.FreStyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.*;

import java.util.List;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.oauth2.jwt.Jwt;

import com.example.FreStyle.dto.AxisAnalysisDto;
import com.example.FreStyle.dto.CommunicationScoreSummaryDto;
import com.example.FreStyle.dto.ScoreTrendDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.GetAxisAnalysisUseCase;
import com.example.FreStyle.usecase.GetCommunicationScoreSummaryUseCase;
import com.example.FreStyle.usecase.GetScoreTrendUseCase;

@ExtendWith(MockitoExtension.class)
@DisplayName("ScoreTrendController テスト")
class ScoreTrendControllerTest {

    @Mock
    private GetScoreTrendUseCase getScoreTrendUseCase;

    @Mock
    private GetAxisAnalysisUseCase getAxisAnalysisUseCase;

    @Mock
    private GetCommunicationScoreSummaryUseCase getCommunicationScoreSummaryUseCase;

    @Mock
    private UserIdentityService userIdentityService;

    @InjectMocks
    private ScoreTrendController scoreTrendController;

    @Test
    @DisplayName("スコアトレンドを取得できる")
    void getTrendReturnsTrend() {
        Jwt jwt = mock(Jwt.class);
        when(jwt.getSubject()).thenReturn("sub-123");
        User user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
        ScoreTrendDto dto = new ScoreTrendDto(30, List.of(), 0.0, null, 0, null);
        when(getScoreTrendUseCase.execute(1, 30)).thenReturn(dto);

        ResponseEntity<ScoreTrendDto> response = scoreTrendController.getTrend(jwt, 30);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        assertThat(response.getBody()).isEqualTo(dto);
    }

    @Test
    @DisplayName("UseCaseに正しいuserIdとdaysを渡している")
    void getTrendPassesCorrectArguments() {
        Jwt jwt = mock(Jwt.class);
        when(jwt.getSubject()).thenReturn("sub-456");
        User user = new User();
        user.setId(42);
        when(userIdentityService.findUserBySub("sub-456")).thenReturn(user);
        ScoreTrendDto dto = new ScoreTrendDto(7, List.of(), 0.0, null, 0, null);
        when(getScoreTrendUseCase.execute(42, 7)).thenReturn(dto);

        scoreTrendController.getTrend(jwt, 7);

        verify(getScoreTrendUseCase).execute(42, 7);
    }

    @Test
    @DisplayName("軸別分析を取得できる")
    void getAxisAnalysisReturnsAnalysis() {
        Jwt jwt = mock(Jwt.class);
        when(jwt.getSubject()).thenReturn("sub-123");
        User user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
        AxisAnalysisDto dto = new AxisAnalysisDto(
                List.of(new AxisAnalysisDto.AxisStat("傾聴力", 8.0, 3)),
                "傾聴力", "傾聴力", 3);
        when(getAxisAnalysisUseCase.execute(1)).thenReturn(dto);

        ResponseEntity<AxisAnalysisDto> response = scoreTrendController.getAxisAnalysis(jwt);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        assertThat(response.getBody()).isEqualTo(dto);
    }

    @Test
    @DisplayName("スコアサマリーを取得できる")
    void getSummaryReturnsSummary() {
        Jwt jwt = mock(Jwt.class);
        when(jwt.getSubject()).thenReturn("sub-123");
        User user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
        CommunicationScoreSummaryDto dto = new CommunicationScoreSummaryDto(
                3, 7.5, List.of(), "論理的構成力", "配慮表現");
        when(getCommunicationScoreSummaryUseCase.execute(1)).thenReturn(dto);

        ResponseEntity<CommunicationScoreSummaryDto> response = scoreTrendController.getSummary(jwt);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        assertThat(response.getBody()).isEqualTo(dto);
        verify(getCommunicationScoreSummaryUseCase).execute(1);
    }
}
