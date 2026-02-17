package com.example.FreStyle.dto;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;

class ScoreCardDtoTest {

    @Test
    @DisplayName("スコアカードが正しく作成される")
    void constructorWorks() {
        List<ScoreCardDto.AxisScoreDto> scores = List.of(
                new ScoreCardDto.AxisScoreDto("論理的構成力", 8, "良い構成です"),
                new ScoreCardDto.AxisScoreDto("表現力", 7, "改善の余地あり"));

        ScoreCardDto dto = new ScoreCardDto(1, scores, 7.5);

        assertThat(dto.sessionId()).isEqualTo(1);
        assertThat(dto.scores()).hasSize(2);
        assertThat(dto.overallScore()).isEqualTo(7.5);
    }

    @Test
    @DisplayName("AxisScoreDtoのフィールドが正しく設定される")
    void axisScoreDtoFieldsAreSet() {
        ScoreCardDto.AxisScoreDto axisScore = new ScoreCardDto.AxisScoreDto("傾聴力", 9, "素晴らしい");

        assertThat(axisScore.axis()).isEqualTo("傾聴力");
        assertThat(axisScore.score()).isEqualTo(9);
        assertThat(axisScore.comment()).isEqualTo("素晴らしい");
    }
}
