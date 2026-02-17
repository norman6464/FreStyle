package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
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
import com.example.FreStyle.entity.AiChatMessage;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.mapper.AiChatMessageMapper;
import com.example.FreStyle.repository.AiChatMessageRepository;
import com.example.FreStyle.repository.AiChatSessionRepository;
import com.example.FreStyle.repository.UserRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("AddAiChatMessageUseCase")
class AddAiChatMessageUseCaseTest {

    @Mock
    private AiChatMessageRepository aiChatMessageRepository;

    @Mock
    private AiChatSessionRepository aiChatSessionRepository;

    @Mock
    private UserRepository userRepository;

    @Mock
    private AiChatMessageMapper mapper;

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

            AiChatMessage savedMessage = new AiChatMessage();
            savedMessage.setId(10);
            when(aiChatMessageRepository.save(any(AiChatMessage.class))).thenReturn(savedMessage);

            AiChatMessageResponseDto expectedDto = new AiChatMessageResponseDto(10, null, null, "user", null, null);
            when(mapper.toDto(savedMessage)).thenReturn(expectedDto);

            AiChatMessageResponseDto result = useCase.executeUserMessage(1, 1, "こんにちは");

            assertThat(result.id()).isEqualTo(10);
            assertThat(result.role()).isEqualTo("user");
            verify(aiChatMessageRepository).save(any(AiChatMessage.class));
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

            AiChatMessage savedMessage = new AiChatMessage();
            savedMessage.setId(11);
            when(aiChatMessageRepository.save(any(AiChatMessage.class))).thenReturn(savedMessage);

            AiChatMessageResponseDto expectedDto = new AiChatMessageResponseDto(11, null, null, "assistant", null, null);
            when(mapper.toDto(savedMessage)).thenReturn(expectedDto);

            AiChatMessageResponseDto result = useCase.executeAssistantMessage(1, 1, "応答です");

            assertThat(result.id()).isEqualTo(11);
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
