package main

import (
	"github.com/Oleg-Neevin/distributed_calc/internal/agent"
	"github.com/Oleg-Neevin/distributed_calc/internal/orchestrator"
)

func main() {
	go agent.StartAgent()
	orchestrator.RunOrchestrator()

}
