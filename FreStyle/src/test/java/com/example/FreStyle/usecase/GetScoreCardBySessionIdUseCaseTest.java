package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.List;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.ScoreCardDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.CommunicationScore;
import com.example.FreStyle.mapper.ScoreCardMapper;
import com.example.FreStyle.repository.CommunicationScoreRepository;

/**
 * GetScoreCardBySessionIdUseCaseのテストクラス
 *
 * <p>テスト対象:</p>
 * <ul>
 *   <li>セッションIDによるスコアカード取得</li>
 *   <li>エンティティからDTOへの変換</li>
 *   <li>スコアが0件の場合のハンドリング</li>
 * </ul>
 */
@ExtendWith(MockitoExtension.class)
class GetScoreCardBySessionIdUseCaseTest {

    @Mock
    private CommunicationScoreRepository communicationScoreRepository;

    @Mock
    private ScoreCardMapper mapper;

    @InjectMocks
    private GetScoreCardBySessionIdUseCase useCase;

    @Nested
    @DisplayName("execute - セッションIDでスコアカード取得")
    class ExecuteTest {

        @Test
        @DisplayName("セッションIDに紐づくスコアカードをDTOで取得する")
        void shouldReturnScoreCardDtoBySessionId() {
            // Arrange
            Integer sessionId = 1;
            List<CommunicationScore> entities = List.of(
                    createEntity(1, sessionId, "論理的構成力", 8, "良い"),
                    createEntity(2, sessionId, "配慮表現", 6, "改善余地")
            );

            ScoreCardDto expectedDto = new ScoreCardDto(
                    sessionId,
                    List.of(
                            new ScoreCardDto.AxisScoreDto("論理的構成力", 8, "良い"),
                            new ScoreCardDto.AxisScoreDto("配慮表現", 6, "改善余地")
                    ),
                    7.0
            );

            when(communicationScoreRepository.findBySessionId(sessionId)).thenReturn(entities);
            when(mapper.toScoreCardDto(sessionId, entities)).thenReturn(expectedDto);

            // Act
            ScoreCardDto result = useCase.execute(sessionId);

            // Assert
            assertThat(result.sessionId()).isEqualTo(sessionId);
            assertThat(result.scores()).hasSize(2);
            assertThat(result.overallScore()).isEqualTo(7.0);

            verify(communicationScoreRepository, times(1)).findBySessionId(sessionId);
            verify(mapper, times(1)).toScoreCardDto(sessionId, entities);
        }

        @Test
        @DisplayName("スコアが0件の場合でも正常にDTOを返す")
        void shouldReturnEmptyScoreCardWhenNoScores() {
            // Arrange
            Integer sessionId = 99;
            List<CommunicationScore> emptyList = List.of();

            ScoreCardDto expectedDto = new ScoreCardDto(sessionId, List.of(), 0.0);

            when(communicationScoreRepository.findBySessionId(sessionId)).thenReturn(emptyList);
            when(mapper.toScoreCardDto(sessionId, emptyList)).thenReturn(expectedDto);

            // Act
            ScoreCardDto result = useCase.execute(sessionId);

            // Assert
            assertThat(result.sessionId()).isEqualTo(sessionId);
            assertThat(result.scores()).isEmpty();
            assertThat(result.overallScore()).isEqualTo(0.0);

            verify(communicationScoreRepository, times(1)).findBySessionId(sessionId);
            verify(mapper, times(1)).toScoreCardDto(sessionId, emptyList);
        }
    }

        @Test
        @DisplayName("異なるセッションIDでも正しく取得する")
        void shouldReturnScoreCardForDifferentSessionId() {
            Integer sessionId = 42;
            List<CommunicationScore> entities = List.of(
                    createEntity(10, sessionId, "要約力", 9, "素晴らしい")
            );

            ScoreCardDto expectedDto = new ScoreCardDto(
                    sessionId,
                    List.of(new ScoreCardDto.AxisScoreDto("要約力", 9, "素晴らしい")),
                    9.0
            );

            when(communicationScoreRepository.findBySessionId(sessionId)).thenReturn(entities);
            when(mapper.toScoreCardDto(sessionId, entities)).thenReturn(expectedDto);

            ScoreCardDto result = useCase.execute(sessionId);

            assertThat(result.sessionId()).isEqualTo(42);
            assertThat(result.scores()).hasSize(1);
            assertThat(result.scores().getFirst().axis()).isEqualTo("要約力");
            verify(communicationScoreRepository).findBySessionId(sessionId);
        }

        @Test
        @DisplayName("スコアが多い場合でも正しく取得する")
        void shouldReturnScoreCardWithManyScores() {
            Integer sessionId = 5;
            List<CommunicationScore> entities = List.of(
                    createEntity(1, sessionId, "論理的構成力", 8, "良い"),
                    createEntity(2, sessionId, "配慮表現", 6, "改善余地"),
                    createEntity(3, sessionId, "要約力", 7, "普通"),
                    createEntity(4, sessionId, "提案力", 9, "素晴らしい")
            );

            ScoreCardDto expectedDto = new ScoreCardDto(
                    sessionId,
                    List.of(
                            new ScoreCardDto.AxisScoreDto("論理的構成力", 8, "良い"),
                            new ScoreCardDto.AxisScoreDto("配慮表現", 6, "改善余地"),
                            new ScoreCardDto.AxisScoreDto("要約力", 7, "普通"),
                            new ScoreCardDto.AxisScoreDto("提案力", 9, "素晴らしい")
                    ),
                    7.5
            );

            when(communicationScoreRepository.findBySessionId(sessionId)).thenReturn(entities);
            when(mapper.toScoreCardDto(sessionId, entities)).thenReturn(expectedDto);

            ScoreCardDto result = useCase.execute(sessionId);

            assertThat(result.scores()).hasSize(4);
            assertThat(result.overallScore()).isEqualTo(7.5);
            verify(communicationScoreRepository).findBySessionId(sessionId);
            verify(mapper).toScoreCardDto(sessionId, entities);
        }

    // ヘルパーメソッド
    private CommunicationScore createEntity(Integer id, Integer sessionId, String axisName, Integer score, String comment) {
        CommunicationScore entity = new CommunicationScore();
        entity.setId(id);

        AiChatSession session = new AiChatSession();
        session.setId(sessionId);
        entity.setSession(session);

        entity.setAxisName(axisName);
        entity.setScore(score);
        entity.setComment(comment);
        return entity;
    }
}
