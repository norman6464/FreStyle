package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.Mockito.*;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.service.UserProfileService;

@ExtendWith(MockitoExtension.class)
@DisplayName("DeleteUserProfileUseCase")
class DeleteUserProfileUseCaseTest {

    @Mock
    private UserProfileService userProfileService;

    @Mock
    private UserIdentityService userIdentityService;

    @InjectMocks
    private DeleteUserProfileUseCase useCase;

    @Test
    @DisplayName("正常にプロファイルを削除する")
    void deletesProfile() {
        User user = new User();
        user.setId(10);
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);

        useCase.execute("sub-123");

        verify(userProfileService).deleteProfile(10);
    }

    @Test
    @DisplayName("存在しない場合はRuntimeExceptionを投げる")
    void throwsWhenNotFound() {
        User user = new User();
        user.setId(10);
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
        doThrow(new RuntimeException("プロファイルが見つかりません。"))
                .when(userProfileService).deleteProfile(10);

        assertThatThrownBy(() -> useCase.execute("sub-123"))
                .isInstanceOf(RuntimeException.class);
    }
}
