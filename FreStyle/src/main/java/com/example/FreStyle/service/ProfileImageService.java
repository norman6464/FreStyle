package com.example.FreStyle.service;

import com.example.FreStyle.dto.PresignedUrlResponse;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import software.amazon.awssdk.services.s3.model.PutObjectRequest;
import software.amazon.awssdk.services.s3.presigner.S3Presigner;
import software.amazon.awssdk.services.s3.presigner.model.PutObjectPresignRequest;
import software.amazon.awssdk.services.s3.presigner.model.PresignedPutObjectRequest;

import java.time.Duration;
import java.util.Set;
import java.util.UUID;

@Service
public class ProfileImageService {

    private static final Set<String> ALLOWED_CONTENT_TYPES = Set.of(
            "image/png", "image/jpeg", "image/gif", "image/webp"
    );

    private final S3Presigner s3Presigner;
    private final String bucketName;
    private final String cdnBaseUrl;

    public ProfileImageService(
            S3Presigner s3Presigner,
            @Value("${aws.s3.note-images-bucket}") String bucketName,
            @Value("${aws.s3.note-images-cdn-url}") String cdnBaseUrl) {
        this.s3Presigner = s3Presigner;
        this.bucketName = bucketName;
        this.cdnBaseUrl = cdnBaseUrl;
    }

    public PresignedUrlResponse generatePresignedUrl(Integer userId, String fileName, String contentType) {
        if (!ALLOWED_CONTENT_TYPES.contains(contentType)) {
            throw new IllegalArgumentException("許可されていないファイル形式です: " + contentType);
        }

        String key = buildKey(userId, fileName);

        PutObjectRequest putObjectRequest = PutObjectRequest.builder()
                .bucket(bucketName)
                .key(key)
                .contentType(contentType)
                .build();

        PutObjectPresignRequest presignRequest = PutObjectPresignRequest.builder()
                .signatureDuration(Duration.ofMinutes(10))
                .putObjectRequest(putObjectRequest)
                .build();

        PresignedPutObjectRequest presignedRequest = s3Presigner.presignPutObject(presignRequest);

        String uploadUrl = presignedRequest.url().toString();
        String imageUrl = cdnBaseUrl + "/" + key;

        return new PresignedUrlResponse(uploadUrl, imageUrl);
    }

    private String buildKey(Integer userId, String fileName) {
        String uuid = UUID.randomUUID().toString().substring(0, 8);
        return "profiles/" + userId + "/" + uuid + "_" + fileName;
    }
}
