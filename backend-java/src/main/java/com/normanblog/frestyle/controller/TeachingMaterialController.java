package com.normanblog.frestyle.controller;

import com.normanblog.frestyle.dto.TeachingMaterialRequest;
import com.normanblog.frestyle.dto.TeachingMaterialResponse;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.security.CurrentUserProvider;
import com.normanblog.frestyle.service.TeachingMaterialService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/** 教材の閲覧・作成・更新・削除 API。作成/編集/削除は所属コースの管理者のみ。 */
@RestController
@RequestMapping("/api/v2/teaching-materials")
public class TeachingMaterialController {

  private final TeachingMaterialService materialService;
  private final CurrentUserProvider currentUser;

  public TeachingMaterialController(
      TeachingMaterialService materialService, CurrentUserProvider currentUser) {
    this.materialService = materialService;
    this.currentUser = currentUser;
  }

  @GetMapping("/{id}")
  public TeachingMaterialResponse get(@PathVariable Long id) {
    User actor = currentUser.require();

    return TeachingMaterialResponse.from(materialService.get(id, actor));
  }

  @PostMapping
  public TeachingMaterialResponse create(@RequestBody TeachingMaterialRequest request) {
    User actor = currentUser.require();

    return TeachingMaterialResponse.from(materialService.create(actor, request));
  }

  @PutMapping("/{id}")
  public TeachingMaterialResponse update(
      @PathVariable Long id, @RequestBody TeachingMaterialRequest request) {
    User actor = currentUser.require();

    return TeachingMaterialResponse.from(materialService.update(id, actor, request));
  }

  @DeleteMapping("/{id}")
  public ResponseEntity<Void> delete(@PathVariable Long id) {
    User actor = currentUser.require();
    materialService.delete(id, actor);

    return ResponseEntity.noContent().build();
  }
}
