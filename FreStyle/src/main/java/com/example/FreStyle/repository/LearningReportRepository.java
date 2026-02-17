package com.example.FreStyle.repository;

import java.util.List;
import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.FreStyle.entity.LearningReport;

@Repository
public interface LearningReportRepository extends JpaRepository<LearningReport, Integer> {

    Optional<LearningReport> findByUserIdAndYearAndMonth(Integer userId, Integer year, Integer month);

    List<LearningReport> findByUserIdOrderByYearDescMonthDesc(Integer userId);
}
