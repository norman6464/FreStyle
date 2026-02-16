package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.repository.ScenarioBookmarkRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class RemoveScenarioBookmarkUseCase {

    private final ScenarioBookmarkRepository scenarioBookmarkRepository;

    @Transactional
    public void execute(Integer userId, Integer scenarioId) {
        scenarioBookmarkRepository.deleteByUserIdAndScenarioId(userId, scenarioId);
    }
}
