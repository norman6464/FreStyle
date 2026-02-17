package com.example.FreStyle.usecase;

import java.util.List;
import java.util.Objects;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.FilteredScenariosDto;
import com.example.FreStyle.dto.PracticeScenarioDto;
import com.example.FreStyle.entity.PracticeScenario;
import com.example.FreStyle.mapper.PracticeScenarioMapper;
import com.example.FreStyle.repository.PracticeScenarioRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class FilterPracticeScenariosUseCase {

    private final PracticeScenarioRepository practiceScenarioRepository;
    private final PracticeScenarioMapper mapper;

    @Transactional(readOnly = true)
    public FilteredScenariosDto execute(String difficulty, String category) {
        List<PracticeScenario> all = practiceScenarioRepository.findAll();

        List<String> availableDifficulties = all.stream()
                .map(PracticeScenario::getDifficulty)
                .filter(Objects::nonNull)
                .distinct()
                .sorted()
                .toList();

        List<String> availableCategories = all.stream()
                .map(PracticeScenario::getCategory)
                .filter(Objects::nonNull)
                .distinct()
                .sorted()
                .toList();

        List<PracticeScenario> filtered;
        if (difficulty != null && category != null) {
            filtered = practiceScenarioRepository.findByDifficulty(difficulty).stream()
                    .filter(s -> category.equals(s.getCategory()))
                    .toList();
        } else if (difficulty != null) {
            filtered = practiceScenarioRepository.findByDifficulty(difficulty);
        } else if (category != null) {
            filtered = practiceScenarioRepository.findByCategory(category);
        } else {
            filtered = all;
        }

        List<PracticeScenarioDto> scenarioDtos = filtered.stream()
                .map(mapper::toDto)
                .toList();

        return new FilteredScenariosDto(scenarioDtos, scenarioDtos.size(),
                availableDifficulties, availableCategories);
    }
}
