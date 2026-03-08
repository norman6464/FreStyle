package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
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

import com.example.FreStyle.dto.ConversationTemplateDto;
import com.example.FreStyle.entity.ConversationTemplate;
import com.example.FreStyle.repository.ConversationTemplateRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetConversationTemplatesUseCase")
class GetConversationTemplatesUseCaseTest {

    @Mock
    private ConversationTemplateRepository repository;

    @InjectMocks
    private GetConversationTemplatesUseCase useCase;

    @Nested
    @DisplayName("execute - テンプレート一覧取得")
    class ExecuteTest {

        @Test
        @DisplayName("全テンプレートを取得する")
        void shouldReturnAllTemplates() {
            // Arrange
            List<ConversationTemplate> templates = List.of(
                    createTemplate(1, "会議の進行", "会議練習", "meeting", "定例会議を始めます", "チームメンバーとして参加", "beginner"),
                    createTemplate(2, "プレゼン質疑応答", "質疑応答対応", "presentation", "質問があります", "質疑応答の場面", "intermediate")
            );
            when(repository.findAll()).thenReturn(templates);

            // Act
            List<ConversationTemplateDto> result = useCase.execute(null);

            // Assert
            assertThat(result).hasSize(2);
            verify(repository).findAll();
        }

        @Test
        @DisplayName("カテゴリ指定で取得する")
        void shouldReturnTemplatesByCategory() {
            // Arrange
            List<ConversationTemplate> templates = List.of(
                    createTemplate(1, "会議の進行", "会議練習", "meeting", "定例会議を始めます", "チームメンバーとして参加", "beginner")
            );
            when(repository.findByCategory("meeting")).thenReturn(templates);

            // Act
            List<ConversationTemplateDto> result = useCase.execute("meeting");

            // Assert
            assertThat(result).hasSize(1);
            assertThat(result.get(0).category()).isEqualTo("meeting");
            verify(repository).findByCategory("meeting");
        }

        @Test
        @DisplayName("DTOに正しく変換する")
        void shouldConvertToDto() {
            // Arrange
            List<ConversationTemplate> templates = List.of(
                    createTemplate(1, "会議の進行", "会議練習", "meeting", "定例会議を始めます", "チームメンバーとして参加", "beginner")
            );
            when(repository.findAll()).thenReturn(templates);

            // Act
            List<ConversationTemplateDto> result = useCase.execute(null);

            // Assert
            assertThat(result).hasSize(1);
            ConversationTemplateDto dto = result.get(0);
            assertThat(dto.id()).isEqualTo(1);
            assertThat(dto.title()).isEqualTo("会議の進行");
            assertThat(dto.description()).isEqualTo("会議練習");
            assertThat(dto.category()).isEqualTo("meeting");
            assertThat(dto.openingMessage()).isEqualTo("定例会議を始めます");
            assertThat(dto.difficulty()).isEqualTo("beginner");
        }

        @Test
        @DisplayName("空文字カテゴリは全件取得する")
        void shouldReturnAllTemplatesWhenCategoryIsEmpty() {
            // Arrange
            when(repository.findAll()).thenReturn(List.of());

            // Act
            List<ConversationTemplateDto> result = useCase.execute("");

            // Assert
            assertThat(result).isEmpty();
            verify(repository).findAll();
        }
    }

    private ConversationTemplate createTemplate(Integer id, String title, String description,
            String category, String openingMessage, String systemContext, String difficulty) {
        ConversationTemplate t = new ConversationTemplate();
        t.setId(id);
        t.setTitle(title);
        t.setDescription(description);
        t.setCategory(category);
        t.setOpeningMessage(openingMessage);
        t.setSystemContext(systemContext);
        t.setDifficulty(difficulty);
        return t;
    }
}
