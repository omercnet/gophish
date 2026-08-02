package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gophish/gophish/models"
)

func TestCampaignsComplete_completes_owned_campaigns(t *testing.T) {
	testCtx := setupTest(t)
	createTestData(t)
	campaigns, err := models.GetCampaigns(testCtx.admin.Id)
	if err != nil || len(campaigns) != 1 {
		t.Fatalf("get campaign: %v", err)
	}
	body, _ := json.Marshal(campaignsCompleteRequest{CampaignIDs: []int64{campaigns[0].Id}})
	request := httptest.NewRequest(http.MethodPost, "/api/campaigns/complete", bytes.NewReader(body))
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testCtx.apiKey))
	response := httptest.NewRecorder()

	testCtx.apiServer.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}
	campaign, err := models.GetCampaign(campaigns[0].Id, testCtx.admin.Id)
	if err != nil || campaign.Status != models.CampaignComplete {
		t.Fatalf("campaign not completed: %v %#v", err, campaign)
	}
}

func TestCampaignResendErrors_requeues_terminal_failures(t *testing.T) {
	testCtx := setupTest(t)
	createTestData(t)
	campaigns, err := models.GetCampaigns(testCtx.admin.Id)
	if err != nil || len(campaigns) != 1 {
		t.Fatalf("get campaign: %v", err)
	}
	campaign, err := models.GetCampaign(campaigns[0].Id, testCtx.admin.Id)
	if err != nil {
		t.Fatalf("get campaign details: %v", err)
	}
	if err := campaign.UpdateStatus(models.CampaignInProgress); err != nil {
		t.Fatalf("activate campaign: %v", err)
	}
	mailLogs, err := models.GetMailLogsByCampaign(campaign.Id)
	if err != nil {
		t.Fatalf("get mail logs: %v", err)
	}
	if err := mailLogs[0].Error(fmt.Errorf("permanent failure")); err != nil {
		t.Fatalf("fail message: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/campaigns/%d/resend-errors", campaign.Id), nil)
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testCtx.apiKey))
	response := httptest.NewRecorder()

	testCtx.apiServer.ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", response.Code, response.Body.String())
	}
}
