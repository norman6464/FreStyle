package com.example.FreStyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.Mockito.*;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.oauth2.jwt.Jwt;

import com.example.FreStyle.dto.SessionNoteDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.GetSessionNoteUseCase;
import com.example.FreStyle.usecase.SaveSessionNoteUseCase;

@ExtendWith(MockitoExtension.class)
@DisplayName("SessionNoteController")
class SessionNoteControllerTest {

    @Mock
    private GetSessionNoteUseCase getSessionNoteUseCase;

    @Mock
    private SaveSessionNoteUseCase saveSessionNoteUseCase;

    @Mock
    private UserIdentityService userIdentityService;

    @InjectMocks
    private SessionNoteController sessionNoteController;

    private Jwt mockJwt(String sub) {
        Jwt jwt = mock(Jwt.class);
        when(jwt.getSubject()).thenReturn(sub);
        return jwt;
    }

    @Test
    @DisplayName("GET: セッションメモを取得できる")
    void getNote_returnsNote() {
        Jwt jwt = mockJwt("sub-123");
        User user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
        SessionNoteDto dto = new SessionNoteDto(100, "テストメモ", "2026-02-16T10:00:00");
        when(getSessionNoteUseCase.execute(1, 100)).thenReturn(dto);

        ResponseEntity<SessionNoteDto> response = sessionNoteController.getNote(jwt, 100);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        assertThat(response.getBody().note()).isEqualTo("テストメモ");
    }

    @Test
    @DisplayName("GET: メモが存在しない場合も200を返す")
    void getNote_returnsNullBody() {
        Jwt jwt = mockJwt("sub-123");
        User user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
        when(getSessionNoteUseCase.execute(1, 999)).thenReturn(null);

        ResponseEntity<SessionNoteDto> response = sessionNoteController.getNote(jwt, 999);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        assertThat(response.getBody()).isNull();
    }

    @Test
    @DisplayName("PUT: セッションメモを保存できる")
    void saveNote_savesNote() {
        Jwt jwt = mockJwt("sub-123");
        User user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);

        ResponseEntity<Void> response = sessionNoteController.saveNote(jwt, 100, new SessionNoteController.SaveNoteRequest("保存メモ"));

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        verify(saveSessionNoteUseCase).execute(user, 100, "保存メモ");
    }

    @Test
    @DisplayName("GET: UseCaseに正しいuserIdとsessionIdを渡している")
    void getNote_passesCorrectArguments() {
        Jwt jwt = mockJwt("sub-789");
        User user = new User();
        user.setId(55);
        when(userIdentityService.findUserBySub("sub-789")).thenReturn(user);
        when(getSessionNoteUseCase.execute(55, 200)).thenReturn(new SessionNoteDto(200, "メモ", "2026-02-16T12:00:00"));

        sessionNoteController.getNote(jwt, 200);

        verify(getSessionNoteUseCase).execute(55, 200);
    }

    @Test
    @DisplayName("PUT: UseCaseの例外がそのまま伝搬する")
    void saveNote_propagatesException() {
        Jwt jwt = mockJwt("sub-123");
        User user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
        doThrow(new RuntimeException("DB接続エラー")).when(saveSessionNoteUseCase).execute(user, 100, "メモ");

        assertThatThrownBy(
                () -> sessionNoteController.saveNote(jwt, 100, new SessionNoteController.SaveNoteRequest("メモ"))
        ).isInstanceOf(RuntimeException.class).hasMessage("DB接続エラー");
    }
}
