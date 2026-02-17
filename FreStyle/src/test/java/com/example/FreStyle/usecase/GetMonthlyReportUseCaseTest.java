package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.when;

import java.time.LocalDateTime;
import java.util.Optional;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.Spy;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.LearningReportDto;
import com.example.FreStyle.entity.LearningReport;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.mapper.LearningReportMapper;
import com.example.FreStyle.repository.LearningReportRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetMonthlyReportUseCase テスト")
class GetMonthlyReportUseCaseTest {

    @Mock
    private LearningReportRepository learningReportRepository;

    @Spy
    private LearningReportMapper learningReportMapper;

    @InjectMocks
    private GetMonthlyReportUseCase getMonthlyReportUseCase;

    @Test
    @DisplayName("月次レポートを取得できる")
    void execute_returnsReport() {
        User user = new User();
        user.setId(1);
        LearningReport report = new LearningReport();
        report.setId(10);
        report.setUser(user);
        report.setYear(2026);
        report.setMonth(1);
        report.setTotalSessions(5);
        report.setAverageScore(75.0);
        report.setPreviousAverageScore(70.0);
        report.setBestAxis("論理的構成力");
        report.setWorstAxis("配慮表現");
        report.setPracticeDays(3);
        report.setCreatedAt(LocalDateTime.of(2026, 2, 1, 0, 0));

        when(learningReportRepository.findByUserIdAndYearAndMonth(1, 2026, 1))
                .thenReturn(Optional.of(report));

        LearningReportDto result = getMonthlyReportUseCase.execute(1, 2026, 1);

        assertThat(result).isNotNull();
        assertThat(result.totalSessions()).isEqualTo(5);
        assertThat(result.averageScore()).isEqualTo(75.0);
        assertThat(result.scoreChange()).isEqualTo(5.0);
        assertThat(result.bestAxis()).isEqualTo("論理的構成力");
    }

    @Test
    @DisplayName("レポートが存在しない場合はnullを返す")
    void execute_returnsNull() {
        when(learningReportRepository.findByUserIdAndYearAndMonth(1, 2026, 3))
                .thenReturn(Optional.empty());

        LearningReportDto result = getMonthlyReportUseCase.execute(1, 2026, 3);

        assertThat(result).isNull();
    }
}
