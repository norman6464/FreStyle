package com.example.FreStyle.usecase;

import java.util.List;
import java.util.stream.Collectors;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.FriendshipDto;
import com.example.FreStyle.mapper.FriendshipMapper;
import com.example.FreStyle.repository.FriendshipRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetFollowingUseCase {

    private final FriendshipRepository friendshipRepository;
    private final FriendshipMapper friendshipMapper;

    @Transactional(readOnly = true)
    public List<FriendshipDto> execute(Integer userId) {
        return friendshipRepository.findByFollowerIdOrderByCreatedAtDesc(userId)
                .stream()
                .map(f -> {
                    boolean mutual = friendshipRepository.existsByFollowerIdAndFollowingId(
                            f.getFollowing().getId(), userId);
                    return friendshipMapper.toFollowingDto(f, mutual);
                })
                .collect(Collectors.toList());
    }
}
