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
class UpdateNoteUseCaseTest {

    @Mock
    private NoteRepository noteRepository;

    @InjectMocks
    private UpdateNoteUseCase useCase;

    @Nested
    @DisplayName("execute - ノート更新")
    class ExecuteTest {

        @Test
        @DisplayName("ノートのタイトル・内容・ピン留めを更新する")
        void shouldUpdateNote() {
            doNothing().when(noteRepository).update(1, "note-1", "更新タイトル", "更新内容", true);

            useCase.execute(1, "note-1", "更新タイトル", "更新内容", true);

            verify(noteRepository, times(1)).update(1, "note-1", "更新タイトル", "更新内容", true);
        }

        @Test
        @DisplayName("ピン留め状態のみを変更する")
        void shouldUpdateOnlyPinnedStatus() {
            doNothing().when(noteRepository).update(1, "note-1", "同じタイトル", "同じ内容", true);

            useCase.execute(1, "note-1", "同じタイトル", "同じ内容", true);

            verify(noteRepository, times(1)).update(1, "note-1", "同じタイトル", "同じ内容", true);
        }

        @Test
        @DisplayName("空文字のタイトルと内容で更新する")
        void shouldUpdateWithEmptyValues() {
            doNothing().when(noteRepository).update(1, "note-1", "", "", false);

            useCase.execute(1, "note-1", "", "", false);

            verify(noteRepository, times(1)).update(1, "note-1", "", "", false);
        }
    }
}
