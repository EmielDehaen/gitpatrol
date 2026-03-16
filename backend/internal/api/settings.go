package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetSettings(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"github_token_set": h.config.GithubToken != "",
		"gitlab_token_set": h.config.GitlabToken != "",
	})
}

func (h *Handler) UpdateSettings(c echo.Context) error {
	var body struct {
		GithubToken string `json:"github_token"`
		GitlabToken string `json:"gitlab_token"`
	}
	if err := c.Bind(&body); err != nil {
		return err
	}

	if err := h.config.UpdateTokens(body.GithubToken, body.GitlabToken); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to save tokens"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "settings updated"})
}
