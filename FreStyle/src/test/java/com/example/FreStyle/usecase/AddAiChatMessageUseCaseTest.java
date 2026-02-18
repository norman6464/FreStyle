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

import com.example.FreStyle.dto.AiChatMessageResponseDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.repository.AiChatMessageDynamoRepository;
import com.example.FreStyle.repository.AiChatSessionRepository;
import com.example.FreStyle.repository.UserRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("AddAiChatMessageUseCase")
class AddAiChatMessageUseCaseTest {

    @Mock
    private AiChatMessageDynamoRepository aiChatMessageDynamoRepository;

    @Mock
    private AiChatSessionRepository aiChatSessionRepository;

    @Mock
    private UserRepository userRepository;

    @InjectMocks
    private AddAiChatMessageUseCase useCase;

    @Nested
    @DisplayName("正常系")
    class Success {

        @Test
        @DisplayName("ユーザーメッセージを追加してDTOを返す")
        void shouldAddUserMessage() {
            AiChatSession session = new AiChatSession();
            session.setId(1);
            when(aiChatSessionRepository.findById(1)).thenReturn(Optional.of(session));

            User user = new User();
            user.setId(1);
            when(userRepository.findById(1)).thenReturn(Optional.of(user));

            AiChatMessageResponseDto expectedDto = new AiChatMessageResponseDto("msg-1", 1, 1, "user", "こんにちは", 1000L);
            when(aiChatMessageDynamoRepository.save(1, 1, "user", "こんにちは")).thenReturn(expectedDto);

            AiChatMessageResponseDto result = useCase.executeUserMessage(1, 1, "こんにちは");

            assertThat(result.id()).isEqualTo("msg-1");
            assertThat(result.role()).isEqualTo("user");
            verify(aiChatMessageDynamoRepository).save(1, 1, "user", "こんにちは");
        }

        @Test
        @DisplayName("アシスタントメッセージを追加してDTOを返す")
        void shouldAddAssistantMessage() {
            AiChatSession session = new AiChatSession();
            session.setId(1);
            when(aiChatSessionRepository.findById(1)).thenReturn(Optional.of(session));

            User user = new User();
            user.setId(1);
            when(userRepository.findById(1)).thenReturn(Optional.of(user));

            AiChatMessageResponseDto expectedDto = new AiChatMessageResponseDto("msg-2", 1, 1, "assistant", "応答です", 2000L);
            when(aiChatMessageDynamoRepository.save(1, 1, "assistant", "応答です")).thenReturn(expectedDto);

            AiChatMessageResponseDto result = useCase.executeAssistantMessage(1, 1, "応答です");

            assertThat(result.id()).isEqualTo("msg-2");
            assertThat(result.role()).isEqualTo("assistant");
        }
    }

    @Nested
    @DisplayName("異常系")
    class Error {

        @Test
        @DisplayName("セッションが見つからない場合はResourceNotFoundExceptionをスローする")
        void shouldThrowWhenSessionNotFound() {
            when(aiChatSessionRepository.findById(999)).thenReturn(Optional.empty());

            assertThatThrownBy(() -> useCase.executeUserMessage(999, 1, "テスト"))
                    .isInstanceOf(ResourceNotFoundException.class);
        }

        @Test
        @DisplayName("ユーザーが見つからない場合はResourceNotFoundExceptionをスローする")
        void shouldThrowWhenUserNotFound() {
            AiChatSession session = new AiChatSession();
            session.setId(1);
            when(aiChatSessionRepository.findById(1)).thenReturn(Optional.of(session));
            when(userRepository.findById(999)).thenReturn(Optional.empty());

            assertThatThrownBy(() -> useCase.executeUserMessage(1, 999, "テスト"))
                    .isInstanceOf(ResourceNotFoundException.class);
        }
    }
}
