// waterfall_overview.go — Gastown Waterfall Overview
// Canvas swim-lane view of agent runs grouped by rig.
// Handler: GET /waterfall_overview
package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"time"
)

func handleWaterfallOverview(w http.ResponseWriter, r *http.Request) {
	winStart := since(r)
	winEnd := winEndTime(r)
	if winEnd.IsZero() {
		winEnd = time.Now()
	}
	data, err := loadWaterfallV2(globalCfg, winStart, winEnd)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	type pageData struct {
		JSONData template.JS
		Window   string
		Instance string
		TownRoot string
	}
	render(w, tmplWaterfallOverview, pageData{
		JSONData: template.JS(jsonBytes),
		Window:   windowLabel(r),
		Instance: data.Instance,
		TownRoot: data.TownRoot,
	})
}
