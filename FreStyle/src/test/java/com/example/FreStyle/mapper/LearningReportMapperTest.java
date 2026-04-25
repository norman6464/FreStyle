package com.example.FreStyle.mapper;

import static org.assertj.core.api.Assertions.assertThat;

import java.time.LocalDateTime;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;

import com.example.FreStyle.dto.LearningReportDto;
import com.example.FreStyle.entity.LearningReport;

/**
 * LearningReportMapper のDTO変換テスト。
 * 前月平均スコアの有無によるscoreChange計算ロジックを網羅する。
 */
@DisplayName("LearningReportMapper")
class LearningReportMapperTest {

    private final LearningReportMapper mapper = new LearningReportMapper();

    private LearningReport createReport(
            Integer id,
            Integer year,
            Integer month,
            Integer totalSessions,
            Double averageScore,
            Double previousAverageScore,
            String bestAxis,
            String worstAxis,
            Integer practiceDays,
            LocalDateTime createdAt) {
        LearningReport report = new LearningReport();
        report.setId(id);
        report.setYear(year);
        report.setMonth(month);
        report.setTotalSessions(totalSessions);
        report.setAverageScore(averageScore);
        report.setPreviousAverageScore(previousAverageScore);
        report.setBestAxis(bestAxis);
        report.setWorstAxis(worstAxis);
        report.setPracticeDays(practiceDays);
        report.setCreatedAt(createdAt);
        return report;
    }

    @Nested
    @DisplayName("scoreChange（前月比）の計算")
    class ScoreChangeTest {

        @Test
        @DisplayName("前月平均があるとき、scoreChange = 今月平均 - 前月平均 で計算される")
        void computesDifferenceWhenPreviousExists() {
            LearningReport report = createReport(
                    1, 2026, 2, 20, 7.8, 7.0, "論理的構成力", "配慮表現", 15,
                    LocalDateTime.of(2026, 3, 1, 0, 0));

            LearningReportDto dto = mapper.toDto(report);

            assertThat(dto.scoreChange()).isCloseTo(0.8, org.assertj.core.data.Offset.offset(0.0001));
            assertThat(dto.previousAverageScore()).isEqualTo(7.0);
        }

        @Test
        @DisplayName("今月の方が平均が低い場合、scoreChangeは負の値になる")
        void computesNegativeDifference() {
            LearningReport report = createReport(
                    1, 2026, 2, 20, 6.0, 7.0, "a", "b", 15, LocalDateTime.now());

            LearningReportDto dto = mapper.toDto(report);

            assertThat(dto.scoreChange()).isCloseTo(-1.0, org.assertj.core.data.Offset.offset(0.0001));
        }

        @Test
        @DisplayName("前月平均がnullのとき、scoreChangeもnullになる")
        void returnsNullWhenNoPrevious() {
            LearningReport report = createReport(
                    1, 2026, 1, 10, 6.5, null, "要約力", "提案力", 8, LocalDateTime.now());

            LearningReportDto dto = mapper.toDto(report);

            assertThat(dto.scoreChange()).isNull();
            assertThat(dto.previousAverageScore()).isNull();
        }
    }

    @Nested
    @DisplayName("フィールドのコピー")
    class FieldCopy {

        @Test
        @DisplayName("id / year / month / 各軸などがDTOへそのまま渡される")
        void copiesAllFields() {
            LocalDateTime createdAt = LocalDateTime.of(2026, 3, 1, 10, 15, 30);
            LearningReport report = createReport(
                    42, 2026, 2, 25, 8.1, 7.5, "論理的構成力", "配慮表現", 20, createdAt);

            LearningReportDto dto = mapper.toDto(report);

            assertThat(dto.id()).isEqualTo(42);
            assertThat(dto.year()).isEqualTo(2026);
            assertThat(dto.month()).isEqualTo(2);
            assertThat(dto.totalSessions()).isEqualTo(25);
            assertThat(dto.averageScore()).isEqualTo(8.1);
            assertThat(dto.bestAxis()).isEqualTo("論理的構成力");
            assertThat(dto.worstAxis()).isEqualTo("配慮表現");
            assertThat(dto.practiceDays()).isEqualTo(20);
            assertThat(dto.createdAt()).isEqualTo(createdAt.toString());
        }

        @Test
        @DisplayName("createdAtがnullのとき、DTOのcreatedAtもnullになる")
        void nullCreatedAtIsPassedThrough() {
            LearningReport report = createReport(
                    1, 2026, 1, 5, 6.0, null, "a", "b", 3, null);

            LearningReportDto dto = mapper.toDto(report);

            assertThat(dto.createdAt()).isNull();
        }
    }
}
