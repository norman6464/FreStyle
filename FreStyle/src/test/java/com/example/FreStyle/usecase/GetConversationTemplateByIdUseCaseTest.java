package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
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

import com.example.FreStyle.entity.ConversationTemplate;
import com.example.FreStyle.repository.ConversationTemplateRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetConversationTemplateByIdUseCase")
class GetConversationTemplateByIdUseCaseTest {

    @Mock
    private ConversationTemplateRepository repository;

    @InjectMocks
    private GetConversationTemplateByIdUseCase useCase;

    @Nested
    @DisplayName("execute - IDでテンプレート取得")
    class ExecuteTest {

        @Test
        @DisplayName("IDでテンプレートを取得する")
        void shouldReturnTemplateById() {
            // Arrange
            ConversationTemplate template = new ConversationTemplate();
            template.setId(1);
            template.setTitle("会議の進行");
            template.setCategory("meeting");
            template.setOpeningMessage("定例会議を始めます");
            template.setSystemContext("チームメンバーとして参加");
            template.setDifficulty("beginner");

            when(repository.findById(1)).thenReturn(Optional.of(template));

            // Act
            ConversationTemplate result = useCase.execute(1);

            // Assert
            assertThat(result.getId()).isEqualTo(1);
            assertThat(result.getTitle()).isEqualTo("会議の進行");
            verify(repository).findById(1);
        }

        @Test
        @DisplayName("存在しないIDで例外を投げる")
        void shouldThrowExceptionWhenNotFound() {
            // Arrange
            when(repository.findById(999)).thenReturn(Optional.empty());

            // Act & Assert
            assertThatThrownBy(() -> useCase.execute(999))
                    .isInstanceOf(IllegalArgumentException.class)
                    .hasMessageContaining("テンプレートが見つかりません: 999");

            verify(repository).findById(999);
        }
    }
}
