package com.example.FreStyle.repository;

import java.util.List;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.FreStyle.entity.PracticeScenario;

@Repository
public interface PracticeScenarioRepository extends JpaRepository<PracticeScenario, Integer> {

    List<PracticeScenario> findByDifficulty(String difficulty);

    List<PracticeScenario> findByCategory(String category);
}
