package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.when;

import java.util.List;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.ScoreCardDto;
import com.example.FreStyle.entity.CommunicationScore;
import com.example.FreStyle.repository.CommunicationScoreRepository;
import com.example.FreStyle.service.ScoreCardService;

/**
 * SaveScoreCardUseCaseのテストクラス
 *
 * <p>テスト対象:</p>
 * <ul>
 *   <li>AI応答からスコアを抽出してDBに保存</li>
 *   <li>保存成功時にScoreCardDtoを返却</li>
 *   <li>スコアが抽出できない場合のハンドリング</li>
 * </ul>
 */
@ExtendWith(MockitoExtension.class)
class SaveScoreCardUseCaseTest {

    @Mock
    private CommunicationScoreRepository communicationScoreRepository;

    @Mock
    private ScoreCardService scoreCardService;

    @InjectMocks
    private SaveScoreCardUseCase useCase;

    @Nested
    @DisplayName("execute - AI応答からスコアを保存")
    class ExecuteTest {

        @Test
        @DisplayName("正常なAI応答からスコアを抽出・保存し、ScoreCardDtoを返す")
        void shouldParseAndSaveScoresFromAiResponse() {
            // Arrange
            Integer sessionId = 1;
            Integer userId = 10;
            String scene = "meeting";
            String aiResponse = "フィードバック...\n```json\n{\"scores\":[{\"axis\":\"論理的構成力\",\"score\":8,\"comment\":\"良い\"}]}\n```";

            List<ScoreCardService.AxisScore> parsedScores = List.of(
                    new ScoreCardService.AxisScore("論理的構成力", 8, "良い")
            );

            when(scoreCardService.parseScoresFromResponse(aiResponse)).thenReturn(parsedScores);
            when(scoreCardService.calculateOverallScore(parsedScores)).thenReturn(8.0);
            when(communicationScoreRepository.save(any(CommunicationScore.class)))
                    .thenAnswer(invocation -> invocation.getArgument(0));

            // Act
            ScoreCardDto result = useCase.execute(sessionId, userId, aiResponse, scene);

            // Assert
            assertThat(result).isNotNull();
            assertThat(result.sessionId()).isEqualTo(sessionId);
            assertThat(result.scores()).hasSize(1);
            assertThat(result.scores().get(0).axis()).isEqualTo("論理的構成力");
            assertThat(result.scores().get(0).score()).isEqualTo(8);
            assertThat(result.overallScore()).isEqualTo(8.0);

            verify(communicationScoreRepository, times(1)).save(any(CommunicationScore.class));
        }

        @Test
        @DisplayName("複数軸のスコアを全て保存する")
        void shouldSaveMultipleAxisScores() {
            // Arrange
            Integer sessionId = 1;
            Integer userId = 10;
            String aiResponse = "```json\n{\"scores\":[" +
                    "{\"axis\":\"論理的構成力\",\"score\":8,\"comment\":\"良い\"}," +
                    "{\"axis\":\"配慮表現\",\"score\":6,\"comment\":\"改善余地\"}," +
                    "{\"axis\":\"要約力\",\"score\":7,\"comment\":\"まあまあ\"}" +
                    "]}\n```";

            List<ScoreCardService.AxisScore> parsedScores = List.of(
                    new ScoreCardService.AxisScore("論理的構成力", 8, "良い"),
                    new ScoreCardService.AxisScore("配慮表現", 6, "改善余地"),
                    new ScoreCardService.AxisScore("要約力", 7, "まあまあ")
            );

            when(scoreCardService.parseScoresFromResponse(aiResponse)).thenReturn(parsedScores);
            when(scoreCardService.calculateOverallScore(parsedScores)).thenReturn(7.0);
            when(communicationScoreRepository.save(any(CommunicationScore.class)))
                    .thenAnswer(invocation -> invocation.getArgument(0));

            // Act
            ScoreCardDto result = useCase.execute(sessionId, userId, aiResponse, null);

            // Assert
            assertThat(result.scores()).hasSize(3);
            assertThat(result.overallScore()).isEqualTo(7.0);

            ArgumentCaptor<CommunicationScore> captor = ArgumentCaptor.forClass(CommunicationScore.class);
            verify(communicationScoreRepository, times(3)).save(captor.capture());

            List<CommunicationScore> savedEntities = captor.getAllValues();
            assertThat(savedEntities.get(0).getAxisName()).isEqualTo("論理的構成力");
            assertThat(savedEntities.get(1).getAxisName()).isEqualTo("配慮表現");
            assertThat(savedEntities.get(2).getAxisName()).isEqualTo("要約力");
        }

        @Test
        @DisplayName("スコアが抽出できない場合はnullを返す")
        void shouldReturnNullWhenNoScoresExtracted() {
            // Arrange
            Integer sessionId = 1;
            Integer userId = 10;
            String aiResponse = "フィードバックのみで、スコアJSON無し";

            when(scoreCardService.parseScoresFromResponse(aiResponse)).thenReturn(List.of());

            // Act
            ScoreCardDto result = useCase.execute(sessionId, userId, aiResponse, null);

            // Assert
            assertThat(result).isNull();
            verify(communicationScoreRepository, never()).save(any());
        }

        @Test
        @DisplayName("保存時にsceneが正しくエンティティに設定される")
        void shouldSetSceneOnSavedEntities() {
            // Arrange
            Integer sessionId = 1;
            Integer userId = 10;
            String scene = "presentation";
            String aiResponse = "```json\n{\"scores\":[{\"axis\":\"要約力\",\"score\":7,\"comment\":\"OK\"}]}\n```";

            List<ScoreCardService.AxisScore> parsedScores = List.of(
                    new ScoreCardService.AxisScore("要約力", 7, "OK")
            );

            when(scoreCardService.parseScoresFromResponse(aiResponse)).thenReturn(parsedScores);
            when(scoreCardService.calculateOverallScore(parsedScores)).thenReturn(7.0);
            when(communicationScoreRepository.save(any(CommunicationScore.class)))
                    .thenAnswer(invocation -> invocation.getArgument(0));

            // Act
            useCase.execute(sessionId, userId, aiResponse, scene);

            // Assert
            ArgumentCaptor<CommunicationScore> captor = ArgumentCaptor.forClass(CommunicationScore.class);
            verify(communicationScoreRepository, times(1)).save(captor.capture());

            CommunicationScore saved = captor.getValue();
            assertThat(saved.getScene()).isEqualTo("presentation");
            assertThat(saved.getSession().getId()).isEqualTo(sessionId);
            assertThat(saved.getUser().getId()).isEqualTo(userId);
        }
    }
}
