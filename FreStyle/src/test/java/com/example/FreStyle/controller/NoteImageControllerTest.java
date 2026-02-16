package com.example.FreStyle.controller;

import com.example.FreStyle.dto.PresignedUrlRequest;
import com.example.FreStyle.dto.PresignedUrlResponse;
import com.example.FreStyle.usecase.GeneratePresignedUrlUseCase;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.ResponseEntity;
import org.springframework.security.oauth2.jwt.Jwt;

import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
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
    @DisplayName("不正なcontentTypeは400エラーを返す")
    void shouldReturn400ForInvalidContentType() {
        when(generatePresignedUrlUseCase.execute("test-sub", "note1", "file.exe", "application/octet-stream"))
                .thenThrow(new IllegalArgumentException("許可されていないファイル形式です"));

        ResponseEntity<?> response = noteImageController.getPresignedUrl(
                jwt, "note1", new PresignedUrlRequest("file.exe", "application/octet-stream")
        );

        assertThat(response.getStatusCode().value()).isEqualTo(400);
    }

    @SuppressWarnings("unchecked")
    @Test
    @DisplayName("不正なcontentType時のエラーレスポンスにメッセージが含まれる")
    void shouldReturn400WithErrorMessage() {
        when(generatePresignedUrlUseCase.execute("test-sub", "note1", "file.exe", "application/octet-stream"))
                .thenThrow(new IllegalArgumentException("許可されていないファイル形式です"));

        ResponseEntity<?> response = noteImageController.getPresignedUrl(
                jwt, "note1", new PresignedUrlRequest("file.exe", "application/octet-stream")
        );

        Map<String, String> body = (Map<String, String>) response.getBody();
        assertThat(body).containsEntry("error", "許可されていないファイル形式です");
    }

    @Test
    @DisplayName("S3エラー時は500エラーを返す")
    void shouldReturn500ForS3Error() {
        when(generatePresignedUrlUseCase.execute("test-sub", "note1", "image.png", "image/png"))
                .thenThrow(new RuntimeException("S3 connection failed"));

        ResponseEntity<?> response = noteImageController.getPresignedUrl(
                jwt, "note1", new PresignedUrlRequest("image.png", "image/png")
        );

        assertThat(response.getStatusCode().value()).isEqualTo(500);
    }

    @SuppressWarnings("unchecked")
    @Test
    @DisplayName("S3エラー時のエラーレスポンスに固定メッセージが含まれる")
    void shouldReturn500WithFixedErrorMessage() {
        when(generatePresignedUrlUseCase.execute("test-sub", "note1", "image.png", "image/png"))
                .thenThrow(new RuntimeException("S3 connection failed"));

        ResponseEntity<?> response = noteImageController.getPresignedUrl(
                jwt, "note1", new PresignedUrlRequest("image.png", "image/png")
        );

        Map<String, String> body = (Map<String, String>) response.getBody();
        assertThat(body).containsEntry("error", "画像アップロードの準備に失敗しました");
    }
}
