package com.example.FreStyle.usecase;

import java.util.HashSet;
import java.util.List;
import java.util.Set;
import java.util.stream.Collectors;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.FriendshipDto;
import com.example.FreStyle.entity.Friendship;
import com.example.FreStyle.mapper.FriendshipMapper;
import com.example.FreStyle.repository.FriendshipRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetFollowersUseCase {

    private final FriendshipRepository friendshipRepository;
    private final FriendshipMapper friendshipMapper;

    @Transactional(readOnly = true)
    public List<FriendshipDto> execute(Integer userId) {
        List<Friendship> friendships = friendshipRepository.findByFollowingIdOrderByCreatedAtDesc(userId);

        if (friendships.isEmpty()) {
            return List.of();
        }

        List<Integer> followerIds = friendships.stream()
                .map(f -> f.getFollower().getId())
                .collect(Collectors.toList());

        Set<Integer> mutualIds = new HashSet<>(
                friendshipRepository.findMutualFollowingIds(followerIds, userId));

        return friendships.stream()
                .map(f -> friendshipMapper.toFollowerDto(f, mutualIds.contains(f.getFollower().getId())))
                .collect(Collectors.toList());
    }
}
