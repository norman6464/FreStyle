package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.Optional;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.SharedSession;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.repository.SharedSessionRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("UnshareSessionUseCase")
class UnshareSessionUseCaseTest {

    @Mock
    private SharedSessionRepository sharedSessionRepository;

    @InjectMocks
    private UnshareSessionUseCase useCase;

    @Test
    @DisplayName("共有を解除する")
    void shouldUnshareSession() {
        User user = new User();
        user.setId(1);

        AiChatSession session = new AiChatSession();
        session.setId(10);
        session.setUser(user);

        SharedSession shared = new SharedSession();
        shared.setId(100);
        shared.setSession(session);
        shared.setUser(user);

        when(sharedSessionRepository.findBySessionId(10)).thenReturn(Optional.of(shared));

        useCase.execute(1, 10);

        verify(sharedSessionRepository).delete(shared);
    }

    @Test
    @DisplayName("他人のセッションの共有は解除できない")
    void shouldThrowWhenUnsharingOthersSession() {
        User otherUser = new User();
        otherUser.setId(2);

        AiChatSession session = new AiChatSession();
        session.setId(10);
        session.setUser(otherUser);

        SharedSession shared = new SharedSession();
        shared.setId(100);
        shared.setSession(session);
        shared.setUser(otherUser);

        when(sharedSessionRepository.findBySessionId(10)).thenReturn(Optional.of(shared));

        assertThatThrownBy(() -> useCase.execute(1, 10))
                .isInstanceOf(ResourceNotFoundException.class);
    }

    @Test
    @DisplayName("共有セッションが見つからない場合は例外を投げる")
    void shouldThrowWhenSharedSessionNotFound() {
        when(sharedSessionRepository.findBySessionId(999)).thenReturn(Optional.empty());

        assertThatThrownBy(() -> useCase.execute(1, 999))
                .isInstanceOf(ResourceNotFoundException.class);
    }
}
