package com.example.FreStyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.ResponseEntity;
import org.springframework.security.oauth2.jwt.Jwt;

import com.example.FreStyle.dto.ReminderSettingDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.form.ReminderSettingForm;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.GetReminderSettingUseCase;
import com.example.FreStyle.usecase.SaveReminderSettingUseCase;

@ExtendWith(MockitoExtension.class)
@DisplayName("ReminderSettingController")
class ReminderSettingControllerTest {

    @Mock
    private UserIdentityService userIdentityService;

    @Mock
    private GetReminderSettingUseCase getReminderSettingUseCase;

    @Mock
    private SaveReminderSettingUseCase saveReminderSettingUseCase;

    @InjectMocks
    private ReminderSettingController controller;

    private Jwt jwt;
    private User user;

    @BeforeEach
    void setUp() {
        jwt = mock(Jwt.class);
        when(jwt.getSubject()).thenReturn("test-sub");

        user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("test-sub")).thenReturn(user);
    }

    @Nested
    @DisplayName("GET /api/reminder")
    class GetSetting {

        @Test
        @DisplayName("設定を取得する")
        void shouldReturnSetting() {
            ReminderSettingDto dto = new ReminderSettingDto(true, "20:00", "mon,tue,wed,thu,fri");
            when(getReminderSettingUseCase.execute(1)).thenReturn(dto);

            ResponseEntity<ReminderSettingDto> response = controller.getSetting(jwt);

            assertThat(response.getStatusCode().value()).isEqualTo(200);
            assertThat(response.getBody()).isNotNull();
            assertThat(response.getBody().enabled()).isTrue();
            assertThat(response.getBody().reminderTime()).isEqualTo("20:00");
            assertThat(response.getBody().daysOfWeek()).isEqualTo("mon,tue,wed,thu,fri");
        }

        @Test
        @DisplayName("UseCaseにユーザーIDを渡す")
        void shouldPassUserIdToUseCase() {
            when(getReminderSettingUseCase.execute(1)).thenReturn(
                    new ReminderSettingDto(true, "20:00", "mon,tue,wed,thu,fri"));

            controller.getSetting(jwt);

            verify(getReminderSettingUseCase).execute(1);
        }
    }

    @Nested
    @DisplayName("PUT /api/reminder")
    class SaveSetting {

        @Test
        @DisplayName("設定を保存する")
        void shouldSaveSetting() {
            ReminderSettingForm form = new ReminderSettingForm(false, "18:30", "mon,wed,fri");
            ReminderSettingDto dto = new ReminderSettingDto(false, "18:30", "mon,wed,fri");
            when(saveReminderSettingUseCase.execute(1, form)).thenReturn(dto);

            ResponseEntity<ReminderSettingDto> response = controller.saveSetting(jwt, form);

            assertThat(response.getStatusCode().value()).isEqualTo(200);
            assertThat(response.getBody()).isNotNull();
            assertThat(response.getBody().enabled()).isFalse();
            assertThat(response.getBody().reminderTime()).isEqualTo("18:30");
        }
    }
}
