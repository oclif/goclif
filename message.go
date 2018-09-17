package main

type Message struct {
	Type string `json:"type"`
}

type MessageInit struct {
	Type string `json:"type"`
	ID   int    `json:"id"`
}

type MessageCommand struct {
	Type string   `json:"type"`
	Argv []string `json:"argv"`
}
