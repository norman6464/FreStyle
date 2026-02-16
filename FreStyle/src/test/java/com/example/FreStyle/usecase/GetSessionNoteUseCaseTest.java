package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.*;

import java.time.LocalDateTime;
import java.util.Optional;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.SessionNoteDto;
import com.example.FreStyle.entity.SessionNote;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.SessionNoteRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetSessionNoteUseCase")
class GetSessionNoteUseCaseTest {

    @Mock
    private SessionNoteRepository sessionNoteRepository;

    @InjectMocks
    private GetSessionNoteUseCase useCase;

    @Test
    @DisplayName("セッションIDでメモを取得する")
    void returnsNoteForSession() {
        User user = new User();
        user.setId(1);
        SessionNote note = new SessionNote();
        note.setId(10);
        note.setUser(user);
        note.setSessionId(100);
        note.setNote("振り返りメモ");
        note.setUpdatedAt(LocalDateTime.of(2026, 2, 16, 10, 0));
        when(sessionNoteRepository.findByUserIdAndSessionId(1, 100))
                .thenReturn(Optional.of(note));

        SessionNoteDto result = useCase.execute(1, 100);

        assertThat(result).isNotNull();
        assertThat(result.getSessionId()).isEqualTo(100);
        assertThat(result.getNote()).isEqualTo("振り返りメモ");
    }

    @Test
    @DisplayName("メモが存在しない場合nullを返す")
    void returnsNullWhenNotFound() {
        when(sessionNoteRepository.findByUserIdAndSessionId(1, 999))
                .thenReturn(Optional.empty());

        SessionNoteDto result = useCase.execute(1, 999);

        assertThat(result).isNull();
    }
}
