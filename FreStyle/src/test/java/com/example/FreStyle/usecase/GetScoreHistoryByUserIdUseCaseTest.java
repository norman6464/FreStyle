package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.sql.Timestamp;
import java.util.List;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.ScoreCardDto;
import com.example.FreStyle.dto.ScoreHistoryDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.CommunicationScore;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.mapper.ScoreCardMapper;
import com.example.FreStyle.repository.CommunicationScoreRepository;

/**
 * GetScoreHistoryByUserIdUseCaseのテストクラス
 *
 * <p>テスト対象:</p>
 * <ul>
 *   <li>ユーザーIDによるスコア履歴取得</li>
 *   <li>セッション単位でのグループ化</li>
 *   <li>スコアが0件の場合のハンドリング</li>
 * </ul>
 */
@ExtendWith(MockitoExtension.class)
class GetScoreHistoryByUserIdUseCaseTest {

    @Mock
    private CommunicationScoreRepository communicationScoreRepository;

    @Mock
    private ScoreCardMapper mapper;

    @InjectMocks
    private GetScoreHistoryByUserIdUseCase useCase;

    @Nested
    @DisplayName("execute - ユーザーIDでスコア履歴取得")
    class ExecuteTest {

        @Test
        @DisplayName("ユーザーIDに紐づくスコア履歴をDTO化して取得する")
        void shouldReturnScoreHistoryByUserId() {
            // Arrange
            Integer userId = 1;
            Timestamp now = new Timestamp(System.currentTimeMillis());

            List<CommunicationScore> entities = List.of(
                    createEntity(1, 10, userId, "論理的構成力", 8, "良い", now),
                    createEntity(2, 10, userId, "配慮表現", 6, "改善余地", now)
            );

            List<ScoreHistoryDto> expectedHistory = List.of(
                    new ScoreHistoryDto(
                            10,
                            "セッション1",
                            7.0,
                            List.of(
                                    new ScoreCardDto.AxisScoreDto("論理的構成力", 8, "良い"),
                                    new ScoreCardDto.AxisScoreDto("配慮表現", 6, "改善余地")
                            ),
                            now
                    )
            );

            when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(userId)).thenReturn(entities);
            when(mapper.toScoreHistoryDtoList(entities)).thenReturn(expectedHistory);

            // Act
            List<ScoreHistoryDto> result = useCase.execute(userId);

            // Assert
            assertThat(result).hasSize(1);
            assertThat(result.get(0).getSessionId()).isEqualTo(10);
            assertThat(result.get(0).getOverallScore()).isEqualTo(7.0);
            assertThat(result.get(0).getScores()).hasSize(2);

            verify(communicationScoreRepository, times(1)).findByUserIdOrderByCreatedAtDesc(userId);
            verify(mapper, times(1)).toScoreHistoryDtoList(entities);
        }

        @Test
        @DisplayName("スコア履歴が0件の場合は空リストを返す")
        void shouldReturnEmptyListWhenNoHistory() {
            // Arrange
            Integer userId = 99;
            List<CommunicationScore> emptyList = List.of();

            when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(userId)).thenReturn(emptyList);
            when(mapper.toScoreHistoryDtoList(emptyList)).thenReturn(List.of());

            // Act
            List<ScoreHistoryDto> result = useCase.execute(userId);

            // Assert
            assertThat(result).isEmpty();

            verify(communicationScoreRepository, times(1)).findByUserIdOrderByCreatedAtDesc(userId);
            verify(mapper, times(1)).toScoreHistoryDtoList(emptyList);
        }
    }

    // ヘルパーメソッド
    private CommunicationScore createEntity(Integer id, Integer sessionId, Integer userId,
                                            String axisName, Integer score, String comment, Timestamp createdAt) {
        CommunicationScore entity = new CommunicationScore();
        entity.setId(id);

        AiChatSession session = new AiChatSession();
        session.setId(sessionId);
        session.setTitle("セッション1");
        entity.setSession(session);

        User user = new User();
        user.setId(userId);
        entity.setUser(user);

        entity.setAxisName(axisName);
        entity.setScore(score);
        entity.setComment(comment);
        entity.setCreatedAt(createdAt);
        return entity;
    }
}
