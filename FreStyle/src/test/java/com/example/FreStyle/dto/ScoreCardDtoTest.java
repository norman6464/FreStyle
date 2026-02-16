package com.example.FreStyle.dto;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;

class ScoreCardDtoTest {

    @Test
    @DisplayName("AllArgsConstructorでスコアカードが作成される")
    void allArgsConstructorWorks() {
        List<ScoreCardDto.AxisScoreDto> scores = List.of(
                new ScoreCardDto.AxisScoreDto("論理的構成力", 8, "良い構成です"),
                new ScoreCardDto.AxisScoreDto("表現力", 7, "改善の余地あり"));

        ScoreCardDto dto = new ScoreCardDto(1, scores, 7.5);

        assertThat(dto.getSessionId()).isEqualTo(1);
        assertThat(dto.getScores()).hasSize(2);
        assertThat(dto.getOverallScore()).isEqualTo(7.5);
    }

    @Test
    @DisplayName("AxisScoreDtoのフィールドが正しく設定される")
    void axisScoreDtoFieldsAreSet() {
        ScoreCardDto.AxisScoreDto axisScore = new ScoreCardDto.AxisScoreDto("傾聴力", 9, "素晴らしい");

        assertThat(axisScore.getAxis()).isEqualTo("傾聴力");
        assertThat(axisScore.getScore()).isEqualTo(9);
        assertThat(axisScore.getComment()).isEqualTo("素晴らしい");
    }

    @Test
    @DisplayName("NoArgsConstructorで空のスコアカードが作成される")
    void noArgsConstructorWorks() {
        ScoreCardDto dto = new ScoreCardDto();

        assertThat(dto.getSessionId()).isNull();
        assertThat(dto.getScores()).isNull();
        assertThat(dto.getOverallScore()).isZero();
    }

    @Test
    @DisplayName("AxisScoreDto NoArgsConstructorで空のスコアが作成される")
    void axisScoreDtoNoArgsConstructorWorks() {
        ScoreCardDto.AxisScoreDto axisScore = new ScoreCardDto.AxisScoreDto();

        assertThat(axisScore.getAxis()).isNull();
        assertThat(axisScore.getScore()).isZero();
        assertThat(axisScore.getComment()).isNull();
    }
}
