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
import com.example.FreStyle.form.UserProfileForm;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.service.UserProfileService;

@ExtendWith(MockitoExtension.class)
@DisplayName("CreateUserProfileUseCase")
class CreateUserProfileUseCaseTest {

    @Mock
    private UserProfileService userProfileService;

    @Mock
    private UserIdentityService userIdentityService;

    @InjectMocks
    private CreateUserProfileUseCase useCase;

    @Test
    @DisplayName("正常にプロファイルを作成する")
    void createsProfile() {
        User user = new User();
        user.setId(10);
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
        UserProfileForm form = new UserProfileForm();
        UserProfileDto dto = new UserProfileDto(null, 10, "新規", null, null, null, null, null, null);
        when(userProfileService.createProfile(user, form)).thenReturn(dto);

        UserProfileDto result = useCase.execute("sub-123", form);

        assertThat(result.displayName()).isEqualTo("新規");
    }

    @Test
    @DisplayName("既に存在する場合はRuntimeExceptionを投げる")
    void throwsWhenAlreadyExists() {
        User user = new User();
        user.setId(10);
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
        UserProfileForm form = new UserProfileForm();
        when(userProfileService.createProfile(user, form))
                .thenThrow(new RuntimeException("プロファイルは既に存在します。"));

        assertThatThrownBy(() -> useCase.execute("sub-123", form))
                .isInstanceOf(RuntimeException.class);
    }
}
