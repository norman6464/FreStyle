package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import com.example.FreStyle.dto.ReminderSettingDto;
import com.example.FreStyle.repository.ReminderSettingRepository;
import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
@Transactional(readOnly = true)
public class GetReminderSettingUseCase {
    private final ReminderSettingRepository repository;

    public ReminderSettingDto execute(Integer userId) {
        return repository.findByUserId(userId)
            .map(s -> new ReminderSettingDto(s.getEnabled(), s.getReminderTime(), s.getDaysOfWeek()))
            .orElse(new ReminderSettingDto(true, "20:00", "mon,tue,wed,thu,fri"));
    }
}
