package com.example.FreStyle.controller;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.oauth2.jwt.Jwt;

import com.example.FreStyle.dto.WeeklyChallengeDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.GetCurrentChallengeUseCase;
import com.example.FreStyle.usecase.IncrementChallengeProgressUseCase;

@ExtendWith(MockitoExtension.class)
@DisplayName("WeeklyChallengeController テスト")
class WeeklyChallengeControllerTest {

    @Mock
    private GetCurrentChallengeUseCase getCurrentChallengeUseCase;

    @Mock
    private IncrementChallengeProgressUseCase incrementChallengeProgressUseCase;

    @Mock
    private UserIdentityService userIdentityService;

    @InjectMocks
    private WeeklyChallengeController weeklyChallengeController;

    private Jwt createMockJwt() {
        Jwt jwt = mock(Jwt.class);
        when(jwt.getSubject()).thenReturn("cognito-sub-123");
        return jwt;
    }

    @Test
    @DisplayName("getCurrentChallenge: 今週のチャレンジとユーザー進捗を取得できる")
    void getCurrentChallenge_returnsChallengeWithProgress() {
        Jwt jwt = createMockJwt();
        User user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("cognito-sub-123")).thenReturn(user);
        WeeklyChallengeDto dto = new WeeklyChallengeDto(
                1, "今週のチャレンジ", "説明文", "communication",
                3, 1, false, "2026-03-02", "2026-03-08");
        when(getCurrentChallengeUseCase.execute(1)).thenReturn(dto);

        ResponseEntity<WeeklyChallengeDto> response = weeklyChallengeController.getCurrentChallenge(jwt);

        assertEquals(HttpStatus.OK, response.getStatusCode());
        assertEquals("今週のチャレンジ", response.getBody().title());
        assertEquals(1, response.getBody().completedSessions());
        assertFalse(response.getBody().isCompleted());
    }

    @Test
    @DisplayName("incrementProgress: チャレンジ進捗をインクリメントできる")
    void incrementProgress_incrementsAndReturnsUpdated() {
        Jwt jwt = createMockJwt();
        User user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("cognito-sub-123")).thenReturn(user);
        WeeklyChallengeDto dto = new WeeklyChallengeDto(
                1, "今週のチャレンジ", "説明文", "communication",
                3, 2, false, "2026-03-02", "2026-03-08");
        when(incrementChallengeProgressUseCase.execute(1)).thenReturn(dto);

        ResponseEntity<WeeklyChallengeDto> response = weeklyChallengeController.incrementProgress(jwt);

        assertEquals(HttpStatus.OK, response.getStatusCode());
        assertEquals(2, response.getBody().completedSessions());
        verify(incrementChallengeProgressUseCase).execute(1);
    }
}
