package com.example.FreStyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
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

import com.example.FreStyle.dto.ScoreGoalDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.GetScoreGoalUseCase;
import com.example.FreStyle.usecase.SaveScoreGoalUseCase;

@ExtendWith(MockitoExtension.class)
@DisplayName("ScoreGoalController")
class ScoreGoalControllerTest {

    @Mock
    private GetScoreGoalUseCase getScoreGoalUseCase;

    @Mock
    private SaveScoreGoalUseCase saveScoreGoalUseCase;

    @Mock
    private UserIdentityService userIdentityService;

    @InjectMocks
    private ScoreGoalController scoreGoalController;

    private Jwt mockJwt(String sub) {
        Jwt jwt = mock(Jwt.class);
        when(jwt.getSubject()).thenReturn(sub);
        return jwt;
    }

    @Test
    @DisplayName("GET: 目標スコアを取得する")
    void getGoal_returnsGoal() {
        Jwt jwt = mockJwt("sub-123");
        User user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
        when(getScoreGoalUseCase.execute(1)).thenReturn(new ScoreGoalDto(8.5));

        ResponseEntity<ScoreGoalDto> response = scoreGoalController.getGoal(jwt);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        assertThat(response.getBody().getGoalScore()).isEqualTo(8.5);
    }

    @Test
    @DisplayName("GET: 未設定の場合は204を返す")
    void getGoal_returnsNoContent() {
        Jwt jwt = mockJwt("sub-123");
        User user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
        when(getScoreGoalUseCase.execute(1)).thenReturn(null);

        ResponseEntity<ScoreGoalDto> response = scoreGoalController.getGoal(jwt);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.NO_CONTENT);
    }

    @Test
    @DisplayName("PUT: 目標スコアを保存する")
    void saveGoal_saves() {
        Jwt jwt = mockJwt("sub-123");
        User user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);

        ResponseEntity<Void> response = scoreGoalController.saveGoal(jwt, new ScoreGoalController.SaveGoalRequest(9.0));

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.NO_CONTENT);
        verify(saveScoreGoalUseCase).execute(user, 9.0);
    }
}
