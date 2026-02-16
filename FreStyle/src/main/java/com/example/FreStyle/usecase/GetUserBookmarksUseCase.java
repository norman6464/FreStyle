package com.example.FreStyle.usecase;

import java.util.List;
import java.util.stream.Collectors;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.repository.ScenarioBookmarkRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetUserBookmarksUseCase {

    private final ScenarioBookmarkRepository scenarioBookmarkRepository;

    @Transactional(readOnly = true)
    public List<Integer> execute(Integer userId) {
        return scenarioBookmarkRepository.findByUserId(userId).stream()
                .map(bookmark -> bookmark.getScenario().getId())
                .collect(Collectors.toList());
    }
}
