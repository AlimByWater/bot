package scheduler_job

import "github.com/go-co-op/gocron"

type ListenerCheck struct {
	usecase listenerCheckUseCase
}

func NewListenerCheckJob(usecase listenerCheckUseCase) ListenerCheck {
	return ListenerCheck{usecase: usecase}
}

type listenerCheckUseCase interface {
	CheckIdleListenersAndSaveListeningDuration()
}

func (j ListenerCheck) Do(sh *gocron.Scheduler) (*gocron.Job, error) {
	return sh.Every(1).Minutes().SingletonMode().Do(j.usecase.CheckIdleListenersAndSaveListeningDuration)
}
