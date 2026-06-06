package com.normanblog.frestyle.controller;

import com.normanblog.frestyle.dto.TeachingMaterialResponse;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.security.CurrentUserProvider;
import com.normanblog.frestyle.service.TeachingMaterialService;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/** 教材単体の閲覧 API。 */
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
}
