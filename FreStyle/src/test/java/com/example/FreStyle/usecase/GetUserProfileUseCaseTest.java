package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.Mockito.*;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.UserProfileDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.service.UserProfileService;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetUserProfileUseCase")
class GetUserProfileUseCaseTest {

    @Mock
    private UserProfileService userProfileService;

    @Mock
    private UserIdentityService userIdentityService;

    @InjectMocks
    private GetUserProfileUseCase useCase;

    @Test
    @DisplayName("subからユーザーを特定してプロファイルを取得する")
    void returnsProfileForUser() {
        User user = new User();
        user.setId(10);
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
        UserProfileDto dto = new UserProfileDto(null, 10, "テスト", null, null, null, null, null, null);
        when(userProfileService.getProfileByUserId(10)).thenReturn(dto);

        UserProfileDto result = useCase.execute("sub-123");

        assertThat(result).isNotNull();
        assertThat(result.displayName()).isEqualTo("テスト");
    }

    @Test
    @DisplayName("プロファイル未作成の場合nullを返す")
    void returnsNullWhenNoProfile() {
        User user = new User();
        user.setId(10);
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
        when(userProfileService.getProfileByUserId(10)).thenReturn(null);

        UserProfileDto result = useCase.execute("sub-123");

        assertThat(result).isNull();
    }

    @Test
    @DisplayName("findUserBySubとgetProfileByUserIdが正しく呼び出される")
    void verifiesServiceCalls() {
        User user = new User();
        user.setId(20);
        when(userIdentityService.findUserBySub("sub-456")).thenReturn(user);
        UserProfileDto dto = new UserProfileDto(null, 20, "ユーザー2", null, null, null, null, null, null);
        when(userProfileService.getProfileByUserId(20)).thenReturn(dto);

        useCase.execute("sub-456");

        verify(userIdentityService).findUserBySub("sub-456");
        verify(userProfileService).getProfileByUserId(20);
    }

    @Test
    @DisplayName("userIdentityServiceが例外をスローした場合そのまま伝搬する")
    void propagatesExceptionFromUserIdentityService() {
        when(userIdentityService.findUserBySub("invalid-sub"))
                .thenThrow(new RuntimeException("ユーザーが見つかりません"));

        assertThatThrownBy(() -> useCase.execute("invalid-sub"))
                .isInstanceOf(RuntimeException.class)
                .hasMessage("ユーザーが見つかりません");
    }
}
