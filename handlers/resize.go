package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type resizeRequest struct {
	DiskGiB   *int `json:"disk_gib"`
	CPUCores  *int `json:"cpu_cores"`
	MemoryMiB *int `json:"memory_mib"`
}

func (h *Handler) Resize(w http.ResponseWriter, r *http.Request) {
	var req resizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.DiskGiB == nil && req.CPUCores == nil && req.MemoryMiB == nil {
		jsonErr(w, "at least one field required", http.StatusBadRequest)
		return
	}

	type diskResult struct {
		Expanded   bool `json:"expanded"`
		NewSizeGiB int  `json:"new_size_gib"`
	}
	type cpuResult struct {
		Online int `json:"online"`
	}
	type memResult struct {
		OnlineMB int `json:"online_mb"`
	}
	type resizeResp struct {
		Disk   *diskResult `json:"disk,omitempty"`
		CPU    *cpuResult  `json:"cpu,omitempty"`
		Memory *memResult  `json:"memory,omitempty"`
	}

	resp := resizeResp{}

	if req.DiskGiB != nil {
		_, err := callScript("resize-disk", nil)
		if err != nil {
			jsonErr(w, fmt.Sprintf("resize-disk failed: %s", err), http.StatusInternalServerError)
			return
		}
		resp.Disk = &diskResult{Expanded: true, NewSizeGiB: *req.DiskGiB}
	}

	if req.CPUCores != nil {
		out, err := callScript("online-cpus", nil)
		if err != nil {
			jsonErr(w, fmt.Sprintf("online-cpus failed: %s", err), http.StatusInternalServerError)
			return
		}
		n, _ := strconv.Atoi(strings.TrimSpace(out))
		resp.CPU = &cpuResult{Online: n}
	}

	if req.MemoryMiB != nil {
		_, err := callScript("online-memory", nil)
		if err != nil {
			jsonErr(w, fmt.Sprintf("online-memory failed: %s", err), http.StatusInternalServerError)
			return
		}
		resp.Memory = &memResult{OnlineMB: *req.MemoryMiB}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
