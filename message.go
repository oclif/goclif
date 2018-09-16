package main

type commandMessage struct {
	Type string   `json:"type"`
	Argv []string `json:"argv"`
}
