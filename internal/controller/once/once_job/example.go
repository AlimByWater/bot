package once_job

import (
	"sync"
)

type ExampleJob struct {
	usecase exampleJobUseCase
	wg      *sync.WaitGroup
	once    sync.Once
}

func NewExampleJob(usecase exampleJobUseCase) *ExampleJob {
	return &ExampleJob{usecase: usecase}
}

type exampleJobUseCase interface {
	Gen()
}

func (j *ExampleJob) Init(wg *sync.WaitGroup) {
	j.wg = wg
}
func (j *ExampleJob) Do() {
	defer j.wg.Done()
	j.once.Do(j.usecase.Gen)
}
