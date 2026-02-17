package com.example.FreStyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
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
import com.example.FreStyle.form.UserProfileForm;
import com.example.FreStyle.usecase.CreateUserProfileUseCase;
import com.example.FreStyle.usecase.DeleteUserProfileUseCase;
import com.example.FreStyle.usecase.GetUserProfileUseCase;
import com.example.FreStyle.usecase.UpdateUserProfileUseCase;
import com.example.FreStyle.usecase.UpsertUserProfileUseCase;

@ExtendWith(MockitoExtension.class)
@DisplayName("UserProfileController")
class UserProfileControllerTest {

    @Mock
    private GetUserProfileUseCase getUserProfileUseCase;

    @Mock
    private CreateUserProfileUseCase createUserProfileUseCase;

    @Mock
    private UpdateUserProfileUseCase updateUserProfileUseCase;

    @Mock
    private UpsertUserProfileUseCase upsertUserProfileUseCase;

    @Mock
    private DeleteUserProfileUseCase deleteUserProfileUseCase;

    @InjectMocks
    private UserProfileController userProfileController;

    private Jwt mockJwt(String sub) {
        Jwt jwt = mock(Jwt.class);
        when(jwt.getSubject()).thenReturn(sub);
        return jwt;
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
            when(getUserProfileUseCase.execute("sub-123")).thenReturn(createProfileDto());

            ResponseEntity<?> response = userProfileController.getMyProfile(jwt);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(response.getBody()).isInstanceOf(UserProfileDto.class);
        }

        @Test
        @DisplayName("プロファイル未作成の場合メッセージを返す")
        void returnsMessageWhenNoProfile() {
            Jwt jwt = mockJwt("sub-123");
            when(getUserProfileUseCase.execute("sub-123")).thenReturn(null);

            ResponseEntity<?> response = userProfileController.getMyProfile(jwt);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        }
    }

    @Nested
    @DisplayName("createMyProfile")
    class CreateMyProfile {

        @Test
        @DisplayName("正常にプロファイルを作成できる")
        void createsProfile() {
            Jwt jwt = mockJwt("sub-123");
            UserProfileForm form = new UserProfileForm();
            when(createUserProfileUseCase.execute("sub-123", form)).thenReturn(createProfileDto());

            ResponseEntity<?> response = userProfileController.createMyProfile(jwt, form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.CREATED);
        }

        @Test
        @DisplayName("既に存在する場合は例外がスローされる")
        void throwsWhenAlreadyExists() {
            Jwt jwt = mockJwt("sub-123");
            UserProfileForm form = new UserProfileForm();
            when(createUserProfileUseCase.execute("sub-123", form))
                    .thenThrow(new RuntimeException("プロファイルは既に存在します。"));

            assertThatThrownBy(() -> userProfileController.createMyProfile(jwt, form))
                    .isInstanceOf(RuntimeException.class);
        }
    }

    @Nested
    @DisplayName("updateMyProfile")
    class UpdateMyProfile {

        @Test
        @DisplayName("正常にプロファイルを更新できる")
        void updatesProfile() {
            Jwt jwt = mockJwt("sub-123");
            UserProfileForm form = new UserProfileForm();
            when(updateUserProfileUseCase.execute("sub-123", form)).thenReturn(createProfileDto());

            ResponseEntity<?> response = userProfileController.updateMyProfile(jwt, form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        }

        @Test
        @DisplayName("存在しない場合は例外がスローされる")
        void throwsWhenNotFound() {
            Jwt jwt = mockJwt("sub-123");
            UserProfileForm form = new UserProfileForm();
            when(updateUserProfileUseCase.execute("sub-123", form))
                    .thenThrow(new RuntimeException("プロファイルが見つかりません。"));

            assertThatThrownBy(() -> userProfileController.updateMyProfile(jwt, form))
                    .isInstanceOf(RuntimeException.class);
        }
    }

    @Nested
    @DisplayName("upsertMyProfile")
    class UpsertMyProfile {

        @Test
        @DisplayName("正常にupsertできる")
        void upsertsProfile() {
            Jwt jwt = mockJwt("sub-123");
            UserProfileForm form = new UserProfileForm();
            when(upsertUserProfileUseCase.execute("sub-123", form)).thenReturn(createProfileDto());

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

            ResponseEntity<?> response = userProfileController.deleteMyProfile(jwt);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            verify(deleteUserProfileUseCase).execute("sub-123");
        }

        @Test
        @DisplayName("存在しない場合は例外がスローされる")
        void throwsWhenNotFound() {
            Jwt jwt = mockJwt("sub-123");
            doThrow(new RuntimeException("プロファイルが見つかりません。"))
                    .when(deleteUserProfileUseCase).execute("sub-123");

            assertThatThrownBy(() -> userProfileController.deleteMyProfile(jwt))
                    .isInstanceOf(RuntimeException.class);
        }
    }
}
