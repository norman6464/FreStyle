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

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class GetNotesByUserIdUseCaseTest {

    @Mock
    private NoteRepository noteRepository;

    @InjectMocks
    private GetNotesByUserIdUseCase useCase;

    @Nested
    @DisplayName("execute - ユーザーIDでノート一覧取得")
    class ExecuteTest {

        @Test
        @DisplayName("ユーザーのノート一覧をRepository経由で取得する")
        void shouldReturnNotesByUserId() {
            NoteDto note1 = new NoteDto("note-1", 1, "タイトル1", "内容1", false, 1000L, 2000L);
            NoteDto note2 = new NoteDto("note-2", 1, "タイトル2", "内容2", true, 1500L, 3000L);
            when(noteRepository.findByUserId(1)).thenReturn(List.of(note1, note2));

            List<NoteDto> result = useCase.execute(1);

            assertThat(result).hasSize(2);
            assertThat(result.get(0).getNoteId()).isEqualTo("note-1");
            assertThat(result.get(1).getNoteId()).isEqualTo("note-2");
            verify(noteRepository, times(1)).findByUserId(1);
        }

        @Test
        @DisplayName("ノートが存在しない場合は空リストを返す")
        void shouldReturnEmptyListWhenNoNotes() {
            when(noteRepository.findByUserId(1)).thenReturn(List.of());

            List<NoteDto> result = useCase.execute(1);

            assertThat(result).isEmpty();
            verify(noteRepository, times(1)).findByUserId(1);
        }
    }
}
