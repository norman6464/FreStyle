package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.entity.PracticeScenario;
import com.example.FreStyle.entity.ScenarioBookmark;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.repository.PracticeScenarioRepository;
import com.example.FreStyle.repository.ScenarioBookmarkRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class AddScenarioBookmarkUseCase {

    private final ScenarioBookmarkRepository scenarioBookmarkRepository;
    private final PracticeScenarioRepository practiceScenarioRepository;

    @Transactional
    public void execute(User user, Integer scenarioId) {
        if (scenarioBookmarkRepository.existsByUserIdAndScenarioId(user.getId(), scenarioId)) {
            return;
        }

        PracticeScenario scenario = practiceScenarioRepository.findById(scenarioId)
                .orElseThrow(() -> new ResourceNotFoundException("シナリオが見つかりません: ID=" + scenarioId));

        ScenarioBookmark bookmark = new ScenarioBookmark();
        bookmark.setUser(user);
        bookmark.setScenario(scenario);
        scenarioBookmarkRepository.save(bookmark);
    }
}
