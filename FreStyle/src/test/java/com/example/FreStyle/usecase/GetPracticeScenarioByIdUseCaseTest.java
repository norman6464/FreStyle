package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.Optional;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.PracticeScenarioDto;
import com.example.FreStyle.entity.PracticeScenario;
import com.example.FreStyle.mapper.PracticeScenarioMapper;
import com.example.FreStyle.repository.PracticeScenarioRepository;

/**
 * GetPracticeScenarioByIdUseCaseのテストクラス
 *
 * <p>テスト対象:</p>
 * <ul>
 *   <li>ID指定でのシナリオ取得</li>
 *   <li>存在しないIDのエラーハンドリング</li>
 *   <li>エンティティからDTOへの変換</li>
 * </ul>
 */
@ExtendWith(MockitoExtension.class)
class GetPracticeScenarioByIdUseCaseTest {

    @Mock
    private PracticeScenarioRepository practiceScenarioRepository;

    @Mock
    private PracticeScenarioMapper mapper;

    @InjectMocks
    private GetPracticeScenarioByIdUseCase useCase;

    @Nested
    @DisplayName("execute - ID指定シナリオ取得")
    class ExecuteTest {

        @Test
        @DisplayName("指定IDのシナリオを取得してDTOに変換する")
        void shouldReturnScenarioDtoById() {
            // Arrange
            Integer scenarioId = 1;
            PracticeScenario entity = createEntity(
                    scenarioId,
                    "本番障害の緊急報告",
                    "customer",
                    "怒っている顧客（SIer企業のPM）",
                    "intermediate"
            );
            PracticeScenarioDto dto = createDto(
                    scenarioId,
                    "本番障害の緊急報告",
                    "customer",
                    "怒っている顧客（SIer企業のPM）",
                    "intermediate"
            );

            when(practiceScenarioRepository.findById(scenarioId))
                    .thenReturn(Optional.of(entity));
            when(mapper.toDto(entity)).thenReturn(dto);

            // Act
            PracticeScenarioDto result = useCase.execute(scenarioId);

            // Assert
            assertThat(result).isNotNull();
            assertThat(result.getId()).isEqualTo(scenarioId);
            assertThat(result.getName()).isEqualTo("本番障害の緊急報告");
            assertThat(result.getCategory()).isEqualTo("customer");
            assertThat(result.getRoleName()).isEqualTo("怒っている顧客（SIer企業のPM）");
            assertThat(result.getDifficulty()).isEqualTo("intermediate");

            verify(practiceScenarioRepository, times(1)).findById(scenarioId);
            verify(mapper, times(1)).toDto(entity);
        }

        @Test
        @DisplayName("存在しないIDの場合はRuntimeExceptionがスローされる")
        void shouldThrowExceptionWhenScenarioNotFound() {
            // Arrange
            Integer nonExistentId = 999;
            when(practiceScenarioRepository.findById(nonExistentId))
                    .thenReturn(Optional.empty());

            // Act & Assert
            assertThatThrownBy(() -> useCase.execute(nonExistentId))
                    .isInstanceOf(RuntimeException.class)
                    .hasMessage("シナリオが見つかりません: ID=" + nonExistentId);

            verify(practiceScenarioRepository, times(1)).findById(nonExistentId);
        }

        @Test
        @DisplayName("MapperがエンティティをDTOに正しく変換する")
        void shouldConvertEntityToDtoThroughMapper() {
            // Arrange
            Integer scenarioId = 2;
            PracticeScenario entity = createEntity(
                    scenarioId,
                    "コードレビュー指摘",
                    "team",
                    "先輩エンジニア",
                    "beginner"
            );
            PracticeScenarioDto expectedDto = createDto(
                    scenarioId,
                    "コードレビュー指摘",
                    "team",
                    "先輩エンジニア",
                    "beginner"
            );

            when(practiceScenarioRepository.findById(scenarioId))
                    .thenReturn(Optional.of(entity));
            when(mapper.toDto(entity)).thenReturn(expectedDto);

            // Act
            PracticeScenarioDto result = useCase.execute(scenarioId);

            // Assert
            assertThat(result).isEqualTo(expectedDto);
            verify(mapper, times(1)).toDto(entity);
        }
    }

    // ヘルパーメソッド
    private PracticeScenario createEntity(
            Integer id,
            String name,
            String category,
            String roleName,
            String difficulty
    ) {
        PracticeScenario entity = new PracticeScenario();
        entity.setId(id);
        entity.setName(name);
        entity.setDescription("説明: " + name);
        entity.setCategory(category);
        entity.setRoleName(roleName);
        entity.setDifficulty(difficulty);
        entity.setSystemPrompt("あなたは" + roleName + "です");
        return entity;
    }

    private PracticeScenarioDto createDto(
            Integer id,
            String name,
            String category,
            String roleName,
            String difficulty
    ) {
        return new PracticeScenarioDto(
                id,
                name,
                "説明: " + name,
                category,
                roleName,
                difficulty,
                "あなたは" + roleName + "です"
        );
    }
}
