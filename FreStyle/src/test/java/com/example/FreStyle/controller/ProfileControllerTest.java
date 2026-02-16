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
    }

    @Nested
    @DisplayName("updateProfile")
    class UpdateProfile {

        @Test
        @DisplayName("Cognitoユーザーのプロフィールを更新できる")
        void updatesCognitoUserProfile() {
            Jwt jwt = mock(Jwt.class);
            ProfileForm form = new ProfileForm("新しい名前", "新しい自己紹介");

            ResponseEntity<Map<String, String>> response = profileController.updateProfile(jwt, form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            verify(updateProfileUseCase).execute(jwt, form);
        }

        @Test
        @DisplayName("入力値不正時はIllegalArgumentExceptionが伝搬する")
        void throwsWhenInvalidInput() {
            Jwt jwt = mock(Jwt.class);
            ProfileForm form = new ProfileForm("", "自己紹介");

            doThrow(new IllegalArgumentException("名前が不正です"))
                    .when(updateProfileUseCase).execute(jwt, form);

            assertThrows(IllegalArgumentException.class,
                    () -> profileController.updateProfile(jwt, form));
        }
    }
}
