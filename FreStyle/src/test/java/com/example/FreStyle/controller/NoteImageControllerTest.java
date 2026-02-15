package com.example.FreStyle.controller;

import com.example.FreStyle.dto.PresignedUrlRequest;
import com.example.FreStyle.dto.PresignedUrlResponse;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.NoteImageService;
import com.example.FreStyle.service.UserIdentityService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.ResponseEntity;
import org.springframework.security.oauth2.jwt.Jwt;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class NoteImageControllerTest {

    @Mock
    private NoteImageService noteImageService;

    @Mock
    private UserIdentityService userIdentityService;

    @InjectMocks
    private NoteImageController noteImageController;

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

    @Test
    @DisplayName("Presigned URLを生成して返す")
    void shouldReturnPresignedUrl() {
        PresignedUrlResponse expected = new PresignedUrlResponse(
                "https://s3.example.com/upload", "https://cdn.example.com/notes/1/note1/abc_image.png"
        );
        when(noteImageService.generatePresignedUrl(1, "note1", "image.png", "image/png"))
                .thenReturn(expected);

        ResponseEntity<PresignedUrlResponse> response = noteImageController.getPresignedUrl(
                jwt, "note1", new PresignedUrlRequest("image.png", "image/png")
        );

        assertThat(response.getStatusCode().value()).isEqualTo(200);
        assertThat(response.getBody().uploadUrl()).isEqualTo("https://s3.example.com/upload");
        assertThat(response.getBody().imageUrl()).startsWith("https://cdn.example.com/notes/1/note1/");
        verify(noteImageService).generatePresignedUrl(1, "note1", "image.png", "image/png");
    }

    @Test
    @DisplayName("NoteImageServiceにユーザーIDを渡す")
    void shouldPassUserIdToService() {
        user.setId(42);
        PresignedUrlResponse expected = new PresignedUrlResponse(
                "https://s3.example.com/upload", "https://cdn.example.com/notes/42/noteXYZ/abc_photo.jpg"
        );
        when(noteImageService.generatePresignedUrl(42, "noteXYZ", "photo.jpg", "image/jpeg"))
                .thenReturn(expected);

        noteImageController.getPresignedUrl(
                jwt, "noteXYZ", new PresignedUrlRequest("photo.jpg", "image/jpeg")
        );

        verify(noteImageService).generatePresignedUrl(42, "noteXYZ", "photo.jpg", "image/jpeg");
    }

    @Test
    @DisplayName("不正なcontentTypeはサービス層でエラーになる")
    void shouldPropagateServiceException() {
        when(noteImageService.generatePresignedUrl(1, "note1", "file.exe", "application/octet-stream"))
                .thenThrow(new IllegalArgumentException("許可されていないファイル形式です: application/octet-stream"));

        org.junit.jupiter.api.Assertions.assertThrows(IllegalArgumentException.class, () ->
                noteImageController.getPresignedUrl(
                        jwt, "note1", new PresignedUrlRequest("file.exe", "application/octet-stream")
                )
        );
    }
}
