package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.PracticeScenarioDto;
import com.example.FreStyle.entity.PracticeScenario;
import com.example.FreStyle.form.PracticeScenarioForm;
import com.example.FreStyle.repository.PracticeScenarioRepository;

import lombok.RequiredArgsConstructor;

/**
 * 管理者がシナリオを新規作成するユースケース。
 */
@Service
@RequiredArgsConstructor
public class CreatePracticeScenarioUseCase {

    private final PracticeScenarioRepository repository;

    @Transactional
    public PracticeScenarioDto execute(PracticeScenarioForm form) {
        PracticeScenario entity = new PracticeScenario();
        entity.setName(form.name());
        entity.setDescription(form.description());
        entity.setCategory(form.category());
        entity.setRoleName(form.roleName());
        entity.setDifficulty(form.difficulty());
        entity.setSystemPrompt(form.systemPrompt());

        PracticeScenario saved = repository.save(entity);
        return new PracticeScenarioDto(
                saved.getId(),
                saved.getName(),
                saved.getDescription(),
                saved.getCategory(),
                saved.getRoleName(),
                saved.getDifficulty(),
                saved.getSystemPrompt()
        );
    }
}
