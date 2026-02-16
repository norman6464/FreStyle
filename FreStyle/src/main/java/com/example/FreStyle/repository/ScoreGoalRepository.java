package com.example.FreStyle.repository;

import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;

import com.example.FreStyle.entity.ScoreGoal;

public interface ScoreGoalRepository extends JpaRepository<ScoreGoal, Integer> {
    Optional<ScoreGoal> findByUserId(Integer userId);
}
