package com.example.FreStyle.usecase;

import static org.mockito.Mockito.*;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.repository.ScenarioBookmarkRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("RemoveScenarioBookmarkUseCase テスト")
class RemoveScenarioBookmarkUseCaseTest {

    @Mock
    private ScenarioBookmarkRepository scenarioBookmarkRepository;

    @InjectMocks
    private RemoveScenarioBookmarkUseCase removeScenarioBookmarkUseCase;

    @Test
    @DisplayName("ブックマークを削除できる")
    void execute_RemovesBookmark() {
        removeScenarioBookmarkUseCase.execute(1, 1);

        verify(scenarioBookmarkRepository).deleteByUserIdAndScenarioId(1, 1);
    }

    @Test
    @DisplayName("存在しないブックマークの削除も正常に完了する")
    void execute_DoesNotThrow_WhenBookmarkNotExists() {
        doNothing().when(scenarioBookmarkRepository).deleteByUserIdAndScenarioId(1, 999);

        removeScenarioBookmarkUseCase.execute(1, 999);

        verify(scenarioBookmarkRepository).deleteByUserIdAndScenarioId(1, 999);
    }
}
