package com.example.FreStyle.usecase;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

import java.util.Optional;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.ReminderSettingDto;
import com.example.FreStyle.entity.ReminderSetting;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.form.ReminderSettingForm;
import com.example.FreStyle.repository.ReminderSettingRepository;
import com.example.FreStyle.repository.UserRepository;

@ExtendWith(MockitoExtension.class)
class SaveReminderSettingUseCaseTest {

    @Mock
    private ReminderSettingRepository reminderSettingRepository;

    @Mock
    private UserRepository userRepository;

    @InjectMocks
    private SaveReminderSettingUseCase useCase;

    @Test
    @DisplayName("既存設定を更新する")
    void execute_updatesExistingSetting() {
        User user = new User();
        user.setId(1);
        when(userRepository.findById(1)).thenReturn(Optional.of(user));

        ReminderSetting existing = new ReminderSetting();
        existing.setUser(user);
        existing.setEnabled(true);
        existing.setReminderTime("20:00");
        existing.setDaysOfWeek("mon,tue,wed,thu,fri");
        when(reminderSettingRepository.findByUserId(1)).thenReturn(Optional.of(existing));

        ReminderSettingForm form = new ReminderSettingForm(false, "18:30", "mon,wed,fri");

        ReminderSettingDto result = useCase.execute(1, form);

        verify(reminderSettingRepository).save(argThat(s ->
                !s.getEnabled() &&
                s.getReminderTime().equals("18:30") &&
                s.getDaysOfWeek().equals("mon,wed,fri")));
        assertEquals(false, result.enabled());
        assertEquals("18:30", result.reminderTime());
        assertEquals("mon,wed,fri", result.daysOfWeek());
    }

    @Test
    @DisplayName("新規設定を作成する")
    void execute_createsNewSetting() {
        User user = new User();
        user.setId(1);
        when(userRepository.findById(1)).thenReturn(Optional.of(user));
        when(reminderSettingRepository.findByUserId(1)).thenReturn(Optional.empty());

        ReminderSettingForm form = new ReminderSettingForm(true, "21:00", "sat,sun");

        ReminderSettingDto result = useCase.execute(1, form);

        verify(reminderSettingRepository).save(argThat(s ->
                s.getUser().equals(user) &&
                s.getEnabled() &&
                s.getReminderTime().equals("21:00") &&
                s.getDaysOfWeek().equals("sat,sun")));
        assertEquals(true, result.enabled());
        assertEquals("21:00", result.reminderTime());
        assertEquals("sat,sun", result.daysOfWeek());
    }
}
