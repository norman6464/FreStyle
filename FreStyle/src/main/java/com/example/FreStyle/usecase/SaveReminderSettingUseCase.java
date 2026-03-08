package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import com.example.FreStyle.dto.ReminderSettingDto;
import com.example.FreStyle.entity.ReminderSetting;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.form.ReminderSettingForm;
import com.example.FreStyle.repository.ReminderSettingRepository;
import com.example.FreStyle.repository.UserRepository;
import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class SaveReminderSettingUseCase {
    private final ReminderSettingRepository reminderSettingRepository;
    private final UserRepository userRepository;

    @Transactional
    public ReminderSettingDto execute(Integer userId, ReminderSettingForm form) {
        User user = userRepository.findById(userId)
            .orElseThrow(() -> new IllegalArgumentException("ユーザーが見つかりません"));

        ReminderSetting setting = reminderSettingRepository.findByUserId(userId)
            .orElseGet(() -> {
                ReminderSetting s = new ReminderSetting();
                s.setUser(user);
                return s;
            });

        setting.setEnabled(form.enabled());
        setting.setReminderTime(form.reminderTime());
        setting.setDaysOfWeek(form.daysOfWeek());

        reminderSettingRepository.save(setting);

        return new ReminderSettingDto(setting.getEnabled(), setting.getReminderTime(), setting.getDaysOfWeek());
    }
}
