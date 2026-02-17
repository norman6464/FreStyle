package com.example.FreStyle.usecase;

import com.example.FreStyle.repository.NoteRepository;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import static org.assertj.core.api.Assertions.assertThatCode;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
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

        @Test
        @DisplayName("NoteRepositoryが例外をスローした場合そのまま伝搬する")
        void shouldPropagateRepositoryException() {
            doThrow(new RuntimeException("DynamoDB error"))
                    .when(noteRepository).delete(1, "note-1");

            assertThatThrownBy(() -> useCase.execute(1, "note-1"))
                    .isInstanceOf(RuntimeException.class)
                    .hasMessage("DynamoDB error");
        }

        @Test
        @DisplayName("異なるuserIdとnoteIdの組み合わせで正しく削除される")
        void shouldDeleteWithDifferentParams() {
            doNothing().when(noteRepository).delete(99, "note-xyz");

            assertThatCode(() -> useCase.execute(99, "note-xyz"))
                    .doesNotThrowAnyException();

            verify(noteRepository, times(1)).delete(99, "note-xyz");
        }

        @Test
        @DisplayName("deleteが複数回呼ばれても各回で正しいパラメータが渡される")
        void shouldPassCorrectParamsOnMultipleCalls() {
            doNothing().when(noteRepository).delete(anyInt(), anyString());

            useCase.execute(1, "note-a");
            useCase.execute(2, "note-b");

            verify(noteRepository).delete(1, "note-a");
            verify(noteRepository).delete(2, "note-b");
            verify(noteRepository, times(2)).delete(anyInt(), anyString());
        }

        @Test
        @DisplayName("IllegalArgumentExceptionがリポジトリから発生した場合そのまま伝搬する")
        void shouldPropagateIllegalArgumentException() {
            doThrow(new IllegalArgumentException("不正なノートID"))
                    .when(noteRepository).delete(1, "");

            assertThatThrownBy(() -> useCase.execute(1, ""))
                    .isInstanceOf(IllegalArgumentException.class)
                    .hasMessage("不正なノートID");
        }
    }
}
