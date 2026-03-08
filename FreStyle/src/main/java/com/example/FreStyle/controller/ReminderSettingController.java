package com.example.FreStyle.controller;

import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.*;
import com.example.FreStyle.dto.ReminderSettingDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.form.ReminderSettingForm;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.GetReminderSettingUseCase;
import com.example.FreStyle.usecase.SaveReminderSettingUseCase;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/reminder")
@Slf4j
public class ReminderSettingController {
    private final UserIdentityService userIdentityService;
    private final GetReminderSettingUseCase getReminderSettingUseCase;
    private final SaveReminderSettingUseCase saveReminderSettingUseCase;

    @GetMapping
    public ResponseEntity<ReminderSettingDto> getSetting(@AuthenticationPrincipal Jwt jwt) {
        log.info("========== GET /api/reminder ==========");
        User user = userIdentityService.findUserBySub(jwt.getSubject());
        ReminderSettingDto setting = getReminderSettingUseCase.execute(user.getId());
        return ResponseEntity.ok(setting);
    }

    @PutMapping
    public ResponseEntity<ReminderSettingDto> saveSetting(
            @AuthenticationPrincipal Jwt jwt,
            @Valid @RequestBody ReminderSettingForm form
    ) {
        log.info("========== PUT /api/reminder ==========");
        User user = userIdentityService.findUserBySub(jwt.getSubject());
        ReminderSettingDto setting = saveReminderSettingUseCase.execute(user.getId(), form);
        log.info("リマインダー設定保存成功 - userId: {}", user.getId());
        return ResponseEntity.ok(setting);
    }
}
