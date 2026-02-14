package com.example.FreStyle.usecase;

import com.example.FreStyle.dto.NoteDto;
import com.example.FreStyle.repository.NoteRepository;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class CreateNoteUseCaseTest {

    @Mock
    private NoteRepository noteRepository;

    @InjectMocks
    private CreateNoteUseCase useCase;

    @Nested
    @DisplayName("execute - ノート作成")
    class ExecuteTest {

        @Test
        @DisplayName("新しいノートを作成して返す")
        void shouldCreateNoteAndReturnDto() {
            NoteDto expected = new NoteDto("note-new", 1, "新しいノート", "", false, 1000L, 1000L);
            when(noteRepository.save(1, "新しいノート")).thenReturn(expected);

            NoteDto result = useCase.execute(1, "新しいノート");

            assertThat(result).isNotNull();
            assertThat(result.getNoteId()).isEqualTo("note-new");
            assertThat(result.getTitle()).isEqualTo("新しいノート");
            assertThat(result.getUserId()).isEqualTo(1);
            verify(noteRepository, times(1)).save(1, "新しいノート");
        }

        @Test
        @DisplayName("空文字タイトルでもノートを作成できる")
        void shouldCreateNoteWithEmptyTitle() {
            NoteDto expected = new NoteDto("note-empty", 1, "", "", false, 1000L, 1000L);
            when(noteRepository.save(1, "")).thenReturn(expected);

            NoteDto result = useCase.execute(1, "");

            assertThat(result.getTitle()).isEmpty();
            verify(noteRepository, times(1)).save(1, "");
        }
    }
}
