package _const

type ProgramType string
type ProType string

const (
	WeightLoss  = ProgramType("weight_loss")
	StressWork  = ProgramType("stress_work")
	Recommended = ProType("recommended")
	Personal    = ProType("personal")
)
