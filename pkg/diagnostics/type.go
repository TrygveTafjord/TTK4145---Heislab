package diagnostics

type Diagnose int

const (
	Healthy Diagnose = iota
	Obstructed
	Reinitialize
	
)
