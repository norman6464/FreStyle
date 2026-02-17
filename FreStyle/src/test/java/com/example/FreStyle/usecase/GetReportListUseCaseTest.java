package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.when;

import java.time.LocalDateTime;
import java.util.List;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.LearningReportDto;
import com.example.FreStyle.entity.LearningReport;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.LearningReportRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetReportListUseCase テスト")
class GetReportListUseCaseTest {

    @Mock
    private LearningReportRepository learningReportRepository;

    @InjectMocks
    private GetReportListUseCase getReportListUseCase;

    @Test
    @DisplayName("レポート一覧を取得できる")
    void execute_returnsReportList() {
        User user = new User();
        user.setId(1);
        LearningReport report1 = new LearningReport(1, user, 2026, 2, 8, 80.0, 75.0, "論理的構成力", "配慮表現", 5, LocalDateTime.now());
        LearningReport report2 = new LearningReport(2, user, 2026, 1, 5, 75.0, null, "要約力", "提案力", 3, LocalDateTime.now());

        when(learningReportRepository.findByUserIdOrderByYearDescMonthDesc(1))
                .thenReturn(List.of(report1, report2));

        List<LearningReportDto> result = getReportListUseCase.execute(1);

        assertThat(result).hasSize(2);
        assertThat(result.get(0).getYear()).isEqualTo(2026);
        assertThat(result.get(0).getMonth()).isEqualTo(2);
    }

    @Test
    @DisplayName("レポートがない場合は空リストを返す")
    void execute_returnsEmptyList() {
        when(learningReportRepository.findByUserIdOrderByYearDescMonthDesc(1))
                .thenReturn(List.of());

        List<LearningReportDto> result = getReportListUseCase.execute(1);

        assertThat(result).isEmpty();
    }
}
