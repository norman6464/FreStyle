package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.Optional;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InOrder;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.Mockito;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.repository.AiChatMessageDynamoRepository;
import com.example.FreStyle.repository.AiChatSessionRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("DeleteAiChatSessionUseCase")
class DeleteAiChatSessionUseCaseTest {

    @Mock
    private AiChatSessionRepository aiChatSessionRepository;

    @Mock
    private AiChatMessageDynamoRepository aiChatMessageDynamoRepository;

    @InjectMocks
    private DeleteAiChatSessionUseCase useCase;

    @Nested
    @DisplayName("正常系")
    class Success {

        @Test
        @DisplayName("DynamoDBメッセージ削除後にセッションを削除する")
        void shouldDeleteMessagesFirstThenSession() {
            User user = new User();
            user.setId(1);
            AiChatSession session = new AiChatSession();
            session.setId(10);
            session.setUser(user);
            when(aiChatSessionRepository.findByIdAndUserId(10, 1))
                    .thenReturn(Optional.of(session));

            useCase.execute(10, 1);

            // DynamoDBメッセージが先に削除され、その後セッションが削除される
            InOrder inOrder = Mockito.inOrder(aiChatMessageDynamoRepository, aiChatSessionRepository);
            inOrder.verify(aiChatMessageDynamoRepository).deleteBySessionId(10);
            inOrder.verify(aiChatSessionRepository).delete(session);
        }
    }

    @Nested
    @DisplayName("異常系")
    class Error {

        @Test
        @DisplayName("セッションが見つからない場合はResourceNotFoundExceptionをスローする")
        void shouldThrowWhenSessionNotFound() {
            when(aiChatSessionRepository.findByIdAndUserId(999, 1))
                    .thenReturn(Optional.empty());

            assertThatThrownBy(() -> useCase.execute(999, 1))
                    .isInstanceOf(ResourceNotFoundException.class)
                    .hasMessageContaining("999")
                    .hasMessageContaining("1");
        }

        @Test
        @DisplayName("他ユーザーのセッションは削除できない")
        void shouldThrowWhenDifferentUser() {
            when(aiChatSessionRepository.findByIdAndUserId(10, 2))
                    .thenReturn(Optional.empty());

            assertThatThrownBy(() -> useCase.execute(10, 2))
                    .isInstanceOf(ResourceNotFoundException.class);
        }
    }
}
