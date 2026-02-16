package com.example.FreStyle.usecase;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

import java.util.List;

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
import com.example.FreStyle.repository.ScenarioBookmarkRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetUserBookmarksUseCase テスト")
class GetUserBookmarksUseCaseTest {

    @Mock
    private ScenarioBookmarkRepository scenarioBookmarkRepository;

    @InjectMocks
    private GetUserBookmarksUseCase getUserBookmarksUseCase;

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
    @DisplayName("ユーザーのブックマーク済みシナリオIDリストを取得できる")
    void execute_ReturnsBookmarkedScenarioIds() {
        ScenarioBookmark bookmark1 = new ScenarioBookmark();
        bookmark1.setId(1);
        bookmark1.setUser(testUser);
        PracticeScenario scenario1 = new PracticeScenario();
        scenario1.setId(1);
        bookmark1.setScenario(scenario1);

        ScenarioBookmark bookmark2 = new ScenarioBookmark();
        bookmark2.setId(2);
        bookmark2.setUser(testUser);
        PracticeScenario scenario2 = new PracticeScenario();
        scenario2.setId(3);
        bookmark2.setScenario(scenario2);

        when(scenarioBookmarkRepository.findByUserId(1)).thenReturn(List.of(bookmark1, bookmark2));

        List<Integer> result = getUserBookmarksUseCase.execute(testUser.getId());

        assertEquals(2, result.size());
        assertTrue(result.contains(1));
        assertTrue(result.contains(3));
        verify(scenarioBookmarkRepository).findByUserId(1);
    }

    @Test
    @DisplayName("ブックマークが空の場合、空のリストを返す")
    void execute_ReturnsEmptyList_WhenNoBookmarks() {
        when(scenarioBookmarkRepository.findByUserId(1)).thenReturn(List.of());

        List<Integer> result = getUserBookmarksUseCase.execute(testUser.getId());

        assertTrue(result.isEmpty());
        verify(scenarioBookmarkRepository).findByUserId(1);
    }
}
