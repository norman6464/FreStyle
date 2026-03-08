package com.example.FreStyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.List;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.ResponseEntity;
import org.springframework.security.oauth2.jwt.Jwt;

import com.example.FreStyle.dto.ConversationTemplateDto;
import com.example.FreStyle.entity.ConversationTemplate;
import com.example.FreStyle.usecase.GetConversationTemplatesUseCase;
import com.example.FreStyle.usecase.GetConversationTemplateByIdUseCase;

@ExtendWith(MockitoExtension.class)
@DisplayName("ConversationTemplateController")
class ConversationTemplateControllerTest {

    @Mock
    private GetConversationTemplatesUseCase getTemplatesUseCase;

    @Mock
    private GetConversationTemplateByIdUseCase getTemplateByIdUseCase;

    @InjectMocks
    private ConversationTemplateController controller;

    private Jwt jwt;

    @BeforeEach
    void setUp() {
        jwt = mock(Jwt.class);
    }

    @Nested
    @DisplayName("GET /api/templates")
    class GetTemplates {

        @Test
        @DisplayName("テンプレート一覧を取得する")
        void shouldReturnTemplateList() {
            // Arrange
            List<ConversationTemplateDto> templates = List.of(
                    new ConversationTemplateDto(1, "会議の進行", "会議練習", "meeting", "定例会議を始めます", "beginner"),
                    new ConversationTemplateDto(2, "プレゼン質疑応答", "質疑応答対応", "presentation", "質問があります", "intermediate")
            );
            when(getTemplatesUseCase.execute(null)).thenReturn(templates);

            // Act
            ResponseEntity<List<ConversationTemplateDto>> response = controller.getTemplates(jwt, null);

            // Assert
            assertThat(response.getStatusCode().value()).isEqualTo(200);
            assertThat(response.getBody()).hasSize(2);
            verify(getTemplatesUseCase).execute(null);
        }

        @Test
        @DisplayName("カテゴリでフィルタリングする")
        void shouldFilterByCategory() {
            // Arrange
            List<ConversationTemplateDto> templates = List.of(
                    new ConversationTemplateDto(1, "会議の進行", "会議練習", "meeting", "定例会議を始めます", "beginner")
            );
            when(getTemplatesUseCase.execute("meeting")).thenReturn(templates);

            // Act
            ResponseEntity<List<ConversationTemplateDto>> response = controller.getTemplates(jwt, "meeting");

            // Assert
            assertThat(response.getStatusCode().value()).isEqualTo(200);
            assertThat(response.getBody()).hasSize(1);
            assertThat(response.getBody().get(0).category()).isEqualTo("meeting");
            verify(getTemplatesUseCase).execute("meeting");
        }
    }

    @Nested
    @DisplayName("GET /api/templates/{id}")
    class GetTemplateById {

        @Test
        @DisplayName("IDでテンプレートを取得する")
        void shouldReturnTemplateById() {
            // Arrange
            ConversationTemplate template = new ConversationTemplate();
            template.setId(1);
            template.setTitle("会議の進行");
            template.setDescription("会議練習");
            template.setCategory("meeting");
            template.setOpeningMessage("定例会議を始めます");
            template.setDifficulty("beginner");

            when(getTemplateByIdUseCase.execute(1)).thenReturn(template);

            // Act
            ResponseEntity<ConversationTemplateDto> response = controller.getTemplateById(jwt, 1);

            // Assert
            assertThat(response.getStatusCode().value()).isEqualTo(200);
            assertThat(response.getBody()).isNotNull();
            assertThat(response.getBody().id()).isEqualTo(1);
            assertThat(response.getBody().title()).isEqualTo("会議の進行");
            verify(getTemplateByIdUseCase).execute(1);
        }
    }
}
