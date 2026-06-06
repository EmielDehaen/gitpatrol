package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetSettings(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"github_token_set":   h.config.GithubToken != "",
		"gitlab_url":         h.config.GitlabURL,
		"gitlab_token_set":   h.config.GitlabToken != "",
		"gitea_url":          h.config.GiteaURL,
		"gitea_token_set":    h.config.GiteaToken != "",
		"export_destination": h.config.ExportDestination,
	})
}

func (h *Handler) UpdateSettings(c echo.Context) error {
	var body struct {
		GithubToken       string `json:"github_token"`
		GitlabURL         string `json:"gitlab_url"`
		GitlabToken       string `json:"gitlab_token"`
		GiteaURL          string `json:"gitea_url"`
		GiteaToken        string `json:"gitea_token"`
		ExportDestination string `json:"export_destination"`
	}
	if err := c.Bind(&body); err != nil {
		return err
	}

	if err := h.config.UpdateTokens(body.GithubToken, body.GitlabURL, body.GitlabToken, body.GiteaURL, body.GiteaToken, body.ExportDestination); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to save settings"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "settings updated"})
}
