package com.example.FreStyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.Mockito.*;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.ResponseEntity;
import org.springframework.security.oauth2.jwt.Jwt;

import com.example.FreStyle.dto.PresignedUrlRequest;
import com.example.FreStyle.dto.PresignedUrlResponse;
import com.example.FreStyle.usecase.GeneratePresignedUrlUseCase;

@ExtendWith(MockitoExtension.class)
@DisplayName("NoteImageController")
class NoteImageControllerTest {

    @Mock
    private GeneratePresignedUrlUseCase generatePresignedUrlUseCase;

    @InjectMocks
    private NoteImageController noteImageController;

    private Jwt jwt;

    @BeforeEach
    void setUp() {
        jwt = mock(Jwt.class);
        when(jwt.getSubject()).thenReturn("test-sub");
    }

    @Test
    @DisplayName("Presigned URLを生成して返す")
    void shouldReturnPresignedUrl() {
        PresignedUrlResponse expected = new PresignedUrlResponse(
                "https://s3.example.com/upload", "https://cdn.example.com/notes/1/note1/abc_image.png"
        );
        when(generatePresignedUrlUseCase.execute("test-sub", "note1", "image.png", "image/png"))
                .thenReturn(expected);

        ResponseEntity<?> response = noteImageController.getPresignedUrl(
                jwt, "note1", new PresignedUrlRequest("image.png", "image/png")
        );

        assertThat(response.getStatusCode().value()).isEqualTo(200);
        PresignedUrlResponse body = (PresignedUrlResponse) response.getBody();
        assertThat(body.uploadUrl()).isEqualTo("https://s3.example.com/upload");
    }

    @Test
    @DisplayName("不正なcontentTypeは例外がスローされる")
    void shouldThrowOnInvalidContentType() {
        when(generatePresignedUrlUseCase.execute("test-sub", "note1", "file.exe", "application/octet-stream"))
                .thenThrow(new IllegalArgumentException("許可されていないファイル形式です"));

        assertThatThrownBy(() -> noteImageController.getPresignedUrl(
                jwt, "note1", new PresignedUrlRequest("file.exe", "application/octet-stream")
        )).isInstanceOf(IllegalArgumentException.class)
                .hasMessage("許可されていないファイル形式です");
    }

    @Test
    @DisplayName("レスポンスにimageUrlも含まれる")
    void shouldReturnCdnUrlInResponse() {
        PresignedUrlResponse expected = new PresignedUrlResponse(
                "https://s3.example.com/upload", "https://cdn.example.com/notes/1/note1/abc_image.png"
        );
        when(generatePresignedUrlUseCase.execute("test-sub", "note1", "image.png", "image/png"))
                .thenReturn(expected);

        ResponseEntity<?> response = noteImageController.getPresignedUrl(
                jwt, "note1", new PresignedUrlRequest("image.png", "image/png")
        );

        PresignedUrlResponse body = (PresignedUrlResponse) response.getBody();
        assertThat(body.imageUrl()).isEqualTo("https://cdn.example.com/notes/1/note1/abc_image.png");
    }

    @Test
    @DisplayName("UseCaseに正しい引数が渡される")
    void shouldPassCorrectArgumentsToUseCase() {
        when(generatePresignedUrlUseCase.execute("test-sub", "noteXYZ", "photo.jpg", "image/jpeg"))
                .thenReturn(new PresignedUrlResponse("https://s3.example.com/upload", "https://cdn.example.com/photo.jpg"));

        noteImageController.getPresignedUrl(
                jwt, "noteXYZ", new PresignedUrlRequest("photo.jpg", "image/jpeg")
        );

        verify(generatePresignedUrlUseCase).execute("test-sub", "noteXYZ", "photo.jpg", "image/jpeg");
    }

    @Test
    @DisplayName("S3エラー時は例外がスローされる")
    void shouldThrowOnS3Error() {
        when(generatePresignedUrlUseCase.execute("test-sub", "note1", "image.png", "image/png"))
                .thenThrow(new RuntimeException("S3 connection failed"));

        assertThatThrownBy(() -> noteImageController.getPresignedUrl(
                jwt, "note1", new PresignedUrlRequest("image.png", "image/png")
        )).isInstanceOf(RuntimeException.class);
    }
}
