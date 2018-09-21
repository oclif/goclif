package main

type Message struct {
	ID       string   `json:"id"`
	WorkerID *int     `json:"worker_id"`
	Type     string   `json:"type"`
	Argv     []string `json:"argv"`
}
