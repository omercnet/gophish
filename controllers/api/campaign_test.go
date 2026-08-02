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
