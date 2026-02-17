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
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.oauth2.jwt.Jwt;

import com.example.FreStyle.dto.LearningReportDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.GenerateMonthlyReportUseCase;
import com.example.FreStyle.usecase.GetMonthlyReportUseCase;
import com.example.FreStyle.usecase.GetReportListUseCase;

@ExtendWith(MockitoExtension.class)
@DisplayName("LearningReportController テスト")
class LearningReportControllerTest {

    @Mock
    private GenerateMonthlyReportUseCase generateMonthlyReportUseCase;

    @Mock
    private GetMonthlyReportUseCase getMonthlyReportUseCase;

    @Mock
    private GetReportListUseCase getReportListUseCase;

    @Mock
    private UserIdentityService userIdentityService;

    @InjectMocks
    private LearningReportController learningReportController;

    private Jwt mockJwt;
    private User testUser;

    @BeforeEach
    void setUp() {
        mockJwt = mock(Jwt.class);
        when(mockJwt.getSubject()).thenReturn("sub-123");

        testUser = new User();
        testUser.setId(1);
        testUser.setName("テストユーザー");

        when(userIdentityService.findUserBySub("sub-123")).thenReturn(testUser);
    }

    @Nested
    @DisplayName("getReportList")
    class GetReportList {

        @Test
        @DisplayName("レポート一覧を取得できる")
        void returnsReportList() {
            LearningReportDto dto = new LearningReportDto(1, 2026, 1, 5, 75.0, 70.0, 5.0, "論理的構成力", "配慮表現", 3, "2026-02-01T00:00:00");
            when(getReportListUseCase.execute(1)).thenReturn(List.of(dto));

            ResponseEntity<List<LearningReportDto>> response = learningReportController.getReportList(mockJwt);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(response.getBody()).hasSize(1);
        }
    }

    @Nested
    @DisplayName("getMonthlyReport")
    class GetMonthlyReport {

        @Test
        @DisplayName("月次レポートを取得できる")
        void returnsReport() {
            LearningReportDto dto = new LearningReportDto(1, 2026, 1, 5, 75.0, 70.0, 5.0, "論理的構成力", "配慮表現", 3, "2026-02-01T00:00:00");
            when(getMonthlyReportUseCase.execute(1, 2026, 1)).thenReturn(dto);

            ResponseEntity<LearningReportDto> response = learningReportController.getMonthlyReport(mockJwt, 2026, 1);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(response.getBody().getTotalSessions()).isEqualTo(5);
        }

        @Test
        @DisplayName("レポートがない場合は404を返す")
        void returnsNotFound() {
            when(getMonthlyReportUseCase.execute(1, 2026, 3)).thenReturn(null);

            ResponseEntity<LearningReportDto> response = learningReportController.getMonthlyReport(mockJwt, 2026, 3);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.NOT_FOUND);
        }
    }

    @Nested
    @DisplayName("generateReport")
    class GenerateReport {

        @Test
        @DisplayName("レポートを生成できる")
        void generatesReport() {
            LearningReportDto dto = new LearningReportDto(1, 2026, 1, 5, 75.0, 70.0, 5.0, "論理的構成力", "配慮表現", 3, "2026-02-01T00:00:00");
            when(generateMonthlyReportUseCase.execute(testUser, 2026, 1)).thenReturn(dto);

            ResponseEntity<LearningReportDto> response = learningReportController.generateReport(mockJwt, new LearningReportController.GenerateReportRequest(2026, 1));

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            verify(generateMonthlyReportUseCase).execute(testUser, 2026, 1);
        }
    }
}
