package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.List;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.FilteredScenariosDto;
import com.example.FreStyle.entity.PracticeScenario;
import com.example.FreStyle.mapper.PracticeScenarioMapper;
import com.example.FreStyle.repository.PracticeScenarioRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("FilterPracticeScenariosUseCase テスト")
class FilterPracticeScenariosUseCaseTest {

    @Mock
    private PracticeScenarioRepository practiceScenarioRepository;

    @Mock
    private PracticeScenarioMapper mapper;

    @InjectMocks
    private FilterPracticeScenariosUseCase useCase;

    private PracticeScenario scenario(String name, String difficulty, String category) {
        PracticeScenario s = new PracticeScenario();
        s.setName(name);
        s.setDifficulty(difficulty);
        s.setCategory(category);
        return s;
    }

    @Test
    @DisplayName("難易度でフィルタリングする")
    void filterByDifficulty() {
        PracticeScenario s1 = scenario("初級会議", "easy", "meeting");
        when(practiceScenarioRepository.findByDifficulty("easy")).thenReturn(List.of(s1));
        when(practiceScenarioRepository.findAll()).thenReturn(List.of(
                s1, scenario("上級会議", "hard", "meeting")));

        FilteredScenariosDto result = useCase.execute("easy", null);

        assertThat(result.totalCount()).isEqualTo(1);
        assertThat(result.availableDifficulties()).containsExactlyInAnyOrder("easy", "hard");
        verify(practiceScenarioRepository).findByDifficulty("easy");
    }

    @Test
    @DisplayName("カテゴリでフィルタリングする")
    void filterByCategory() {
        PracticeScenario s1 = scenario("会議シナリオ", "easy", "meeting");
        when(practiceScenarioRepository.findByCategory("meeting")).thenReturn(List.of(s1));
        when(practiceScenarioRepository.findAll()).thenReturn(List.of(
                s1, scenario("メールシナリオ", "easy", "email")));

        FilteredScenariosDto result = useCase.execute(null, "meeting");

        assertThat(result.totalCount()).isEqualTo(1);
        assertThat(result.availableCategories()).containsExactlyInAnyOrder("meeting", "email");
        verify(practiceScenarioRepository).findByCategory("meeting");
    }

    @Test
    @DisplayName("フィルタなしの場合は全件返却する")
    void noFilterReturnsAll() {
        List<PracticeScenario> all = List.of(
                scenario("A", "easy", "meeting"),
                scenario("B", "hard", "email"));
        when(practiceScenarioRepository.findAll()).thenReturn(all);

        FilteredScenariosDto result = useCase.execute(null, null);

        assertThat(result.totalCount()).isEqualTo(2);
        assertThat(result.availableDifficulties()).containsExactlyInAnyOrder("easy", "hard");
        assertThat(result.availableCategories()).containsExactlyInAnyOrder("meeting", "email");
    }

    @Test
    @DisplayName("難易度とカテゴリの両方でフィルタリングする")
    void filterByBoth() {
        PracticeScenario s1 = scenario("初級会議", "easy", "meeting");
        PracticeScenario s2 = scenario("初級メール", "easy", "email");
        when(practiceScenarioRepository.findByDifficulty("easy")).thenReturn(List.of(s1, s2));
        when(practiceScenarioRepository.findAll()).thenReturn(List.of(s1, s2));

        FilteredScenariosDto result = useCase.execute("easy", "meeting");

        assertThat(result.totalCount()).isEqualTo(1);
        assertThat(result.scenarios()).hasSize(1);
    }
}
