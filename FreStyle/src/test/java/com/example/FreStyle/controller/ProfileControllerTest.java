package com.example.FreStyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.mockito.Mockito.doThrow;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.Map;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.oauth2.jwt.Jwt;

import com.example.FreStyle.dto.PresignedUrlRequest;
import com.example.FreStyle.dto.PresignedUrlResponse;
import com.example.FreStyle.dto.ProfileDto;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.form.ProfileForm;
import com.example.FreStyle.usecase.GenerateProfileImageUrlUseCase;
import com.example.FreStyle.usecase.GetProfileUseCase;
import com.example.FreStyle.usecase.UpdateProfileUseCase;

@ExtendWith(MockitoExtension.class)
@DisplayName("ProfileController")
class ProfileControllerTest {

    @Mock
    private GetProfileUseCase getProfileUseCase;

    @Mock
    private UpdateProfileUseCase updateProfileUseCase;

    @Mock
    private GenerateProfileImageUrlUseCase generateProfileImageUrlUseCase;

    @InjectMocks
    private ProfileController profileController;

    @Nested
    @DisplayName("getProfile")
    class GetProfile {

        @Test
        @DisplayName("正常にプロフィールを取得できる")
        void returnsProfile() {
            Jwt jwt = mock(Jwt.class);
            when(jwt.getSubject()).thenReturn("sub-123");
            when(getProfileUseCase.execute("sub-123"))
                    .thenReturn(new ProfileDto("テストユーザー", "テスト自己紹介", null));

            ResponseEntity<ProfileDto> response = profileController.getProfile(jwt);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(response.getBody().getName()).isEqualTo("テストユーザー");
            assertThat(response.getBody().getBio()).isEqualTo("テスト自己紹介");
        }

        @Test
        @DisplayName("ユーザーが見つからない場合ResourceNotFoundExceptionが伝搬する")
        void throwsWhenUserNotFound() {
            Jwt jwt = mock(Jwt.class);
            when(jwt.getSubject()).thenReturn("sub-999");
            when(getProfileUseCase.execute("sub-999"))
                    .thenThrow(new ResourceNotFoundException("ユーザーが見つかりません"));

            assertThrows(ResourceNotFoundException.class,
                    () -> profileController.getProfile(jwt));
        }

        @Test
        @DisplayName("JWTのsubjectがnullの場合useCaseにnullが渡される")
        void passesNullSubjectToUseCase() {
            Jwt jwt = mock(Jwt.class);
            when(jwt.getSubject()).thenReturn(null);
            when(getProfileUseCase.execute(null))
                    .thenThrow(new ResourceNotFoundException("ユーザーが見つかりません"));

            assertThrows(ResourceNotFoundException.class,
                    () -> profileController.getProfile(jwt));
            verify(getProfileUseCase).execute(null);
        }
    }

    @Nested
    @DisplayName("updateProfile")
    class UpdateProfile {

        @Test
        @DisplayName("Cognitoユーザーのプロフィールを更新できる")
        void updatesCognitoUserProfile() {
            Jwt jwt = mock(Jwt.class);
            ProfileForm form = new ProfileForm("新しい名前", "新しい自己紹介", null);

            ResponseEntity<Map<String, String>> response = profileController.updateProfile(jwt, form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            verify(updateProfileUseCase).execute(jwt, form);
        }

        @Test
        @DisplayName("入力値不正時はIllegalArgumentExceptionが伝搬する")
        void throwsWhenInvalidInput() {
            Jwt jwt = mock(Jwt.class);
            ProfileForm form = new ProfileForm("", "自己紹介", null);

            doThrow(new IllegalArgumentException("名前が不正です"))
                    .when(updateProfileUseCase).execute(jwt, form);

            assertThrows(IllegalArgumentException.class,
                    () -> profileController.updateProfile(jwt, form));
        }
    }

    @Nested
    @DisplayName("getProfileImagePresignedUrl")
    class GetProfileImagePresignedUrl {

        @Test
        @DisplayName("Presigned URLを正常に取得できる")
        void returnsPresignedUrl() {
            Jwt jwt = mock(Jwt.class);
            when(jwt.getSubject()).thenReturn("sub-123");
            PresignedUrlRequest request = new PresignedUrlRequest("avatar.png", "image/png");
            PresignedUrlResponse expected = new PresignedUrlResponse(
                    "https://s3.example.com/upload", "https://cdn.example.com/profiles/1/avatar.png");
            when(generateProfileImageUrlUseCase.execute("sub-123", "avatar.png", "image/png"))
                    .thenReturn(expected);

            ResponseEntity<PresignedUrlResponse> response = profileController.getProfileImagePresignedUrl(jwt, request);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(response.getBody()).isEqualTo(expected);
        }

        @Test
        @DisplayName("不正なcontentTypeでIllegalArgumentExceptionが伝搬する")
        void throwsForInvalidContentType() {
            Jwt jwt = mock(Jwt.class);
            when(jwt.getSubject()).thenReturn("sub-123");
            PresignedUrlRequest request = new PresignedUrlRequest("file.exe", "application/octet-stream");
            when(generateProfileImageUrlUseCase.execute("sub-123", "file.exe", "application/octet-stream"))
                    .thenThrow(new IllegalArgumentException("許可されていないファイル形式です"));

            assertThrows(IllegalArgumentException.class,
                    () -> profileController.getProfileImagePresignedUrl(jwt, request));
        }
    }
}
