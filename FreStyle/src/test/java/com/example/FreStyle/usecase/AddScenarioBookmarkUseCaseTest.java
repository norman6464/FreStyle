package com.example.FreStyle.usecase;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

import java.util.Optional;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.entity.PracticeScenario;
import com.example.FreStyle.entity.ScenarioBookmark;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.repository.PracticeScenarioRepository;
import com.example.FreStyle.repository.ScenarioBookmarkRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("AddScenarioBookmarkUseCase テスト")
class AddScenarioBookmarkUseCaseTest {

    @Mock
    private ScenarioBookmarkRepository scenarioBookmarkRepository;

    @Mock
    private PracticeScenarioRepository practiceScenarioRepository;

    @InjectMocks
    private AddScenarioBookmarkUseCase addScenarioBookmarkUseCase;

    private User testUser;
    private PracticeScenario testScenario;

    @BeforeEach
    void setUp() {
        testUser = new User();
        testUser.setId(1);
        testUser.setName("テストユーザー");
        testUser.setEmail("test@example.com");

        testScenario = new PracticeScenario();
        testScenario.setId(1);
        testScenario.setName("テストシナリオ");
    }

    @Test
    @DisplayName("新規ブックマークを追加できる")
    void execute_AddsNewBookmark() {
        when(scenarioBookmarkRepository.existsByUserIdAndScenarioId(1, 1)).thenReturn(false);
        when(practiceScenarioRepository.findById(1)).thenReturn(Optional.of(testScenario));

        addScenarioBookmarkUseCase.execute(testUser, 1);

        verify(scenarioBookmarkRepository).save(any(ScenarioBookmark.class));
    }

    @Test
    @DisplayName("既にブックマーク済みの場合は重複保存しない")
    void execute_SkipsWhenAlreadyBookmarked() {
        when(scenarioBookmarkRepository.existsByUserIdAndScenarioId(1, 1)).thenReturn(true);

        addScenarioBookmarkUseCase.execute(testUser, 1);

        verify(scenarioBookmarkRepository, never()).save(any(ScenarioBookmark.class));
    }

    @Test
    @DisplayName("存在しないシナリオIDの場合はResourceNotFoundExceptionをスロー")
    void execute_ThrowsException_WhenScenarioNotFound() {
        when(scenarioBookmarkRepository.existsByUserIdAndScenarioId(1, 999)).thenReturn(false);
        when(practiceScenarioRepository.findById(999)).thenReturn(Optional.empty());

        assertThrows(ResourceNotFoundException.class, () -> {
            addScenarioBookmarkUseCase.execute(testUser, 999);
        });

        verify(scenarioBookmarkRepository, never()).save(any(ScenarioBookmark.class));
    }
}
