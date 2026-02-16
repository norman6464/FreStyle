package com.example.FreStyle.usecase;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.PresignedUrlResponse;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.NoteImageService;
import com.example.FreStyle.service.UserIdentityService;

@ExtendWith(MockitoExtension.class)
class GeneratePresignedUrlUseCaseTest {

    @Mock
    private NoteImageService noteImageService;

    @Mock
    private UserIdentityService userIdentityService;

    @InjectMocks
    private GeneratePresignedUrlUseCase useCase;

    @Test
    @DisplayName("Presigned URLを生成して返す")
    void execute_generatesPresignedUrl() {
        User user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("test-sub")).thenReturn(user);
        PresignedUrlResponse expected = new PresignedUrlResponse(
                "https://s3.example.com/upload", "https://cdn.example.com/image.png");
        when(noteImageService.generatePresignedUrl(1, "note1", "image.png", "image/png"))
                .thenReturn(expected);

        PresignedUrlResponse result = useCase.execute("test-sub", "note1", "image.png", "image/png");

        assertEquals(expected, result);
    }

    @Test
    @DisplayName("不正なcontentTypeでIllegalArgumentExceptionがスローされる")
    void execute_throwsForInvalidContentType() {
        User user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("test-sub")).thenReturn(user);
        when(noteImageService.generatePresignedUrl(1, "note1", "file.exe", "application/octet-stream"))
                .thenThrow(new IllegalArgumentException("許可されていないファイル形式です"));

        assertThrows(IllegalArgumentException.class,
                () -> useCase.execute("test-sub", "note1", "file.exe", "application/octet-stream"));
    }

    @Test
    @DisplayName("UserIdentityServiceが例外をスローした場合そのまま伝搬する")
    void execute_propagatesUserIdentityException() {
        when(userIdentityService.findUserBySub("unknown-sub"))
                .thenThrow(new RuntimeException("ユーザーが見つかりません"));

        assertThrows(RuntimeException.class,
                () -> useCase.execute("unknown-sub", "note1", "image.png", "image/png"));
    }
}
