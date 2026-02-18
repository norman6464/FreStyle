package com.example.FreStyle.usecase;

import java.util.Comparator;
import java.util.List;
import java.util.Optional;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.PracticeSessionSummaryDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.CommunicationScore;
import com.example.FreStyle.entity.PracticeScenario;
import com.example.FreStyle.entity.SessionNote;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.repository.AiChatMessageDynamoRepository;
import com.example.FreStyle.repository.AiChatSessionRepository;
import com.example.FreStyle.repository.CommunicationScoreRepository;
import com.example.FreStyle.repository.PracticeScenarioRepository;
import com.example.FreStyle.repository.SessionNoteRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetPracticeSessionSummaryUseCase {

    private final AiChatSessionRepository aiChatSessionRepository;
    private final AiChatMessageDynamoRepository aiChatMessageDynamoRepository;
    private final CommunicationScoreRepository communicationScoreRepository;
    private final SessionNoteRepository sessionNoteRepository;
    private final PracticeScenarioRepository practiceScenarioRepository;

    @Transactional(readOnly = true)
    public PracticeSessionSummaryDto execute(Integer sessionId, Integer userId) {
        AiChatSession session = aiChatSessionRepository.findByIdAndUserId(sessionId, userId)
                .orElseThrow(() -> new ResourceNotFoundException("セッションが見つかりません"));

        Long messageCount = aiChatMessageDynamoRepository.countBySessionId(sessionId);

        List<CommunicationScore> scores = communicationScoreRepository.findBySessionId(sessionId);

        List<PracticeSessionSummaryDto.ScoreDetail> scoreDetails = scores.stream()
                .map(s -> new PracticeSessionSummaryDto.ScoreDetail(s.getAxisName(), s.getScore(), s.getComment()))
                .toList();

        Double averageScore = scores.isEmpty() ? null :
                scores.stream().mapToInt(CommunicationScore::getScore).average().orElse(0.0);

        String bestAxis = scores.isEmpty() ? null :
                scores.stream().max(Comparator.comparingInt(CommunicationScore::getScore))
                        .map(CommunicationScore::getAxisName).orElse(null);

        String worstAxis = scores.size() <= 1 ? null :
                scores.stream().min(Comparator.comparingInt(CommunicationScore::getScore))
                        .map(CommunicationScore::getAxisName).orElse(null);

        // bestAxisとworstAxisが同じ場合はworstAxisをnullにする
        if (bestAxis != null && bestAxis.equals(worstAxis)) {
            worstAxis = null;
        }

        Optional<SessionNote> noteOpt = sessionNoteRepository.findByUserIdAndSessionId(userId, sessionId);
        String note = noteOpt.map(SessionNote::getNote).orElse(null);

        String scenarioName = null;
        if (session.getScenarioId() != null) {
            scenarioName = practiceScenarioRepository.findById(session.getScenarioId())
                    .map(PracticeScenario::getName).orElse(null);
        }

        return new PracticeSessionSummaryDto(
                session.getId(),
                session.getTitle(),
                session.getSessionType(),
                session.getScene(),
                session.getCreatedAt(),
                messageCount,
                scoreDetails,
                averageScore,
                bestAxis,
                worstAxis,
                note,
                scenarioName
        );
    }
}
