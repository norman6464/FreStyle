package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyInt;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.ArgumentMatchers.isNull;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.AiChatSessionDto;
import com.example.FreStyle.dto.PracticeScenarioDto;
import com.example.FreStyle.entity.User;

/**
 * CreatePracticeSessionUseCaseのテストクラス
 *
 * <p>テスト対象:</p>
 * <ul>
 *   <li>練習セッションの作成</li>
 *   <li>セッションタイトルの自動生成</li>
 *   <li>シナリオIDが無効な場合のエラーハンドリング</li>
 * </ul>
 */
@ExtendWith(MockitoExtension.class)
class CreatePracticeSessionUseCaseTest {

    @Mock
    private GetPracticeScenarioByIdUseCase getPracticeScenarioByIdUseCase;

    @Mock
    private CreateAiChatSessionUseCase createAiChatSessionUseCase;

    @InjectMocks
    private CreatePracticeSessionUseCase useCase;

    @Nested
    @DisplayName("execute - 練習セッション作成")
    class ExecuteTest {

        @Test
        @DisplayName("有効なシナリオIDで練習セッションを作成する")
        void shouldCreatePracticeSessionWithValidScenarioId() {
            // Arrange
            User user = createUser(1, "テストユーザー");
            Integer scenarioId = 1;
            PracticeScenarioDto scenario = createScenarioDto(
                    scenarioId,
                    "本番障害の緊急報告",
                    "customer"
            );
            AiChatSessionDto expectedSession = createSessionDto(
                    100,
                    "練習: 本番障害の緊急報告",
                    "practice",
                    scenarioId
            );

            when(getPracticeScenarioByIdUseCase.execute(scenarioId))
                    .thenReturn(scenario);
            when(createAiChatSessionUseCase.execute(
                    eq(1),
                    eq("練習: 本番障害の緊急報告"),
                    isNull(),
                    isNull(),
                    eq("practice"),
                    eq(scenarioId)
            )).thenReturn(expectedSession);

            // Act
            AiChatSessionDto result = useCase.execute(user, scenarioId);

            // Assert
            assertThat(result).isNotNull();
            assertThat(result.id()).isEqualTo(100);
            assertThat(result.title()).isEqualTo("練習: 本番障害の緊急報告");
            assertThat(result.sessionType()).isEqualTo("practice");
            assertThat(result.scenarioId()).isEqualTo(scenarioId);

            verify(getPracticeScenarioByIdUseCase, times(1)).execute(scenarioId);
            verify(createAiChatSessionUseCase, times(1)).execute(
                    eq(1),
                    eq("練習: 本番障害の緊急報告"),
                    isNull(),
                    isNull(),
                    eq("practice"),
                    eq(scenarioId)
            );
        }

        @Test
        @DisplayName("セッションタイトルが「練習: {シナリオ名}」の形式で生成される")
        void shouldGenerateCorrectSessionTitle() {
            // Arrange
            User user = createUser(2, "山田太郎");
            Integer scenarioId = 3;
            PracticeScenarioDto scenario = createScenarioDto(
                    scenarioId,
                    "日報作成",
                    "daily"
            );
            AiChatSessionDto session = createSessionDto(
                    101,
                    "練習: 日報作成",
                    "practice",
                    scenarioId
            );

            when(getPracticeScenarioByIdUseCase.execute(scenarioId))
                    .thenReturn(scenario);
            when(createAiChatSessionUseCase.execute(
                    anyInt(),
                    eq("練習: 日報作成"),
                    isNull(),
                    isNull(),
                    eq("practice"),
                    eq(scenarioId)
            )).thenReturn(session);

            // Act
            AiChatSessionDto result = useCase.execute(user, scenarioId);

            // Assert
            assertThat(result.title()).isEqualTo("練習: 日報作成");
            verify(createAiChatSessionUseCase, times(1)).execute(
                    eq(2),
                    eq("練習: 日報作成"),
                    isNull(),
                    isNull(),
                    eq("practice"),
                    eq(scenarioId)
            );
        }

        @Test
        @DisplayName("存在しないシナリオIDの場合はRuntimeExceptionがスローされる")
        void shouldThrowExceptionWhenScenarioNotFound() {
            // Arrange
            User user = createUser(1, "テストユーザー");
            Integer invalidScenarioId = 999;

            when(getPracticeScenarioByIdUseCase.execute(invalidScenarioId))
                    .thenThrow(new RuntimeException("シナリオが見つかりません: ID=" + invalidScenarioId));

            // Act & Assert
            assertThatThrownBy(() -> useCase.execute(user, invalidScenarioId))
                    .isInstanceOf(RuntimeException.class)
                    .hasMessage("シナリオが見つかりません: ID=" + invalidScenarioId);

            verify(getPracticeScenarioByIdUseCase, times(1)).execute(invalidScenarioId);
            verify(createAiChatSessionUseCase, times(0)).execute(
                    anyInt(), any(), any(), any(), any(), anyInt()
            );
        }

        @Test
        @DisplayName("セッション作成時にrelatedRoomIdとsceneはnullで渡される")
        void shouldPassNullForRelatedRoomIdAndScene() {
            // Arrange
            User user = createUser(1, "テストユーザー");
            Integer scenarioId = 2;
            PracticeScenarioDto scenario = createScenarioDto(
                    scenarioId,
                    "コードレビュー指摘",
                    "team"
            );
            AiChatSessionDto session = createSessionDto(
                    102,
                    "練習: コードレビュー指摘",
                    "practice",
                    scenarioId
            );

            when(getPracticeScenarioByIdUseCase.execute(scenarioId))
                    .thenReturn(scenario);
            when(createAiChatSessionUseCase.execute(
                    anyInt(),
                    any(),
                    isNull(),  // relatedRoomId = null
                    isNull(),  // scene = null
                    any(),
                    anyInt()
            )).thenReturn(session);

            // Act
            useCase.execute(user, scenarioId);

            // Assert
            verify(createAiChatSessionUseCase, times(1)).execute(
                    eq(1),
                    any(),
                    isNull(),  // relatedRoomId
                    isNull(),  // scene
                    eq("practice"),
                    eq(scenarioId)
            );
        }
    }

    // ヘルパーメソッド
    private User createUser(Integer id, String name) {
        User user = new User();
        user.setId(id);
        user.setName(name);
        return user;
    }

    private PracticeScenarioDto createScenarioDto(
            Integer id,
            String name,
            String category
    ) {
        return new PracticeScenarioDto(
                id,
                name,
                "説明: " + name,
                category,
                "役割",
                "intermediate",
                "システムプロンプト"
        );
    }

    private AiChatSessionDto createSessionDto(
            Integer id,
            String title,
            String sessionType,
            Integer scenarioId
    ) {
        return new AiChatSessionDto(id, null, title, null, null, sessionType, scenarioId, null, null);
    }
}
