package com.example.FreStyle.service;

import com.example.FreStyle.dto.PresignedUrlResponse;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
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
class ProfileImageServiceTest {

    @Mock
    private S3Presigner s3Presigner;

    @Mock
    private PresignedPutObjectRequest presignedPutObjectRequest;

    private ProfileImageService service;

    @BeforeEach
    void setUp() {
        service = new ProfileImageService(s3Presigner, "test-bucket", "https://cdn.example.com");
    }

    @Test
    @DisplayName("Presigned URLを生成できる")
    void presigned_URLを生成できる() throws Exception {
        when(s3Presigner.presignPutObject(any(PutObjectPresignRequest.class)))
                .thenReturn(presignedPutObjectRequest);
        when(presignedPutObjectRequest.url()).thenReturn(new URL("https://s3.example.com/upload"));

        PresignedUrlResponse response = service.generatePresignedUrl(1, "avatar.png", "image/png");

        assertThat(response.uploadUrl()).startsWith("https://s3.example.com/");
        assertThat(response.imageUrl()).startsWith("https://cdn.example.com/profiles/1/");
        assertThat(response.imageUrl()).endsWith("_avatar.png");
    }

    @Test
    @DisplayName("S3キーにユーザーIDが含まれる")
    void S3キーにユーザーIDが含まれる() throws Exception {
        when(s3Presigner.presignPutObject(any(PutObjectPresignRequest.class)))
                .thenReturn(presignedPutObjectRequest);
        when(presignedPutObjectRequest.url()).thenReturn(new URL("https://s3.example.com/upload"));

        service.generatePresignedUrl(42, "photo.jpg", "image/jpeg");

        ArgumentCaptor<PutObjectPresignRequest> captor = ArgumentCaptor.forClass(PutObjectPresignRequest.class);
        verify(s3Presigner).presignPutObject(captor.capture());

        String key = captor.getValue().putObjectRequest().key();
        assertThat(key).startsWith("profiles/42/");
        assertThat(key).endsWith("_photo.jpg");
    }

    @Test
    @DisplayName("contentTypeがS3リクエストに含まれる")
    void contentTypeがS3リクエストに含まれる() throws Exception {
        when(s3Presigner.presignPutObject(any(PutObjectPresignRequest.class)))
                .thenReturn(presignedPutObjectRequest);
        when(presignedPutObjectRequest.url()).thenReturn(new URL("https://s3.example.com/upload"));

        service.generatePresignedUrl(1, "avatar.png", "image/png");

        ArgumentCaptor<PutObjectPresignRequest> captor = ArgumentCaptor.forClass(PutObjectPresignRequest.class);
        verify(s3Presigner).presignPutObject(captor.capture());

        assertThat(captor.getValue().putObjectRequest().contentType()).isEqualTo("image/png");
    }

    @Test
    @DisplayName("画像以外のcontentTypeはエラー")
    void 画像以外のcontentTypeはエラー() {
        assertThatThrownBy(() ->
                service.generatePresignedUrl(1, "file.exe", "application/octet-stream")
        ).isInstanceOf(IllegalArgumentException.class);
    }

    @Test
    @DisplayName("不正なcontentTypeのエラーメッセージにファイル形式が含まれる")
    void 不正なcontentTypeのエラーメッセージにファイル形式が含まれる() {
        assertThatThrownBy(() ->
                service.generatePresignedUrl(1, "file.txt", "text/plain")
        ).isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("text/plain");
    }

    @Test
    @DisplayName("JPEG形式が許可される")
    void JPEG形式が許可される() throws Exception {
        when(s3Presigner.presignPutObject(any(PutObjectPresignRequest.class)))
                .thenReturn(presignedPutObjectRequest);
        when(presignedPutObjectRequest.url()).thenReturn(new URL("https://s3.example.com/upload"));

        PresignedUrlResponse response = service.generatePresignedUrl(1, "photo.jpg", "image/jpeg");
        assertThat(response.uploadUrl()).isNotNull();
    }

    @Test
    @DisplayName("WebP形式が許可される")
    void WebP形式が許可される() throws Exception {
        when(s3Presigner.presignPutObject(any(PutObjectPresignRequest.class)))
                .thenReturn(presignedPutObjectRequest);
        when(presignedPutObjectRequest.url()).thenReturn(new URL("https://s3.example.com/upload"));

        PresignedUrlResponse response = service.generatePresignedUrl(1, "photo.webp", "image/webp");
        assertThat(response.uploadUrl()).isNotNull();
    }

    @Test
    @DisplayName("バケット名がS3リクエストに含まれる")
    void バケット名がS3リクエストに含まれる() throws Exception {
        when(s3Presigner.presignPutObject(any(PutObjectPresignRequest.class)))
                .thenReturn(presignedPutObjectRequest);
        when(presignedPutObjectRequest.url()).thenReturn(new URL("https://s3.example.com/upload"));

        service.generatePresignedUrl(1, "avatar.png", "image/png");

        ArgumentCaptor<PutObjectPresignRequest> captor = ArgumentCaptor.forClass(PutObjectPresignRequest.class);
        verify(s3Presigner).presignPutObject(captor.capture());

        assertThat(captor.getValue().putObjectRequest().bucket()).isEqualTo("test-bucket");
    }
}
