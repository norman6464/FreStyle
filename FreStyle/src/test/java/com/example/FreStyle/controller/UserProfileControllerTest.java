package com.example.FreStyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.*;

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

import com.example.FreStyle.dto.UserProfileDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.form.UserProfileForm;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.service.UserProfileService;

@ExtendWith(MockitoExtension.class)
@DisplayName("UserProfileController")
class UserProfileControllerTest {

    @Mock
    private UserProfileService userProfileService;

    @Mock
    private UserIdentityService userIdentityService;

    @InjectMocks
    private UserProfileController userProfileController;

    private Jwt mockJwt(String sub) {
        Jwt jwt = mock(Jwt.class);
        when(jwt.getSubject()).thenReturn(sub);
        return jwt;
    }

    private User createUser(Integer id) {
        User user = new User();
        user.setId(id);
        user.setName("テストユーザー");
        return user;
    }

    private UserProfileDto createProfileDto() {
        UserProfileDto dto = new UserProfileDto();
        dto.setId(1);
        dto.setUserId(10);
        dto.setDisplayName("表示名");
        return dto;
    }

    @Nested
    @DisplayName("getMyProfile")
    class GetMyProfile {

        @Test
        @DisplayName("正常にプロファイルを取得できる")
        void returnsProfile() {
            Jwt jwt = mockJwt("sub-123");
            User user = createUser(10);
            when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
            when(userProfileService.getProfileByUserId(10)).thenReturn(createProfileDto());

            ResponseEntity<?> response = userProfileController.getMyProfile(jwt);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(response.getBody()).isInstanceOf(UserProfileDto.class);
        }

        @Test
        @DisplayName("プロファイル未作成の場合メッセージを返す")
        void returnsMessageWhenNoProfile() {
            Jwt jwt = mockJwt("sub-123");
            User user = createUser(10);
            when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
            when(userProfileService.getProfileByUserId(10)).thenReturn(null);

            ResponseEntity<?> response = userProfileController.getMyProfile(jwt);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        }

        @Test
        @DisplayName("JWTがnullの場合401を返す")
        void returns401WhenJwtNull() {
            ResponseEntity<?> response = userProfileController.getMyProfile(null);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.UNAUTHORIZED);
        }
    }

    @Nested
    @DisplayName("createMyProfile")
    class CreateMyProfile {

        @Test
        @DisplayName("正常にプロファイルを作成できる")
        void createsProfile() {
            Jwt jwt = mockJwt("sub-123");
            User user = createUser(10);
            UserProfileForm form = new UserProfileForm();
            when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
            when(userProfileService.createProfile(user, form)).thenReturn(createProfileDto());

            ResponseEntity<?> response = userProfileController.createMyProfile(jwt, form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.CREATED);
        }

        @Test
        @DisplayName("JWTがnullの場合401を返す")
        void returns401WhenJwtNull() {
            ResponseEntity<?> response = userProfileController.createMyProfile(null, new UserProfileForm());

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.UNAUTHORIZED);
        }

        @Test
        @DisplayName("既に存在する場合400を返す")
        void returns400WhenAlreadyExists() {
            Jwt jwt = mockJwt("sub-123");
            User user = createUser(10);
            UserProfileForm form = new UserProfileForm();
            when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
            when(userProfileService.createProfile(user, form))
                    .thenThrow(new RuntimeException("プロファイルは既に存在します。"));

            ResponseEntity<?> response = userProfileController.createMyProfile(jwt, form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.BAD_REQUEST);
        }
    }

    @Nested
    @DisplayName("updateMyProfile")
    class UpdateMyProfile {

        @Test
        @DisplayName("正常にプロファイルを更新できる")
        void updatesProfile() {
            Jwt jwt = mockJwt("sub-123");
            User user = createUser(10);
            UserProfileForm form = new UserProfileForm();
            when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
            when(userProfileService.updateProfile(10, form)).thenReturn(createProfileDto());

            ResponseEntity<?> response = userProfileController.updateMyProfile(jwt, form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        }

        @Test
        @DisplayName("存在しない場合400を返す")
        void returns400WhenNotFound() {
            Jwt jwt = mockJwt("sub-123");
            User user = createUser(10);
            UserProfileForm form = new UserProfileForm();
            when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
            when(userProfileService.updateProfile(10, form))
                    .thenThrow(new RuntimeException("プロファイルが見つかりません。"));

            ResponseEntity<?> response = userProfileController.updateMyProfile(jwt, form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.BAD_REQUEST);
        }
    }

    @Nested
    @DisplayName("upsertMyProfile")
    class UpsertMyProfile {

        @Test
        @DisplayName("正常にupsertできる")
        void upsertsProfile() {
            Jwt jwt = mockJwt("sub-123");
            User user = createUser(10);
            UserProfileForm form = new UserProfileForm();
            when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
            when(userProfileService.createOrUpdateProfile(user, form)).thenReturn(createProfileDto());

            ResponseEntity<?> response = userProfileController.upsertMyProfile(jwt, form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        }
    }

    @Nested
    @DisplayName("deleteMyProfile")
    class DeleteMyProfile {

        @Test
        @DisplayName("正常にプロファイルを削除できる")
        void deletesProfile() {
            Jwt jwt = mockJwt("sub-123");
            User user = createUser(10);
            when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);

            ResponseEntity<?> response = userProfileController.deleteMyProfile(jwt);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            verify(userProfileService).deleteProfile(10);
        }

        @Test
        @DisplayName("存在しない場合400を返す")
        void returns400WhenNotFound() {
            Jwt jwt = mockJwt("sub-123");
            User user = createUser(10);
            when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
            doThrow(new RuntimeException("プロファイルが見つかりません。"))
                    .when(userProfileService).deleteProfile(10);

            ResponseEntity<?> response = userProfileController.deleteMyProfile(jwt);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.BAD_REQUEST);
        }
    }
}
