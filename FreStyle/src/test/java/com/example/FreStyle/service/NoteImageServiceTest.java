package com.example.FreStyle.service;

import com.example.FreStyle.dto.PresignedUrlResponse;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import software.amazon.awssdk.services.s3.presigner.S3Presigner;
import software.amazon.awssdk.services.s3.presigner.model.PutObjectPresignRequest;
import software.amazon.awssdk.services.s3.presigner.model.PresignedPutObjectRequest;

import java.net.URL;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class NoteImageServiceTest {

    @Mock
    private S3Presigner s3Presigner;

    @Mock
    private PresignedPutObjectRequest presignedPutObjectRequest;

    private NoteImageService service;

    @BeforeEach
    void setUp() {
        service = new NoteImageService(s3Presigner, "test-bucket", "https://cdn.example.com");
    }

    @Test
    void presigned_URLを生成できる() throws Exception {
        when(s3Presigner.presignPutObject(any(PutObjectPresignRequest.class)))
                .thenReturn(presignedPutObjectRequest);
        when(presignedPutObjectRequest.url()).thenReturn(new URL("https://s3.example.com/upload"));

        PresignedUrlResponse response = service.generatePresignedUrl(1, "note1", "image.png", "image/png");

        assertThat(response.uploadUrl()).startsWith("https://s3.example.com/");
        assertThat(response.imageUrl()).startsWith("https://cdn.example.com/notes/1/note1/");
        assertThat(response.imageUrl()).endsWith("_image.png");
    }

    @Test
    void S3キーにユーザーIDとノートIDが含まれる() throws Exception {
        when(s3Presigner.presignPutObject(any(PutObjectPresignRequest.class)))
                .thenReturn(presignedPutObjectRequest);
        when(presignedPutObjectRequest.url()).thenReturn(new URL("https://s3.example.com/upload"));

        service.generatePresignedUrl(42, "noteXYZ", "photo.jpg", "image/jpeg");

        ArgumentCaptor<PutObjectPresignRequest> captor = ArgumentCaptor.forClass(PutObjectPresignRequest.class);
        verify(s3Presigner).presignPutObject(captor.capture());

        String key = captor.getValue().putObjectRequest().key();
        assertThat(key).startsWith("notes/42/noteXYZ/");
        assertThat(key).endsWith("_photo.jpg");
    }

    @Test
    void contentTypeがS3リクエストに含まれる() throws Exception {
        when(s3Presigner.presignPutObject(any(PutObjectPresignRequest.class)))
                .thenReturn(presignedPutObjectRequest);
        when(presignedPutObjectRequest.url()).thenReturn(new URL("https://s3.example.com/upload"));

        service.generatePresignedUrl(1, "note1", "image.png", "image/png");

        ArgumentCaptor<PutObjectPresignRequest> captor = ArgumentCaptor.forClass(PutObjectPresignRequest.class);
        verify(s3Presigner).presignPutObject(captor.capture());

        assertThat(captor.getValue().putObjectRequest().contentType()).isEqualTo("image/png");
    }

    @Test
    void 画像以外のcontentTypeはエラー() {
        assertThatThrownBy(() ->
                service.generatePresignedUrl(1, "note1", "file.exe", "application/octet-stream")
        ).isInstanceOf(IllegalArgumentException.class);
    }

    @Test
    void S3Presignerが例外をスローした場合そのまま伝搬する() {
        when(s3Presigner.presignPutObject(any(PutObjectPresignRequest.class)))
                .thenThrow(new RuntimeException("S3 error"));

        assertThatThrownBy(() ->
                service.generatePresignedUrl(1, "note1", "image.png", "image/png")
        ).isInstanceOf(RuntimeException.class)
                .hasMessage("S3 error");
    }

    @Test
    void webp形式のcontentTypeが許可される() throws Exception {
        when(s3Presigner.presignPutObject(any(PutObjectPresignRequest.class)))
                .thenReturn(presignedPutObjectRequest);
        when(presignedPutObjectRequest.url()).thenReturn(new URL("https://s3.example.com/upload"));

        PresignedUrlResponse response = service.generatePresignedUrl(1, "note1", "image.webp", "image/webp");

        assertThat(response.uploadUrl()).isNotNull();
    }

    @Test
    void svg形式のcontentTypeが許可される() throws Exception {
        when(s3Presigner.presignPutObject(any(PutObjectPresignRequest.class)))
                .thenReturn(presignedPutObjectRequest);
        when(presignedPutObjectRequest.url()).thenReturn(new URL("https://s3.example.com/upload"));

        PresignedUrlResponse response = service.generatePresignedUrl(1, "note1", "icon.svg", "image/svg+xml");

        assertThat(response.uploadUrl()).isNotNull();
    }
}
