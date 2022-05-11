package command

import (
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
	"wander/message"
	"wander/nomad"
)

func simulateLoading() {
	for i := 0; i < 1e9; i++ {

	}
}

func FetchJobs(url, token string) tea.Cmd {
	return func() tea.Msg {
		// TODO LEO: error handling
		body, _ := nomad.GetJobs(url, token)
		//simulateLoading()
		//body := MockJobsResponse
		var jobResponse []nomad.JobResponseEntry
		if err := json.Unmarshal(body, &jobResponse); err != nil {
			// TODO LEO: error handling
			fmt.Println("Failed to decode job response")
		}

		return message.NomadJobsMsg(jobResponse)
	}
}

func FetchAllocations(url, token, jobId string) tea.Cmd {
	return func() tea.Msg {
		// TODO LEO: error handling
		body, _ := nomad.GetAllocations(url, token, jobId)
		//simulateLoading()
		//body := MockAllocationResponse
		var allocationResponse []nomad.AllocationResponseEntry
		if err := json.Unmarshal(body, &allocationResponse); err != nil {
			// TODO LEO: error handling
			fmt.Println("Failed to decode allocation response")
		}
		var allocationRowEntries []nomad.AllocationRowEntry
		for _, alloc := range allocationResponse {
			for taskName, task := range alloc.TaskStates {
				allocationRowEntries = append(allocationRowEntries, nomad.AllocationRowEntry{
					ID:         alloc.ID,
					Name:       alloc.Name,
					TaskName:   taskName,
					State:      task.State,
					StartedAt:  task.StartedAt,
					FinishedAt: task.FinishedAt,
				})
			}
		}

		return message.NomadAllocationMsg(allocationRowEntries)
	}
}

func FetchLogs(url, token, allocId, taskName string) tea.Cmd {
	return func() tea.Msg {
		// TODO LEO: error handling
		body, _ := nomad.GetLogs(url, token, allocId, taskName)
		logs := strings.Split(string(body), "\n")
		return message.NomadLogsMsg(logs)
	}
}
