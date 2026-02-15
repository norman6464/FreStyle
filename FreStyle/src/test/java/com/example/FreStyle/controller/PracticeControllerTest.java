package com.example.FreStyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.List;

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

import com.example.FreStyle.dto.AiChatSessionDto;
import com.example.FreStyle.dto.PracticeScenarioDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.CreatePracticeSessionUseCase;
import com.example.FreStyle.usecase.GetAllPracticeScenariosUseCase;
import com.example.FreStyle.usecase.GetPracticeScenarioByIdUseCase;

@ExtendWith(MockitoExtension.class)
@DisplayName("PracticeController")
class PracticeControllerTest {

    @Mock
    private GetAllPracticeScenariosUseCase getAllPracticeScenariosUseCase;

    @Mock
    private GetPracticeScenarioByIdUseCase getPracticeScenarioByIdUseCase;

    @Mock
    private CreatePracticeSessionUseCase createPracticeSessionUseCase;

    @Mock
    private UserIdentityService userIdentityService;

    @InjectMocks
    private PracticeController controller;

    private Jwt jwt;
    private User user;

    @BeforeEach
    void setUp() {
        jwt = mock(Jwt.class);
        when(jwt.getSubject()).thenReturn("test-sub");

        user = new User();
        user.setId(1);
        user.setName("テストユーザー");

        when(userIdentityService.findUserBySub("test-sub")).thenReturn(user);
    }

    @Nested
    @DisplayName("GET /api/practice/scenarios - シナリオ一覧取得")
    class GetScenarios {

        @Test
        @DisplayName("シナリオ一覧を返す")
        void shouldReturnScenarios() {
            PracticeScenarioDto s1 = new PracticeScenarioDto();
            s1.setId(1);
            s1.setName("障害報告");
            PracticeScenarioDto s2 = new PracticeScenarioDto();
            s2.setId(2);
            s2.setName("設計レビュー");
            when(getAllPracticeScenariosUseCase.execute()).thenReturn(List.of(s1, s2));

            ResponseEntity<List<PracticeScenarioDto>> response = controller.getScenarios(jwt);

            assertThat(response.getStatusCode().value()).isEqualTo(200);
            assertThat(response.getBody()).hasSize(2);
            verify(getAllPracticeScenariosUseCase).execute();
        }

        @Test
        @DisplayName("空のリストを返す")
        void shouldReturnEmptyList() {
            when(getAllPracticeScenariosUseCase.execute()).thenReturn(List.of());

            ResponseEntity<List<PracticeScenarioDto>> response = controller.getScenarios(jwt);

            assertThat(response.getStatusCode().value()).isEqualTo(200);
            assertThat(response.getBody()).isEmpty();
        }
    }

    @Nested
    @DisplayName("GET /api/practice/scenarios/{scenarioId} - シナリオ詳細取得")
    class GetScenario {

        @Test
        @DisplayName("指定IDのシナリオを返す")
        void shouldReturnScenarioById() {
            PracticeScenarioDto scenario = new PracticeScenarioDto();
            scenario.setId(5);
            scenario.setName("要件変更説明");
            scenario.setCategory("顧客折衝");
            when(getPracticeScenarioByIdUseCase.execute(5)).thenReturn(scenario);

            ResponseEntity<PracticeScenarioDto> response = controller.getScenario(jwt, 5);

            assertThat(response.getStatusCode().value()).isEqualTo(200);
            assertThat(response.getBody()).isNotNull();
            assertThat(response.getBody().getName()).isEqualTo("要件変更説明");
            verify(getPracticeScenarioByIdUseCase).execute(5);
        }
    }

    @Nested
    @DisplayName("POST /api/practice/sessions - 練習セッション作成")
    class CreatePracticeSession {

        @Test
        @DisplayName("練習セッションを作成する")
        void shouldCreatePracticeSession() {
            AiChatSessionDto sessionDto = new AiChatSessionDto();
            sessionDto.setId(10);
            sessionDto.setSessionType("practice");
            when(createPracticeSessionUseCase.execute(user, 3)).thenReturn(sessionDto);

            // リフレクションでinnerレコードにアクセスする代わりに直接コントローラーメソッドをテスト
            // CreatePracticeSessionRequestはpackage-privateなので、テストから直接アクセス可能
            PracticeController.CreatePracticeSessionRequest request =
                    new PracticeController.CreatePracticeSessionRequest(3);

            ResponseEntity<AiChatSessionDto> response = controller.createPracticeSession(jwt, request);

            assertThat(response.getStatusCode().value()).isEqualTo(200);
            assertThat(response.getBody()).isNotNull();
            assertThat(response.getBody().getId()).isEqualTo(10);
            verify(createPracticeSessionUseCase).execute(user, 3);
        }
    }
}
