package com.example.FreStyle.repository;

import java.util.List;
import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import com.example.FreStyle.dto.UserDto;
import com.example.FreStyle.entity.User;

@Repository
public interface UserRepository extends JpaRepository<User, Integer> {

    boolean existsByEmail(String email);

    boolean existsByName(String name);

    Optional<User> findByEmail(String email);

    @Query("SELECT new com.example.FreStyle.dto.UserDto(u.id, u.email, u.name) FROM User u WHERE (u.email LIKE :query OR u.name LIKE :query) AND u.id <> :id")
    List<UserDto> findIdAndEmailByEmailLikeDtos(@Param("id") Integer id, @Param("query") String query);

    @Query("SELECT new com.example.FreStyle.dto.UserDto(u.id, u.email, u.name) FROM User u WHERE u.id <> :id")
    List<UserDto> findAllUserDtos(@Param("id") Integer id);

}
