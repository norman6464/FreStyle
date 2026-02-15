package com.example.FreStyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
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
import com.example.FreStyle.entity.User;
import com.example.FreStyle.form.ProfileForm;
import com.example.FreStyle.service.CognitoAuthService;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.service.UserService;

@ExtendWith(MockitoExtension.class)
@DisplayName("ProfileController")
class ProfileControllerTest {

    @Mock
    private CognitoAuthService cognitoAuthService;

    @Mock
    private UserService userService;

    @Mock
    private UserIdentityService userIdentityService;

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

            User user = new User();
            user.setName("テストユーザー");
            user.setBio("テスト自己紹介");
            when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);

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
            when(userIdentityService.findUserBySub("sub-999")).thenReturn(null);

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
            when(jwt.hasClaim("cognito:groups")).thenReturn(false);
            when(jwt.getTokenValue()).thenReturn("access-token");

            ProfileForm form = new ProfileForm("新しい名前", "新しい自己紹介");

            ResponseEntity<?> response = profileController.updateProfile(jwt, form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            verify(cognitoAuthService).updateUserProfile("access-token", "新しい名前");
            verify(userService).updateUser(form, "sub-123");
        }

        @Test
        @DisplayName("OIDCユーザーのプロフィールをDB更新のみで更新できる")
        void updatesOidcUserProfile() {
            Jwt jwt = mock(Jwt.class);
            when(jwt.getSubject()).thenReturn("sub-456");
            when(jwt.hasClaim("cognito:groups")).thenReturn(true);

            ProfileForm form = new ProfileForm("OIDC名前", "OIDC自己紹介");

            ResponseEntity<?> response = profileController.updateProfile(jwt, form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            verify(userService).updateUser(form, "sub-456");
        }

        @Test
        @DisplayName("JWTがnullの場合401を返す")
        void returnsUnauthorizedWhenJwtNull() {
            ProfileForm form = new ProfileForm("名前", "自己紹介");

            ResponseEntity<?> response = profileController.updateProfile(null, form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.UNAUTHORIZED);
        }
    }
}
