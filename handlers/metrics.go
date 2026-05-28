package handlers

import (
	"bufio"
	"encoding/json"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type cpuSample struct {
	user, nice, system, idle, iowait, irq, softirq, steal uint64
}

func (s cpuSample) total() uint64 {
	return s.user + s.nice + s.system + s.idle + s.iowait + s.irq + s.softirq + s.steal
}

func readCPUSample() (cpuSample, int, error) {
	f, err := os.Open("/host-proc/stat")
	if err != nil {
		return cpuSample{}, 0, err
	}
	defer f.Close()

	var s cpuSample
	cores := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		switch {
		case fields[0] == "cpu" && len(fields) >= 9:
			s.user, _ = strconv.ParseUint(fields[1], 10, 64)
			s.nice, _ = strconv.ParseUint(fields[2], 10, 64)
			s.system, _ = strconv.ParseUint(fields[3], 10, 64)
			s.idle, _ = strconv.ParseUint(fields[4], 10, 64)
			s.iowait, _ = strconv.ParseUint(fields[5], 10, 64)
			s.irq, _ = strconv.ParseUint(fields[6], 10, 64)
			s.softirq, _ = strconv.ParseUint(fields[7], 10, 64)
			s.steal, _ = strconv.ParseUint(fields[8], 10, 64)
		case strings.HasPrefix(fields[0], "cpu") && fields[0] != "cpu":
			cores++
		}
	}
	return s, cores, scanner.Err()
}

func (h *Handler) runCPULoop() {
	for {
		s1, cores, err := readCPUSample()
		if err == nil {
			time.Sleep(time.Second)
			s2, _, _ := readCPUSample()
			totalDelta := s2.total() - s1.total()
			idleDelta := s2.idle - s1.idle
			var usage float64
			if totalDelta > 0 {
				usage = float64(totalDelta-idleDelta) / float64(totalDelta) * 100
			}
			h.cpuMu.Lock()
			h.cpu = &cpuStat{
				Cores:        cores,
				UsagePercent: math.Round(usage*10) / 10,
			}
			h.cpuMu.Unlock()
		}
		time.Sleep(4 * time.Second)
	}
}

func readMeminfo() (totalMB, usedMB, freeMB int64, err error) {
	f, err := os.Open("/host-proc/meminfo")
	if err != nil {
		return
	}
	defer f.Close()

	vals := map[string]int64{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) >= 2 {
			vals[strings.TrimSuffix(parts[0], ":")] , _ = strconv.ParseInt(parts[1], 10, 64)
		}
	}
	totalMB = vals["MemTotal"] / 1024
	freeMB = vals["MemAvailable"] / 1024
	usedMB = totalMB - freeMB
	return
}

func (h *Handler) Metrics(w http.ResponseWriter, r *http.Request) {
	h.cpuMu.RLock()
	cpu := h.cpu
	h.cpuMu.RUnlock()

	totalMB, usedMB, freeMB, err := readMeminfo()
	if err != nil {
		jsonErr(w, "failed to read meminfo", http.StatusInternalServerError)
		return
	}

	diskOut, err := callScript("disk-info", nil)
	if err != nil {
		jsonErr(w, "disk-info failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	parts := strings.Fields(strings.TrimSpace(diskOut))
	var totalGB, usedGB, freeGB float64
	const bytesPerGB = float64(1 << 30)
	if len(parts) >= 3 {
		if v, err2 := strconv.ParseFloat(parts[0], 64); err2 == nil {
			totalGB = math.Round(v/bytesPerGB*100) / 100
		}
		if v, err2 := strconv.ParseFloat(parts[1], 64); err2 == nil {
			usedGB = math.Round(v/bytesPerGB*100) / 100
		}
		if v, err2 := strconv.ParseFloat(parts[2], 64); err2 == nil {
			freeGB = math.Round(v/bytesPerGB*100) / 100
		}
	}

	type cpuResp struct {
		Cores        int     `json:"cores"`
		UsagePercent float64 `json:"usage_percent"`
	}
	type memResp struct {
		TotalMB int64 `json:"total_mb"`
		UsedMB  int64 `json:"used_mb"`
		FreeMB  int64 `json:"free_mb"`
	}
	type diskResp struct {
		TotalGB float64 `json:"total_gb"`
		UsedGB  float64 `json:"used_gb"`
		FreeGB  float64 `json:"free_gb"`
	}
	type metricsResp struct {
		CPU    cpuResp  `json:"cpu"`
		Memory memResp  `json:"memory"`
		Disk   diskResp `json:"disk"`
	}

	var cpuData cpuResp
	if cpu != nil {
		cpuData = cpuResp{Cores: cpu.Cores, UsagePercent: cpu.UsagePercent}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metricsResp{
		CPU:    cpuData,
		Memory: memResp{TotalMB: totalMB, UsedMB: usedMB, FreeMB: freeMB},
		Disk:   diskResp{TotalGB: totalGB, UsedGB: usedGB, FreeGB: freeGB},
	})
}
