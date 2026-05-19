package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Mininglamp-OSS/octo-speech/internal/store"
)

const maxContentLength = 10000

type VocabularyHandler struct {
	vocabStore *store.VocabularyStore
}

func NewVocabularyHandler(vocabStore *store.VocabularyStore) *VocabularyHandler {
	return &VocabularyHandler{vocabStore: vocabStore}
}

func (h *VocabularyHandler) Put(c *gin.Context) {
	appID, _ := c.Get("app_id")
	appIDStr, _ := appID.(string)

	var req struct {
		SubjectID string `json:"subject_id"`
		ScopeType string `json:"scope_type"`
		ScopeID   string `json:"scope_id"`
		Content   string `json:"content"`
		UpdatedBy string `json:"updated_by"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "invalid request body",
		})
		return
	}

	if req.SubjectID == "" || req.ScopeType == "" || req.ScopeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "subject_id, scope_type, and scope_id are required",
		})
		return
	}

	if !store.IsValidScopeType(req.ScopeType) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "invalid scope_type, expected: global, space, org, project",
		})
		return
	}

	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "content is required",
		})
		return
	}

	if len([]rune(req.Content)) > maxContentLength {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "content exceeds maximum length of 10000 characters",
		})
		return
	}

	if req.UpdatedBy == "" {
		req.UpdatedBy = appIDStr
	}

	if err := h.vocabStore.Upsert(appIDStr, req.SubjectID, req.ScopeType, req.ScopeID, req.Content, req.UpdatedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"msg":    "failed to save vocabulary",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"msg":    "ok",
	})
}

func (h *VocabularyHandler) Get(c *gin.Context) {
	appID, _ := c.Get("app_id")
	appIDStr, _ := appID.(string)

	subjectID := c.Query("subject_id")
	scopeType := c.Query("scope_type")
	scopeID := c.Query("scope_id")

	if subjectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "subject_id is required",
		})
		return
	}

	if scopeType == "" {
		scopeType = "global"
	}
	if scopeID == "" {
		scopeID = "default"
	}

	if !store.IsValidScopeType(scopeType) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "invalid scope_type, expected: global, space, org, project",
		})
		return
	}

	vocab, err := h.vocabStore.QueryWithPriority(appIDStr, subjectID, scopeType, scopeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"msg":    "failed to query vocabulary",
		})
		return
	}

	if vocab == nil {
		c.JSON(http.StatusOK, gin.H{
			"status":      http.StatusOK,
			"has_content": false,
			"content":     "",
			"updated_at":  "",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      http.StatusOK,
		"has_content": true,
		"content":     vocab.Content,
		"updated_at":  vocab.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *VocabularyHandler) Delete(c *gin.Context) {
	appID, _ := c.Get("app_id")
	appIDStr, _ := appID.(string)

	subjectID := c.Query("subject_id")
	scopeType := c.Query("scope_type")
	scopeID := c.Query("scope_id")

	if subjectID == "" || scopeType == "" || scopeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "subject_id, scope_type, and scope_id are required",
		})
		return
	}

	if !store.IsValidScopeType(scopeType) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "invalid scope_type, expected: global, space, org, project",
		})
		return
	}

	if err := h.vocabStore.Delete(appIDStr, subjectID, scopeType, scopeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"msg":    "failed to delete vocabulary",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"msg":    "ok",
	})
}
