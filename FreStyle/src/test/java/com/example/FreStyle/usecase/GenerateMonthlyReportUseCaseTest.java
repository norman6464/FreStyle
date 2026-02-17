package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.sql.Timestamp;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.LearningReportDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.CommunicationScore;
import com.example.FreStyle.entity.LearningReport;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.CommunicationScoreRepository;
import com.example.FreStyle.repository.LearningReportRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GenerateMonthlyReportUseCase テスト")
class GenerateMonthlyReportUseCaseTest {

    @Mock
    private CommunicationScoreRepository communicationScoreRepository;

    @Mock
    private LearningReportRepository learningReportRepository;

    @InjectMocks
    private GenerateMonthlyReportUseCase generateMonthlyReportUseCase;

    private User testUser;

    @BeforeEach
    void setUp() {
        testUser = new User();
        testUser.setId(1);
        testUser.setName("テストユーザー");
    }

    @Test
    @DisplayName("スコアデータから月次レポートを生成できる")
    void execute_generatesReport() {
        AiChatSession session1 = new AiChatSession();
        session1.setId(1);
        AiChatSession session2 = new AiChatSession();
        session2.setId(2);

        CommunicationScore score1 = new CommunicationScore();
        score1.setUser(testUser);
        score1.setSession(session1);
        score1.setAxisName("論理的構成力");
        score1.setScore(80);
        score1.setCreatedAt(Timestamp.valueOf(LocalDateTime.of(2026, 1, 15, 10, 0)));

        CommunicationScore score2 = new CommunicationScore();
        score2.setUser(testUser);
        score2.setSession(session1);
        score2.setAxisName("配慮表現");
        score2.setScore(60);
        score2.setCreatedAt(Timestamp.valueOf(LocalDateTime.of(2026, 1, 15, 10, 0)));

        CommunicationScore score3 = new CommunicationScore();
        score3.setUser(testUser);
        score3.setSession(session2);
        score3.setAxisName("論理的構成力");
        score3.setScore(90);
        score3.setCreatedAt(Timestamp.valueOf(LocalDateTime.of(2026, 1, 20, 10, 0)));

        when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(List.of(score3, score2, score1));
        when(learningReportRepository.findByUserIdAndYearAndMonth(1, 2025, 12))
                .thenReturn(Optional.empty());
        when(learningReportRepository.findByUserIdAndYearAndMonth(1, 2026, 1))
                .thenReturn(Optional.empty());
        when(learningReportRepository.save(any(LearningReport.class)))
                .thenAnswer(inv -> {
                    LearningReport r = inv.getArgument(0);
                    r.setId(100);
                    return r;
                });

        LearningReportDto result = generateMonthlyReportUseCase.execute(testUser, 2026, 1);

        assertThat(result).isNotNull();
        assertThat(result.getYear()).isEqualTo(2026);
        assertThat(result.getMonth()).isEqualTo(1);
        assertThat(result.getTotalSessions()).isEqualTo(2);
        assertThat(result.getBestAxis()).isEqualTo("論理的構成力");
        assertThat(result.getWorstAxis()).isEqualTo("配慮表現");
        verify(learningReportRepository).save(any(LearningReport.class));
    }

    @Test
    @DisplayName("スコアデータがない月でもレポートを生成できる")
    void execute_generatesEmptyReport() {
        when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(List.of());
        when(learningReportRepository.findByUserIdAndYearAndMonth(1, 2026, 1))
                .thenReturn(Optional.empty());
        when(learningReportRepository.findByUserIdAndYearAndMonth(1, 2026, 2))
                .thenReturn(Optional.empty());
        when(learningReportRepository.save(any(LearningReport.class)))
                .thenAnswer(inv -> {
                    LearningReport r = inv.getArgument(0);
                    r.setId(101);
                    return r;
                });

        LearningReportDto result = generateMonthlyReportUseCase.execute(testUser, 2026, 2);

        assertThat(result).isNotNull();
        assertThat(result.getTotalSessions()).isEqualTo(0);
        assertThat(result.getAverageScore()).isEqualTo(0.0);
    }

    @Test
    @DisplayName("既存レポートがある場合は更新する")
    void execute_updatesExistingReport() {
        LearningReport existing = new LearningReport();
        existing.setId(50);
        existing.setUser(testUser);
        existing.setYear(2026);
        existing.setMonth(1);

        when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(List.of());
        when(learningReportRepository.findByUserIdAndYearAndMonth(1, 2025, 12))
                .thenReturn(Optional.empty());
        when(learningReportRepository.findByUserIdAndYearAndMonth(1, 2026, 1))
                .thenReturn(Optional.of(existing));
        when(learningReportRepository.save(any(LearningReport.class)))
                .thenAnswer(inv -> inv.getArgument(0));

        LearningReportDto result = generateMonthlyReportUseCase.execute(testUser, 2026, 1);

        assertThat(result).isNotNull();
        assertThat(result.getId()).isEqualTo(50);
    }
}
