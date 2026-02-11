package com.example.FreStyle.service;

import java.util.List;
import java.util.stream.Collectors;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.PracticeScenarioDto;
import com.example.FreStyle.entity.PracticeScenario;
import com.example.FreStyle.repository.PracticeScenarioRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class PracticeScenarioService {

    private final PracticeScenarioRepository practiceScenarioRepository;

    @Transactional(readOnly = true)
    public List<PracticeScenarioDto> getAllScenarios() {
        return practiceScenarioRepository.findAll().stream()
                .map(this::toDto)
                .collect(Collectors.toList());
    }

    @Transactional(readOnly = true)
    public List<PracticeScenarioDto> getScenariosByCategory(String category) {
        return practiceScenarioRepository.findByCategory(category).stream()
                .map(this::toDto)
                .collect(Collectors.toList());
    }

    @Transactional(readOnly = true)
    public PracticeScenarioDto getScenarioById(Integer id) {
        PracticeScenario scenario = practiceScenarioRepository.findById(id)
                .orElseThrow(() -> new RuntimeException("シナリオが見つかりません: " + id));
        return toDto(scenario);
    }

    @Transactional(readOnly = true)
    public PracticeScenario getScenarioEntityById(Integer id) {
        return practiceScenarioRepository.findById(id)
                .orElseThrow(() -> new RuntimeException("シナリオが見つかりません: " + id));
    }

    private PracticeScenarioDto toDto(PracticeScenario entity) {
        return new PracticeScenarioDto(
                entity.getId(),
                entity.getName(),
                entity.getDescription(),
                entity.getCategory(),
                entity.getRoleName(),
                entity.getDifficulty()
        );
    }
}
