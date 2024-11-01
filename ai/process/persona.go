package process

type TeamLead struct {
	LifeCycle LifeCycle `json:"life_cycle"`
}

type LifeCycle struct {
	Process Process  `json:"process"`
	Actions []Action `json:"actions"`
}
