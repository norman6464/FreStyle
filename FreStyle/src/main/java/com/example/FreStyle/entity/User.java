package com.example.FreStyle.entity;


import lombok.*;

import java.sql.Timestamp;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.Table;

@Entity
@Table(name = "users")
@Data
@NoArgsConstructor
public class User {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Integer id;

    @Column(length = 150, nullable = false, unique = true)
    private String username;

    @Column(length = 254, nullable = false, unique = true)
    private String email;

    @Column(name = "password_hash", length = 128, nullable = false)
    private String passwordHash;

    @Column(name = "icon_url", length = 255)
    private String iconUrl;

    @Column(columnDefinition = "TEXT")
    private String bio;

    @Column(name = "is_active")
    private Boolean isActive = true;

    @Column(name = "created_at", insertable = false, updatable = false)
    private Timestamp createdAt;
    
    @Column(name = "updated_at", insertable = false, updatable = false)
    private Timestamp updatedAt;

    // Optional: mappedBy examples for bidirectional relationships (can be added later)
}
