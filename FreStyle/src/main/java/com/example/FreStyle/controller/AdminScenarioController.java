package com.example.FreStyle.controller;

import java.util.List;

import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.dto.PracticeScenarioDto;
import com.example.FreStyle.entity.PracticeScenario;
import com.example.FreStyle.form.PracticeScenarioForm;
import com.example.FreStyle.repository.PracticeScenarioRepository;
import com.example.FreStyle.usecase.CreatePracticeScenarioUseCase;
import com.example.FreStyle.usecase.DeletePracticeScenarioUseCase;
import com.example.FreStyle.usecase.UpdatePracticeScenarioUseCase;

import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

/**
 * 管理者専用のシナリオ CRUD API。
 *
 * <p>Spring Security により {@code /api/admin/**} は Cognito の {@code cognito:groups} に
 * "admin" を含むユーザーのみアクセス可能。</p>
 */
@RestController
@RequiredArgsConstructor
@RequestMapping("/api/admin/scenarios")
@Slf4j
@Tag(name = "Admin Scenarios", description = "管理者専用: 練習シナリオの作成・更新・削除")
public class AdminScenarioController {

    private final CreatePracticeScenarioUseCase createUseCase;
    private final UpdatePracticeScenarioUseCase updateUseCase;
    private final DeletePracticeScenarioUseCase deleteUseCase;
    private final PracticeScenarioRepository repository;

    @GetMapping
    public ResponseEntity<List<PracticeScenarioDto>> list() {
        List<PracticeScenarioDto> scenarios = repository.findAll().stream()
                .map(this::toDto)
                .toList();
        return ResponseEntity.ok(scenarios);
    }

    @PostMapping
    public ResponseEntity<PracticeScenarioDto> create(@Valid @RequestBody PracticeScenarioForm form) {
        log.info("[AdminScenarioController] create scenario name={}", form.name());
        return ResponseEntity.ok(createUseCase.execute(form));
    }

    @PutMapping("/{id}")
    public ResponseEntity<PracticeScenarioDto> update(
            @PathVariable Integer id,
            @Valid @RequestBody PracticeScenarioForm form
    ) {
        log.info("[AdminScenarioController] update scenario id={}", id);
        return ResponseEntity.ok(updateUseCase.execute(id, form));
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> delete(@PathVariable Integer id) {
        log.info("[AdminScenarioController] delete scenario id={}", id);
        deleteUseCase.execute(id);
        return ResponseEntity.noContent().build();
    }

    private PracticeScenarioDto toDto(PracticeScenario s) {
        return new PracticeScenarioDto(
                s.getId(),
                s.getName(),
                s.getDescription(),
                s.getCategory(),
                s.getRoleName(),
                s.getDifficulty(),
                s.getSystemPrompt()
        );
    }
}
