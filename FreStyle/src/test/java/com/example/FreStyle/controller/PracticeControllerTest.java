package com.example.FreStyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.List;
import java.util.Set;

import jakarta.validation.ConstraintViolation;
import jakarta.validation.Validation;
import jakarta.validation.Validator;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.mockito.junit.jupiter.MockitoSettings;
import org.mockito.quality.Strictness;
import org.springframework.http.ResponseEntity;
import org.springframework.security.oauth2.jwt.Jwt;

import com.example.FreStyle.dto.AiChatSessionDto;
import com.example.FreStyle.dto.PracticeScenarioDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.CreatePracticeSessionUseCase;
import com.example.FreStyle.usecase.GetAllPracticeScenariosUseCase;
import com.example.FreStyle.usecase.GetPracticeScenarioByIdUseCase;
import com.example.FreStyle.usecase.GetRecommendedScenariosUseCase;
import com.example.FreStyle.dto.FilteredScenariosDto;
import com.example.FreStyle.dto.RecommendedScenarioDto;
import com.example.FreStyle.dto.RecommendedScenarioDto.ScenarioRecommendation;
import com.example.FreStyle.usecase.FilterPracticeScenariosUseCase;

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
    private GetRecommendedScenariosUseCase getRecommendedScenariosUseCase;

    @Mock
    private FilterPracticeScenariosUseCase filterPracticeScenariosUseCase;

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
            PracticeScenarioDto s1 = new PracticeScenarioDto(1, "障害報告", null, null, null, null, null);
            PracticeScenarioDto s2 = new PracticeScenarioDto(2, "設計レビュー", null, null, null, null, null);
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
            PracticeScenarioDto scenario = new PracticeScenarioDto(5, "要件変更説明", null, "顧客折衝", null, null, null);
            when(getPracticeScenarioByIdUseCase.execute(5)).thenReturn(scenario);

            ResponseEntity<PracticeScenarioDto> response = controller.getScenario(jwt, 5);

            assertThat(response.getStatusCode().value()).isEqualTo(200);
            assertThat(response.getBody()).isNotNull();
            assertThat(response.getBody().name()).isEqualTo("要件変更説明");
            verify(getPracticeScenarioByIdUseCase).execute(5);
        }
    }

    @Nested
    @DisplayName("POST /api/practice/sessions - 練習セッション作成")
    class CreatePracticeSession {

        @Test
        @DisplayName("練習セッションを作成する")
        void shouldCreatePracticeSession() {
            AiChatSessionDto sessionDto = new AiChatSessionDto(10, null, null, null, null, "practice", null, null, null);
            when(createPracticeSessionUseCase.execute(user, 3)).thenReturn(sessionDto);

            // リフレクションでinnerレコードにアクセスする代わりに直接コントローラーメソッドをテスト
            // CreatePracticeSessionRequestはpackage-privateなので、テストから直接アクセス可能
            PracticeController.CreatePracticeSessionRequest request =
                    new PracticeController.CreatePracticeSessionRequest(3);

            ResponseEntity<AiChatSessionDto> response = controller.createPracticeSession(jwt, request);

            assertThat(response.getStatusCode().value()).isEqualTo(200);
            assertThat(response.getBody()).isNotNull();
            assertThat(response.getBody().id()).isEqualTo(10);
            verify(createPracticeSessionUseCase).execute(user, 3);
        }
    }

    @Nested
    @DisplayName("GET /api/practice/scenarios/recommended - 推奨シナリオ取得")
    class GetRecommendedScenarios {

        @Test
        @DisplayName("推奨シナリオ一覧を返す")
        void shouldReturnRecommendedScenarios() {
            ScenarioRecommendation rec = new ScenarioRecommendation(
                    10, "交渉シナリオ", "business", "hard", 4.5, 2);
            RecommendedScenarioDto dto = new RecommendedScenarioDto(List.of(rec));
            when(getRecommendedScenariosUseCase.execute(1)).thenReturn(dto);

            ResponseEntity<RecommendedScenarioDto> response = controller.getRecommendedScenarios(jwt);

            assertThat(response.getStatusCode().value()).isEqualTo(200);
            assertThat(response.getBody()).isNotNull();
            assertThat(response.getBody().recommendations()).hasSize(1);
            assertThat(response.getBody().recommendations().get(0).scenarioName()).isEqualTo("交渉シナリオ");
            verify(getRecommendedScenariosUseCase).execute(1);
        }
    }

    @Nested
    @DisplayName("GET /api/practice/scenarios/filter - シナリオフィルタリング")
    class FilterScenarios {

        @Test
        @DisplayName("難易度とカテゴリでフィルタリングされたシナリオを返す")
        void shouldReturnFilteredScenarios() {
            PracticeScenarioDto s1 = new PracticeScenarioDto(1, "障害報告", null, null, "easy", null, null);
            FilteredScenariosDto dto = new FilteredScenariosDto(
                    List.of(s1), 1, List.of("easy", "medium"), List.of("business"));
            when(filterPracticeScenariosUseCase.execute("easy", "business")).thenReturn(dto);

            ResponseEntity<FilteredScenariosDto> response = controller.filterScenarios(jwt, "easy", "business");

            assertThat(response.getStatusCode().value()).isEqualTo(200);
            assertThat(response.getBody()).isNotNull();
            assertThat(response.getBody().totalCount()).isEqualTo(1);
            verify(filterPracticeScenariosUseCase).execute("easy", "business");
        }
    }

    @Nested
    @DisplayName("リクエストバリデーション")
    @MockitoSettings(strictness = Strictness.LENIENT)
    class RequestValidation {

        private final Validator validator = Validation.buildDefaultValidatorFactory().getValidator();

        @Test
        @DisplayName("CreatePracticeSessionRequest: scenarioIdがnullの場合バリデーションエラー")
        void createPracticeSessionRequest_nullScenarioId_fails() {
            var request = new PracticeController.CreatePracticeSessionRequest(null);
            Set<ConstraintViolation<PracticeController.CreatePracticeSessionRequest>> violations = validator.validate(request);
            assertThat(violations).isNotEmpty();
        }

        @Test
        @DisplayName("CreatePracticeSessionRequest: scenarioIdが正常な場合バリデーション成功")
        void createPracticeSessionRequest_valid_passes() {
            var request = new PracticeController.CreatePracticeSessionRequest(1);
            Set<ConstraintViolation<PracticeController.CreatePracticeSessionRequest>> violations = validator.validate(request);
            assertThat(violations).isEmpty();
        }
    }
}
