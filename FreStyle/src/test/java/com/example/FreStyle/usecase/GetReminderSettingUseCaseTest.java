package com.example.FreStyle.usecase;

import static org.junit.jupiter.api.Assertions.*;
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
import com.example.FreStyle.repository.ReminderSettingRepository;

@ExtendWith(MockitoExtension.class)
class GetReminderSettingUseCaseTest {

    @Mock
    private ReminderSettingRepository repository;

    @InjectMocks
    private GetReminderSettingUseCase useCase;

    @Test
    @DisplayName("設定が存在する場合にDTOを返す")
    void execute_returnsDtoWhenSettingExists() {
        ReminderSetting setting = new ReminderSetting();
        setting.setEnabled(false);
        setting.setReminderTime("18:30");
        setting.setDaysOfWeek("mon,wed,fri");
        when(repository.findByUserId(1)).thenReturn(Optional.of(setting));

        ReminderSettingDto result = useCase.execute(1);

        assertNotNull(result);
        assertEquals(false, result.enabled());
        assertEquals("18:30", result.reminderTime());
        assertEquals("mon,wed,fri", result.daysOfWeek());
    }

    @Test
    @DisplayName("設定が存在しない場合にデフォルト値を返す")
    void execute_returnsDefaultWhenSettingNotFound() {
        when(repository.findByUserId(999)).thenReturn(Optional.empty());

        ReminderSettingDto result = useCase.execute(999);

        assertNotNull(result);
        assertEquals(true, result.enabled());
        assertEquals("20:00", result.reminderTime());
        assertEquals("mon,tue,wed,thu,fri", result.daysOfWeek());
    }
}
