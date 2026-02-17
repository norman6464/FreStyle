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
        String normalizedDifficulty = isBlank(difficulty) ? null : difficulty.trim();
        String normalizedCategory = isBlank(category) ? null : category.trim();

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

        List<PracticeScenario> filtered = all.stream()
                .filter(s -> normalizedDifficulty == null || normalizedDifficulty.equals(s.getDifficulty()))
                .filter(s -> normalizedCategory == null || normalizedCategory.equals(s.getCategory()))
                .toList();

        List<PracticeScenarioDto> scenarioDtos = filtered.stream()
                .map(mapper::toDto)
                .toList();

        return new FilteredScenariosDto(scenarioDtos, scenarioDtos.size(),
                availableDifficulties, availableCategories);
    }

    private boolean isBlank(String s) {
        return s == null || s.isBlank();
    }
}
