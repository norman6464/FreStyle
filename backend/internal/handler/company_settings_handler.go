package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// CompanySettingsHandler は会社単位の設定（trainee への AI 有効化）の取得・更新を扱う。
type CompanySettingsHandler struct {
	get    *usecase.GetCompanyAiChatSettingUseCase
	update *usecase.UpdateCompanyAiChatSettingUseCase
}

func NewCompanySettingsHandler(
	get *usecase.GetCompanyAiChatSettingUseCase,
	update *usecase.UpdateCompanyAiChatSettingUseCase,
) *CompanySettingsHandler {
	return &CompanySettingsHandler{get: get, update: update}
}

type companySettingsResponse struct {
	AiChatEnabledForTrainees bool `json:"aiChatEnabledForTrainees"`
}

type updateCompanySettingsRequest struct {
	// bool の必須を binding:"required" で表現すると false が弾かれるため、ポインタで「指定の有無」を判定する。
	AiChatEnabledForTrainees *bool `json:"aiChatEnabledForTrainees" binding:"required"`
}

// Get は自社の AI 有効化フラグを返す。
//
//	@Summary      会社設定 取得 (AI 有効化)
//	@Description  自社の trainee への AI チャット有効化フラグを返す。company_admin / super_admin のみ。
//	@Tags         company
//	@Produce      json
//	@Success      200  {object}  companySettingsResponse
//	@Failure      400  {object}  errorResponse  "会社未所属"
//	@Failure      401  {object}  errorResponse  "未認証"
//	@Failure      403  {object}  errorResponse  "管理者以外"
//	@Failure      404  {object}  errorResponse  "会社が存在しない"
//	@Router       /company/settings [get]
//	@Security     CookieAuth
func (h *CompanySettingsHandler) Get(c *gin.Context) {
	actor := middleware.CurrentUserFromContext(c)
	enabled, err := h.get.Execute(c.Request.Context(), actor)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, companySettingsResponse{AiChatEnabledForTrainees: enabled})
}

// Update は自社の AI 有効化フラグを更新する。
//
//	@Summary      会社設定 更新 (AI 有効化)
//	@Description  自社の trainee への AI チャット有効化フラグを更新する。company_admin / super_admin のみ。
//	@Tags         company
//	@Accept       json
//	@Produce      json
//	@Param        body  body      updateCompanySettingsRequest  true  "aiChatEnabledForTrainees"
//	@Success      200   {object}  companySettingsResponse
//	@Failure      400   {object}  errorResponse  "バリデーション / 会社未所属"
//	@Failure      401   {object}  errorResponse  "未認証"
//	@Failure      403   {object}  errorResponse  "管理者以外"
//	@Router       /company/settings [put]
//	@Security     CookieAuth
func (h *CompanySettingsHandler) Update(c *gin.Context) {
	var req updateCompanySettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	actor := middleware.CurrentUserFromContext(c)
	enabled, err := h.update.Execute(c.Request.Context(), actor, *req.AiChatEnabledForTrainees)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, companySettingsResponse{AiChatEnabledForTrainees: enabled})
}

func (h *CompanySettingsHandler) writeErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrCompanySettingsForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	case errors.Is(err, usecase.ErrCompanySettingsNoCompany):
		c.JSON(http.StatusBadRequest, gin.H{"error": "no_company"})
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "company_not_found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
	}
}
