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
import com.example.FreStyle.form.UserProfileForm;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.service.UserProfileService;

@ExtendWith(MockitoExtension.class)
@DisplayName("UpsertUserProfileUseCase")
class UpsertUserProfileUseCaseTest {

    @Mock
    private UserProfileService userProfileService;

    @Mock
    private UserIdentityService userIdentityService;

    @InjectMocks
    private UpsertUserProfileUseCase useCase;

    @Test
    @DisplayName("正常にupsertする")
    void upsertsProfile() {
        User user = new User();
        user.setId(10);
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
        UserProfileForm form = new UserProfileForm();
        UserProfileDto dto = new UserProfileDto(null, 10, "upsert済み", null, null, null, null, null, null);
        when(userProfileService.createOrUpdateProfile(user, form)).thenReturn(dto);

        UserProfileDto result = useCase.execute("sub-123", form);

        assertThat(result.displayName()).isEqualTo("upsert済み");
    }

    @Test
    @DisplayName("新規作成時もDTOを返す")
    void returnsProfileOnCreate() {
        User user = new User();
        user.setId(20);
        when(userIdentityService.findUserBySub("sub-new")).thenReturn(user);
        UserProfileForm form = new UserProfileForm();
        UserProfileDto dto = new UserProfileDto(null, 20, null, null, null, null, null, null, null);
        when(userProfileService.createOrUpdateProfile(user, form)).thenReturn(dto);

        UserProfileDto result = useCase.execute("sub-new", form);

        assertThat(result.userId()).isEqualTo(20);
    }
}
