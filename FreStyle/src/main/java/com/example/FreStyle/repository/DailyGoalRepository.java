package com.example.FreStyle.repository;

import java.time.LocalDate;
import java.util.List;
import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;

import com.example.FreStyle.entity.DailyGoal;

public interface DailyGoalRepository extends JpaRepository<DailyGoal, Integer> {
    Optional<DailyGoal> findByUserIdAndGoalDate(Integer userId, LocalDate goalDate);

    List<DailyGoal> findByUserIdOrderByGoalDateDesc(Integer userId);
}
