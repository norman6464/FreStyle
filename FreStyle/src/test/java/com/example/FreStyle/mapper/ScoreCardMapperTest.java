package com.example.FreStyle.mapper;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

import java.sql.Timestamp;
import java.util.List;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;

import com.example.FreStyle.dto.ScoreCardDto;
import com.example.FreStyle.dto.ScoreHistoryDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.CommunicationScore;

@DisplayName("ScoreCardMapper")
class ScoreCardMapperTest {

    private final ScoreCardMapper mapper = new ScoreCardMapper();

    private CommunicationScore createScore(AiChatSession session, String axis, int score, String comment) {
        CommunicationScore cs = new CommunicationScore();
        cs.setSession(session);
        cs.setAxisName(axis);
        cs.setScore(score);
        cs.setComment(comment);
        cs.setCreatedAt(new Timestamp(System.currentTimeMillis()));
        return cs;
    }

    private AiChatSession createSession(Integer id, String title) {
        AiChatSession session = new AiChatSession();
        session.setId(id);
        session.setTitle(title);
        return session;
    }

    @Nested
    @DisplayName("toScoreCardDto")
    class ToScoreCardDto {

        @Test
        @DisplayName("正常にDTOへ変換できる")
        void convertsToDtoSuccessfully() {
            AiChatSession session = createSession(1, "テストセッション");
            List<CommunicationScore> scores = List.of(
                createScore(session, "明瞭性", 8, "良い"),
                createScore(session, "論理性", 6, "改善の余地あり")
            );

            ScoreCardDto dto = mapper.toScoreCardDto(1, scores);

            assertThat(dto.sessionId()).isEqualTo(1);
            assertThat(dto.scores()).hasSize(2);
            assertThat(dto.scores().get(0).axis()).isEqualTo("明瞭性");
            assertThat(dto.scores().get(0).score()).isEqualTo(8);
            assertThat(dto.scores().get(0).comment()).isEqualTo("良い");
        }

        @Test
        @DisplayName("平均スコアが正しく計算される")
        void calculatesOverallScore() {
            AiChatSession session = createSession(1, "テスト");
            List<CommunicationScore> scores = List.of(
                createScore(session, "明瞭性", 8, ""),
                createScore(session, "論理性", 6, ""),
                createScore(session, "共感性", 4, "")
            );

            ScoreCardDto dto = mapper.toScoreCardDto(1, scores);

            assertThat(dto.overallScore()).isEqualTo(6.0);
        }

        @Test
        @DisplayName("空リストの場合は平均スコア0.0を返す")
        void returnsZeroForEmptyList() {
            ScoreCardDto dto = mapper.toScoreCardDto(1, List.of());

            assertThat(dto.scores()).isEmpty();
            assertThat(dto.overallScore()).isEqualTo(0.0);
        }

        @Test
        @DisplayName("nullの場合はNullPointerExceptionを投げる")
        void throwsOnNull() {
            assertThatThrownBy(() -> mapper.toScoreCardDto(1, null))
                .isInstanceOf(NullPointerException.class)
                .hasMessage("scoresがnullです");
        }
    }

    @Nested
    @DisplayName("toScoreHistoryDtoList")
    class ToScoreHistoryDtoList {

        @Test
        @DisplayName("セッション単位でグループ化される")
        void groupsBySession() {
            AiChatSession session1 = createSession(1, "セッション1");
            AiChatSession session2 = createSession(2, "セッション2");
            List<CommunicationScore> scores = List.of(
                createScore(session1, "明瞭性", 8, ""),
                createScore(session1, "論理性", 6, ""),
                createScore(session2, "明瞭性", 9, ""),
                createScore(session2, "論理性", 7, "")
            );

            List<ScoreHistoryDto> history = mapper.toScoreHistoryDtoList(scores);

            assertThat(history).hasSize(2);
            assertThat(history.get(0).sessionId()).isEqualTo(1);
            assertThat(history.get(0).sessionTitle()).isEqualTo("セッション1");
            assertThat(history.get(0).scores()).hasSize(2);
            assertThat(history.get(1).sessionId()).isEqualTo(2);
        }

        @Test
        @DisplayName("各セッションの平均スコアが正しく計算される")
        void calculatesOverallScorePerSession() {
            AiChatSession session = createSession(1, "テスト");
            List<CommunicationScore> scores = List.of(
                createScore(session, "明瞭性", 10, ""),
                createScore(session, "論理性", 6, "")
            );

            List<ScoreHistoryDto> history = mapper.toScoreHistoryDtoList(scores);

            assertThat(history.get(0).overallScore()).isEqualTo(8.0);
        }

        @Test
        @DisplayName("空リストの場合は空リストを返す")
        void returnsEmptyForEmptyList() {
            List<ScoreHistoryDto> history = mapper.toScoreHistoryDtoList(List.of());

            assertThat(history).isEmpty();
        }

        @Test
        @DisplayName("nullの場合はNullPointerExceptionを投げる")
        void throwsOnNull() {
            assertThatThrownBy(() -> mapper.toScoreHistoryDtoList(null))
                .isInstanceOf(NullPointerException.class)
                .hasMessage("scoresがnullです");
        }

        @Test
        @DisplayName("createdAtが設定される")
        void setsCreatedAt() {
            AiChatSession session = createSession(1, "テスト");
            List<CommunicationScore> scores = List.of(
                createScore(session, "明瞭性", 8, "")
            );

            List<ScoreHistoryDto> history = mapper.toScoreHistoryDtoList(scores);

            assertThat(history.get(0).createdAt()).isNotNull();
        }
    }
}
