package com.example.FreStyle.repository;

import java.util.List;
import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.FreStyle.entity.ScenarioBookmark;

@Repository
public interface ScenarioBookmarkRepository extends JpaRepository<ScenarioBookmark, Integer> {

    List<ScenarioBookmark> findByUserId(Integer userId);

    Optional<ScenarioBookmark> findByUserIdAndScenarioId(Integer userId, Integer scenarioId);

    boolean existsByUserIdAndScenarioId(Integer userId, Integer scenarioId);

    void deleteByUserIdAndScenarioId(Integer userId, Integer scenarioId);
}
