package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.List;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.FilteredScenariosDto;
import com.example.FreStyle.dto.PracticeScenarioDto;
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

    private PracticeScenario easyMeeting;
    private PracticeScenario easyEmail;
    private PracticeScenario hardMeeting;

    private PracticeScenarioDto easyMeetingDto;
    private PracticeScenarioDto easyEmailDto;
    private PracticeScenarioDto hardMeetingDto;

    @BeforeEach
    void setUp() {
        easyMeeting = scenario("初級会議", "easy", "meeting");
        easyEmail = scenario("初級メール", "easy", "email");
        hardMeeting = scenario("上級会議", "hard", "meeting");

        easyMeetingDto = new PracticeScenarioDto(1, "初級会議", null, "meeting", null, "easy", null);
        easyEmailDto = new PracticeScenarioDto(2, "初級メール", null, "email", null, "easy", null);
        hardMeetingDto = new PracticeScenarioDto(3, "上級会議", null, "meeting", null, "hard", null);
    }

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
        when(practiceScenarioRepository.findAll()).thenReturn(List.of(easyMeeting, easyEmail, hardMeeting));
        when(mapper.toDto(easyMeeting)).thenReturn(easyMeetingDto);
        when(mapper.toDto(easyEmail)).thenReturn(easyEmailDto);

        FilteredScenariosDto result = useCase.execute("easy", null);

        assertThat(result.totalCount()).isEqualTo(2);
        assertThat(result.scenarios()).containsExactly(easyMeetingDto, easyEmailDto);
        assertThat(result.availableDifficulties()).containsExactlyInAnyOrder("easy", "hard");
        verify(practiceScenarioRepository).findAll();
    }

    @Test
    @DisplayName("カテゴリでフィルタリングする")
    void filterByCategory() {
        when(practiceScenarioRepository.findAll()).thenReturn(List.of(easyMeeting, easyEmail, hardMeeting));
        when(mapper.toDto(easyMeeting)).thenReturn(easyMeetingDto);
        when(mapper.toDto(hardMeeting)).thenReturn(hardMeetingDto);

        FilteredScenariosDto result = useCase.execute(null, "meeting");

        assertThat(result.totalCount()).isEqualTo(2);
        assertThat(result.scenarios()).containsExactly(easyMeetingDto, hardMeetingDto);
        assertThat(result.availableCategories()).containsExactlyInAnyOrder("meeting", "email");
    }

    @Test
    @DisplayName("フィルタなしの場合は全件返却する")
    void noFilterReturnsAll() {
        when(practiceScenarioRepository.findAll()).thenReturn(List.of(easyMeeting, easyEmail, hardMeeting));
        when(mapper.toDto(any(PracticeScenario.class))).thenAnswer(inv -> {
            PracticeScenario s = inv.getArgument(0);
            if (s == easyMeeting) return easyMeetingDto;
            if (s == easyEmail) return easyEmailDto;
            return hardMeetingDto;
        });

        FilteredScenariosDto result = useCase.execute(null, null);

        assertThat(result.totalCount()).isEqualTo(3);
        assertThat(result.availableDifficulties()).containsExactlyInAnyOrder("easy", "hard");
        assertThat(result.availableCategories()).containsExactlyInAnyOrder("meeting", "email");
    }

    @Test
    @DisplayName("難易度とカテゴリの両方でフィルタリングする")
    void filterByBoth() {
        when(practiceScenarioRepository.findAll()).thenReturn(List.of(easyMeeting, easyEmail, hardMeeting));
        when(mapper.toDto(easyMeeting)).thenReturn(easyMeetingDto);

        FilteredScenariosDto result = useCase.execute("easy", "meeting");

        assertThat(result.totalCount()).isEqualTo(1);
        assertThat(result.scenarios()).hasSize(1);
        assertThat(result.scenarios().getFirst().category()).isEqualTo("meeting");
        assertThat(result.scenarios().getFirst().difficulty()).isEqualTo("easy");
    }
}
