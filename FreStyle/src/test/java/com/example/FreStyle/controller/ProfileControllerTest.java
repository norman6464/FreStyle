package com.example.FreStyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.doThrow;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

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

import com.example.FreStyle.dto.ProfileDto;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.form.ProfileForm;
import com.example.FreStyle.usecase.GetProfileUseCase;
import com.example.FreStyle.usecase.UpdateProfileUseCase;

@ExtendWith(MockitoExtension.class)
@DisplayName("ProfileController")
class ProfileControllerTest {

    @Mock
    private GetProfileUseCase getProfileUseCase;

    @Mock
    private UpdateProfileUseCase updateProfileUseCase;

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
                    .thenReturn(new ProfileDto("テストユーザー", "テスト自己紹介"));

            ResponseEntity<?> response = profileController.getProfile(jwt);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            ProfileDto body = (ProfileDto) response.getBody();
            assertThat(body.getName()).isEqualTo("テストユーザー");
            assertThat(body.getBio()).isEqualTo("テスト自己紹介");
        }

        @Test
        @DisplayName("JWTがnullの場合401を返す")
        void returnsUnauthorizedWhenJwtNull() {
            ResponseEntity<?> response = profileController.getProfile(null);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.UNAUTHORIZED);
        }

        @Test
        @DisplayName("ユーザーが見つからない場合404を返す")
        void returnsNotFoundWhenUserNotFound() {
            Jwt jwt = mock(Jwt.class);
            when(jwt.getSubject()).thenReturn("sub-999");
            when(getProfileUseCase.execute("sub-999"))
                    .thenThrow(new ResourceNotFoundException("ユーザーが見つかりません"));

            ResponseEntity<?> response = profileController.getProfile(jwt);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.NOT_FOUND);
        }
    }

    @Nested
    @DisplayName("updateProfile")
    class UpdateProfile {

        @Test
        @DisplayName("Cognitoユーザーのプロフィールを更新できる")
        void updatesCognitoUserProfile() {
            Jwt jwt = mock(Jwt.class);
            when(jwt.getSubject()).thenReturn("sub-123");

            ProfileForm form = new ProfileForm("新しい名前", "新しい自己紹介");

            ResponseEntity<?> response = profileController.updateProfile(jwt, form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            verify(updateProfileUseCase).execute(jwt, form);
        }

        @Test
        @DisplayName("JWTがnullの場合401を返す")
        void returnsUnauthorizedWhenJwtNull() {
            ProfileForm form = new ProfileForm("名前", "自己紹介");

            ResponseEntity<?> response = profileController.updateProfile(null, form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.UNAUTHORIZED);
        }

        @Test
        @DisplayName("入力値不正時は400を返す")
        void returnsBadRequestWhenInvalidInput() {
            Jwt jwt = mock(Jwt.class);
            when(jwt.getSubject()).thenReturn("sub-123");

            ProfileForm form = new ProfileForm("", "自己紹介");

            doThrow(new IllegalArgumentException("名前が不正です"))
                    .when(updateProfileUseCase).execute(jwt, form);

            ResponseEntity<?> response = profileController.updateProfile(jwt, form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.BAD_REQUEST);
        }
    }
}
