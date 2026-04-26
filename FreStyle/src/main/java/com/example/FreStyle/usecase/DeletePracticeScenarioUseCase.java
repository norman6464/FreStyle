package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.repository.PracticeScenarioRepository;

import lombok.RequiredArgsConstructor;

/**
 * 管理者がシナリオを削除するユースケース。
 */
@Service
@RequiredArgsConstructor
public class DeletePracticeScenarioUseCase {

    private final PracticeScenarioRepository repository;

    @Transactional
    public void execute(Integer id) {
        if (!repository.existsById(id)) {
            throw new ResourceNotFoundException("シナリオ id=" + id + " が見つかりません");
        }
        repository.deleteById(id);
    }
}
