package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.Arrays;
import java.util.List;

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
 * GetAllPracticeScenariosUseCaseのテストクラス
 *
 * <p>テスト対象:</p>
 * <ul>
 *   <li>全シナリオの取得</li>
 *   <li>エンティティからDTOへの変換</li>
 *   <li>空のリストのハンドリング</li>
 * </ul>
 */
@ExtendWith(MockitoExtension.class)
class GetAllPracticeScenariosUseCaseTest {

    @Mock
    private PracticeScenarioRepository practiceScenarioRepository;

    @Mock
    private PracticeScenarioMapper mapper;

    @InjectMocks
    private GetAllPracticeScenariosUseCase useCase;

    @Nested
    @DisplayName("execute - 全シナリオ取得")
    class ExecuteTest {

        @Test
        @DisplayName("全シナリオを取得してDTOリストに変換する")
        void shouldReturnAllScenariosAsDtoList() {
            // Arrange
            PracticeScenario entity1 = createEntity(1, "シナリオ1", "customer");
            PracticeScenario entity2 = createEntity(2, "シナリオ2", "team");
            PracticeScenario entity3 = createEntity(3, "シナリオ3", "management");

            PracticeScenarioDto dto1 = createDto(1, "シナリオ1", "customer");
            PracticeScenarioDto dto2 = createDto(2, "シナリオ2", "team");
            PracticeScenarioDto dto3 = createDto(3, "シナリオ3", "management");

            when(practiceScenarioRepository.findAll())
                    .thenReturn(Arrays.asList(entity1, entity2, entity3));
            when(mapper.toDto(entity1)).thenReturn(dto1);
            when(mapper.toDto(entity2)).thenReturn(dto2);
            when(mapper.toDto(entity3)).thenReturn(dto3);

            // Act
            List<PracticeScenarioDto> result = useCase.execute();

            // Assert
            assertThat(result).hasSize(3);
            assertThat(result.get(0).getName()).isEqualTo("シナリオ1");
            assertThat(result.get(1).getName()).isEqualTo("シナリオ2");
            assertThat(result.get(2).getName()).isEqualTo("シナリオ3");

            verify(practiceScenarioRepository, times(1)).findAll();
            verify(mapper, times(3)).toDto(org.mockito.ArgumentMatchers.any(PracticeScenario.class));
        }

        @Test
        @DisplayName("シナリオが0件の場合は空リストを返す")
        void shouldReturnEmptyListWhenNoScenarios() {
            // Arrange
            when(practiceScenarioRepository.findAll()).thenReturn(List.of());

            // Act
            List<PracticeScenarioDto> result = useCase.execute();

            // Assert
            assertThat(result).isEmpty();
            verify(practiceScenarioRepository, times(1)).findAll();
        }

        @Test
        @DisplayName("各シナリオがMapperを通じてDTOに変換される")
        void shouldConvertEachScenarioThroughMapper() {
            // Arrange
            PracticeScenario entity = createEntity(1, "テストシナリオ", "customer");
            PracticeScenarioDto dto = createDto(1, "テストシナリオ", "customer");

            when(practiceScenarioRepository.findAll()).thenReturn(List.of(entity));
            when(mapper.toDto(entity)).thenReturn(dto);

            // Act
            List<PracticeScenarioDto> result = useCase.execute();

            // Assert
            assertThat(result).hasSize(1);
            assertThat(result.get(0)).isEqualTo(dto);
            verify(mapper, times(1)).toDto(entity);
        }
    }

    // ヘルパーメソッド
    private PracticeScenario createEntity(Integer id, String name, String category) {
        PracticeScenario entity = new PracticeScenario();
        entity.setId(id);
        entity.setName(name);
        entity.setDescription("説明: " + name);
        entity.setCategory(category);
        entity.setRoleName("役割");
        entity.setDifficulty("intermediate");
        entity.setSystemPrompt("システムプロンプト");
        return entity;
    }

    private PracticeScenarioDto createDto(Integer id, String name, String category) {
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
}
