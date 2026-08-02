package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	ctx "github.com/gophish/gophish/context"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

type campaignsCompleteRequest struct {
	CampaignIDs []int64 `json:"campaign_ids"`
}

type campaignActionResult struct {
	ID      int64  `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// Campaigns returns a list of campaigns if requested via GET.
// If requested via POST, APICampaigns creates a new campaign and returns a reference to it.
func (as *Server) Campaigns(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		cs, err := models.GetCampaigns(ctx.Get(r, "user_id").(int64))
		if err != nil {
			log.Error(err)
		}
		JSONResponse(w, cs, http.StatusOK)
	//POST: Create a new campaign and return it as JSON
	case r.Method == "POST":
		c := models.Campaign{}
		// Put the request into a campaign
		err := json.NewDecoder(r.Body).Decode(&c)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid JSON structure"}, http.StatusBadRequest)
			return
		}
		err = models.PostCampaign(&c, ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		// If the campaign is scheduled to launch immediately, send it to the worker.
		// Otherwise, the worker will pick it up at the scheduled time
		if c.Status == models.CampaignInProgress {
			go as.worker.LaunchCampaign(c)
		}
		JSONResponse(w, c, http.StatusCreated)
	}
}

// CampaignsSummary returns the summary for the current user's campaigns
func (as *Server) CampaignsSummary(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		cs, err := models.GetCampaignSummaries(ctx.Get(r, "user_id").(int64))
		if err != nil {
			log.Error(err)
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, cs, http.StatusOK)
	}
}

// Campaign returns details about the requested campaign. If the campaign is not
// valid, APICampaign returns null.
func (as *Server) Campaign(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	c, err := models.GetCampaign(id, ctx.Get(r, "user_id").(int64))
	if err != nil {
		log.Error(err)
		JSONResponse(w, models.Response{Success: false, Message: "Campaign not found"}, http.StatusNotFound)
		return
	}
	switch {
	case r.Method == "GET":
		JSONResponse(w, c, http.StatusOK)
	case r.Method == "DELETE":
		err = models.DeleteCampaign(id)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error deleting campaign"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Campaign deleted successfully!"}, http.StatusOK)
	}
}

// CampaignResults returns just the results for a given campaign to
// significantly reduce the information returned.
func (as *Server) CampaignResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	cr, err := models.GetCampaignResults(id, ctx.Get(r, "user_id").(int64))
	if err != nil {
		log.Error(err)
		JSONResponse(w, models.Response{Success: false, Message: "Campaign not found"}, http.StatusNotFound)
		return
	}
	if r.Method == "GET" {
		JSONResponse(w, cr, http.StatusOK)
		return
	}
}

// CampaignSummary returns the summary for a given campaign.
func (as *Server) CampaignSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	switch {
	case r.Method == "GET":
		cs, err := models.GetCampaignSummary(id, ctx.Get(r, "user_id").(int64))
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				JSONResponse(w, models.Response{Success: false, Message: "Campaign not found"}, http.StatusNotFound)
			} else {
				JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			}
			log.Error(err)
			return
		}
		JSONResponse(w, cs, http.StatusOK)
	}
}

// CampaignComplete effectively "ends" a campaign.
// Future phishing emails clicked will return a simple "404" page.
func (as *Server) CampaignComplete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	switch {
	case r.Method == "GET":
		err := models.CompleteCampaign(id, ctx.Get(r, "user_id").(int64))
		if err != nil {
			if errors.Is(err, models.ErrCampaignProcessing) {
				JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusConflict)
				return
			}
			JSONResponse(w, models.Response{Success: false, Message: "Error completing campaign"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Campaign completed successfully!"}, http.StatusOK)
	}
}

func (as *Server) CampaignsComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	request := campaignsCompleteRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || len(request.CampaignIDs) == 0 {
		JSONResponse(w, models.Response{Success: false, Message: "Invalid campaign IDs"}, http.StatusBadRequest)
		return
	}
	seen := make(map[int64]struct{}, len(request.CampaignIDs))
	for _, id := range request.CampaignIDs {
		if id <= 0 {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid campaign IDs"}, http.StatusBadRequest)
			return
		}
		if _, exists := seen[id]; exists {
			JSONResponse(w, models.Response{Success: false, Message: "Duplicate campaign IDs"}, http.StatusBadRequest)
			return
		}
		seen[id] = struct{}{}
	}
	userID := ctx.Get(r, "user_id").(int64)
	results := make([]campaignActionResult, 0, len(request.CampaignIDs))
	allSucceeded := true
	for _, id := range request.CampaignIDs {
		result := campaignActionResult{ID: id, Status: "completed"}
		err := models.CompleteCampaign(id, userID)
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			result.Status = "not_found"
			result.Message = "Campaign not found"
			allSucceeded = false
		case errors.Is(err, models.ErrCampaignProcessing):
			result.Status = "processing"
			result.Message = err.Error()
			allSucceeded = false
		case err != nil:
			result.Status = "error"
			result.Message = "Error completing campaign"
			allSucceeded = false
		}
		results = append(results, result)
	}
	JSONResponse(w, models.Response{Success: allSucceeded, Data: map[string]interface{}{"results": results}}, http.StatusOK)
}

func (as *Server) CampaignResendErrors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	queued, err := models.ResendErroredResults(id, ctx.Get(r, "user_id").(int64), time.Now().UTC())
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		JSONResponse(w, models.Response{Success: false, Message: "Campaign not found"}, http.StatusNotFound)
	case errors.Is(err, models.ErrCampaignNotActive):
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusConflict)
	case err != nil:
		JSONResponse(w, models.Response{Success: false, Message: "Error retrying failed emails"}, http.StatusInternalServerError)
	default:
		JSONResponse(w, models.Response{Success: true, Data: map[string]interface{}{"queued": queued}}, http.StatusAccepted)
	}
}
