package scheduler

type Scheduler interface {
	SelectCanidateNodes()
	Score()
	Pick()
}