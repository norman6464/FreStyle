package com.example.FreStyle.usecase;

import com.example.FreStyle.repository.NoteRepository;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class DeleteNoteUseCaseTest {

    @Mock
    private NoteRepository noteRepository;

    @InjectMocks
    private DeleteNoteUseCase useCase;

    @Nested
    @DisplayName("execute - ノート削除")
    class ExecuteTest {

        @Test
        @DisplayName("指定されたノートを削除する")
        void shouldDeleteNote() {
            doNothing().when(noteRepository).delete(1, "note-1");

            useCase.execute(1, "note-1");

            verify(noteRepository, times(1)).delete(1, "note-1");
        }
    }
}
