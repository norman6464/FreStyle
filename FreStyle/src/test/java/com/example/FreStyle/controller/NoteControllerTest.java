package com.example.FreStyle.controller;

import com.example.FreStyle.dto.NoteDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.CreateNoteUseCase;
import com.example.FreStyle.usecase.DeleteNoteUseCase;
import com.example.FreStyle.usecase.GetNotesByUserIdUseCase;
import com.example.FreStyle.usecase.UpdateNoteUseCase;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.ResponseEntity;
import org.springframework.security.oauth2.jwt.Jwt;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class NoteControllerTest {

    @Mock
    private GetNotesByUserIdUseCase getNotesByUserIdUseCase;

    @Mock
    private CreateNoteUseCase createNoteUseCase;

    @Mock
    private UpdateNoteUseCase updateNoteUseCase;

    @Mock
    private DeleteNoteUseCase deleteNoteUseCase;

    @Mock
    private UserIdentityService userIdentityService;

    @InjectMocks
    private NoteController noteController;

    private Jwt jwt;
    private User user;

    @BeforeEach
    void setUp() {
        jwt = mock(Jwt.class);
        when(jwt.getSubject()).thenReturn("test-sub");

        user = new User();
        user.setId(1);
        user.setName("テストユーザー");

        when(userIdentityService.findUserBySub("test-sub")).thenReturn(user);
    }

    @Nested
    @DisplayName("GET /api/notes - ノート一覧取得")
    class GetNotesTest {

        @Test
        @DisplayName("UseCase経由でユーザーのノート一覧を返す")
        void shouldReturnUserNotesViaUseCase() {
            NoteDto note1 = new NoteDto("note-1", 1, "タイトル1", "内容1", false, 1000L, 2000L);
            NoteDto note2 = new NoteDto("note-2", 1, "タイトル2", "内容2", true, 1500L, 3000L);
            when(getNotesByUserIdUseCase.execute(1)).thenReturn(List.of(note1, note2));

            ResponseEntity<List<NoteDto>> response = noteController.getNotes(jwt);

            assertThat(response.getStatusCode().value()).isEqualTo(200);
            assertThat(response.getBody()).hasSize(2);
            assertThat(response.getBody().get(0).getTitle()).isEqualTo("タイトル1");
            verify(getNotesByUserIdUseCase).execute(1);
        }
    }

    @Nested
    @DisplayName("POST /api/notes - ノート作成")
    class CreateNoteTest {

        @Test
        @DisplayName("UseCase経由で新しいノートを作成して返す")
        void shouldCreateNoteViaUseCase() {
            NoteDto created = new NoteDto("note-new", 1, "新しいノート", "", false, 1000L, 1000L);
            when(createNoteUseCase.execute(1, "新しいノート")).thenReturn(created);

            ResponseEntity<NoteDto> response = noteController.createNote(jwt,
                    new NoteController.CreateNoteRequest("新しいノート"));

            assertThat(response.getStatusCode().value()).isEqualTo(200);
            assertThat(response.getBody().getTitle()).isEqualTo("新しいノート");
            assertThat(response.getBody().getNoteId()).isEqualTo("note-new");
            verify(createNoteUseCase).execute(1, "新しいノート");
        }
    }

    @Nested
    @DisplayName("PUT /api/notes/{noteId} - ノート更新")
    class UpdateNoteTest {

        @Test
        @DisplayName("UseCase経由でノートを更新して204を返す")
        void shouldUpdateNoteViaUseCase() {
            doNothing().when(updateNoteUseCase).execute(1, "note-1", "更新タイトル", "更新内容", false);

            ResponseEntity<Void> response = noteController.updateNote(jwt, "note-1",
                    new NoteController.UpdateNoteRequest("更新タイトル", "更新内容", false));

            assertThat(response.getStatusCode().value()).isEqualTo(204);
            verify(updateNoteUseCase).execute(1, "note-1", "更新タイトル", "更新内容", false);
        }
    }

    @Nested
    @DisplayName("DELETE /api/notes/{noteId} - ノート削除")
    class DeleteNoteTest {

        @Test
        @DisplayName("UseCase経由でノートを削除して204を返す")
        void shouldDeleteNoteViaUseCase() {
            doNothing().when(deleteNoteUseCase).execute(1, "note-1");

            ResponseEntity<Void> response = noteController.deleteNote(jwt, "note-1");

            assertThat(response.getStatusCode().value()).isEqualTo(204);
            verify(deleteNoteUseCase).execute(1, "note-1");
        }
    }
}
