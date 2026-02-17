package com.example.FreStyle.mapper;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;

import com.example.FreStyle.dto.PracticeScenarioDto;
import com.example.FreStyle.entity.PracticeScenario;

/**
 * PracticeScenarioMapperのテストクラス
 *
 * <p>テスト対象:</p>
 * <ul>
 *   <li>エンティティ → DTOの変換（toDto）</li>
 *   <li>DTO → エンティティの変換（toEntity）</li>
 *   <li>null値のハンドリング</li>
 * </ul>
 */
class PracticeScenarioMapperTest {

    private final PracticeScenarioMapper mapper = new PracticeScenarioMapper();

    @Nested
    @DisplayName("toDto - エンティティからDTOへの変換")
    class ToDtoTest {

        @Test
        @DisplayName("全フィールドが正しくDTOに変換される")
        void shouldConvertAllFieldsToDto() {
            // Arrange
            PracticeScenario entity = new PracticeScenario();
            entity.setId(1);
            entity.setName("本番障害の緊急報告");
            entity.setDescription("本番環境で重大な障害が発生");
            entity.setCategory("customer");
            entity.setRoleName("怒っている顧客（SIer企業のPM）");
            entity.setDifficulty("intermediate");
            entity.setSystemPrompt("あなたは怒っている顧客です");

            // Act
            PracticeScenarioDto dto = mapper.toDto(entity);

            // Assert
            assertThat(dto.id()).isEqualTo(1);
            assertThat(dto.name()).isEqualTo("本番障害の緊急報告");
            assertThat(dto.description()).isEqualTo("本番環境で重大な障害が発生");
            assertThat(dto.category()).isEqualTo("customer");
            assertThat(dto.roleName()).isEqualTo("怒っている顧客（SIer企業のPM）");
            assertThat(dto.difficulty()).isEqualTo("intermediate");
            assertThat(dto.systemPrompt()).isEqualTo("あなたは怒っている顧客です");
        }

        @Test
        @DisplayName("nullエンティティを渡すとNullPointerExceptionがスローされる")
        void shouldThrowExceptionWhenEntityIsNull() {
            // Act & Assert
            assertThatThrownBy(() -> mapper.toDto(null))
                    .isInstanceOf(NullPointerException.class)
                    .hasMessage("PracticeScenarioエンティティがnullです");
        }
    }

    @Nested
    @DisplayName("toEntity - DTOからエンティティへの変換")
    class ToEntityTest {

        @Test
        @DisplayName("全フィールドが正しくエンティティに変換される")
        void shouldConvertAllFieldsToEntity() {
            // Arrange
            PracticeScenarioDto dto = new PracticeScenarioDto(
                    2,
                    "コードレビュー指摘",
                    "新人が書いたコードをレビュー",
                    "team",
                    "先輩エンジニア",
                    "beginner",
                    "あなたは先輩エンジニアです"
            );

            // Act
            PracticeScenario entity = mapper.toEntity(dto);

            // Assert
            assertThat(entity.getId()).isEqualTo(2);
            assertThat(entity.getName()).isEqualTo("コードレビュー指摘");
            assertThat(entity.getDescription()).isEqualTo("新人が書いたコードをレビュー");
            assertThat(entity.getCategory()).isEqualTo("team");
            assertThat(entity.getRoleName()).isEqualTo("先輩エンジニア");
            assertThat(entity.getDifficulty()).isEqualTo("beginner");
            assertThat(entity.getSystemPrompt()).isEqualTo("あなたは先輩エンジニアです");
        }

        @Test
        @DisplayName("nullDTOを渡すとNullPointerExceptionがスローされる")
        void shouldThrowExceptionWhenDtoIsNull() {
            // Act & Assert
            assertThatThrownBy(() -> mapper.toEntity(null))
                    .isInstanceOf(NullPointerException.class)
                    .hasMessage("PracticeScenarioDTOがnullです");
        }
    }

    @Nested
    @DisplayName("双方向変換の整合性")
    class BidirectionalConversionTest {

        @Test
        @DisplayName("エンティティ → DTO → エンティティの変換で情報が保持される")
        void shouldPreserveDataInBidirectionalConversion() {
            // Arrange
            PracticeScenario originalEntity = new PracticeScenario();
            originalEntity.setId(3);
            originalEntity.setName("日報作成");
            originalEntity.setDescription("今日の作業内容を報告");
            originalEntity.setCategory("daily");
            originalEntity.setRoleName("上司");
            originalEntity.setDifficulty("beginner");
            originalEntity.setSystemPrompt("あなたは上司です");

            // Act
            PracticeScenarioDto dto = mapper.toDto(originalEntity);
            PracticeScenario convertedEntity = mapper.toEntity(dto);

            // Assert
            assertThat(convertedEntity.getId()).isEqualTo(originalEntity.getId());
            assertThat(convertedEntity.getName()).isEqualTo(originalEntity.getName());
            assertThat(convertedEntity.getDescription()).isEqualTo(originalEntity.getDescription());
            assertThat(convertedEntity.getCategory()).isEqualTo(originalEntity.getCategory());
            assertThat(convertedEntity.getRoleName()).isEqualTo(originalEntity.getRoleName());
            assertThat(convertedEntity.getDifficulty()).isEqualTo(originalEntity.getDifficulty());
            assertThat(convertedEntity.getSystemPrompt()).isEqualTo(originalEntity.getSystemPrompt());
        }
    }
}
