package com.example.FreStyle.controller;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

import java.time.LocalDate;
import java.util.Map;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.oauth2.jwt.Jwt;

import com.example.FreStyle.dto.DailyGoalDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.GetTodayDailyGoalUseCase;
import com.example.FreStyle.usecase.IncrementDailyGoalUseCase;
import com.example.FreStyle.usecase.SetDailyGoalTargetUseCase;

@ExtendWith(MockitoExtension.class)
@DisplayName("DailyGoalController テスト")
class DailyGoalControllerTest {

    @Mock
    private GetTodayDailyGoalUseCase getTodayDailyGoalUseCase;

    @Mock
    private SetDailyGoalTargetUseCase setDailyGoalTargetUseCase;

    @Mock
    private IncrementDailyGoalUseCase incrementDailyGoalUseCase;

    @Mock
    private UserIdentityService userIdentityService;

    @InjectMocks
    private DailyGoalController dailyGoalController;

    private Jwt createMockJwt() {
        Jwt jwt = mock(Jwt.class);
        when(jwt.getSubject()).thenReturn("cognito-sub-123");
        return jwt;
    }

    @Test
    @DisplayName("getToday: 今日のゴールを取得できる")
    void getToday_returnsGoal() {
        Jwt jwt = createMockJwt();
        User user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("cognito-sub-123")).thenReturn(user);
        DailyGoalDto dto = new DailyGoalDto(LocalDate.now().toString(), 5, 2);
        when(getTodayDailyGoalUseCase.execute(1)).thenReturn(dto);

        ResponseEntity<DailyGoalDto> response = dailyGoalController.getToday(jwt);

        assertEquals(HttpStatus.OK, response.getStatusCode());
        assertEquals(5, response.getBody().getTarget());
    }

    @Test
    @DisplayName("setTarget: 目標回数を設定できる")
    void setTarget_setsTarget() {
        Jwt jwt = createMockJwt();
        User user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("cognito-sub-123")).thenReturn(user);

        ResponseEntity<Void> response = dailyGoalController.setTarget(jwt, Map.of("target", 7));

        assertEquals(HttpStatus.OK, response.getStatusCode());
        verify(setDailyGoalTargetUseCase).execute(user, 7);
    }

    @Test
    @DisplayName("increment: 完了数をインクリメントできる")
    void increment_incrementsCompleted() {
        Jwt jwt = createMockJwt();
        User user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("cognito-sub-123")).thenReturn(user);
        DailyGoalDto dto = new DailyGoalDto(LocalDate.now().toString(), 3, 1);
        when(incrementDailyGoalUseCase.execute(user)).thenReturn(dto);

        ResponseEntity<DailyGoalDto> response = dailyGoalController.increment(jwt);

        assertEquals(HttpStatus.OK, response.getStatusCode());
        assertEquals(1, response.getBody().getCompleted());
    }
}
