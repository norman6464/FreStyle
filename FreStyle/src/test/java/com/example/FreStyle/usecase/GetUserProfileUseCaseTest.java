package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
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
}
