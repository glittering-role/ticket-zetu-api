package queue

import (
	"sync"
)

type Job func()

type JobQueue struct {
	workers   int
	jobChan   chan Job
	waitGroup sync.WaitGroup
}

func NewJobQueue(workers int) *JobQueue {
	q := &JobQueue{
		workers: workers,
		jobChan: make(chan Job, 100),
	}
	q.startWorkers()
	return q
}

func (q *JobQueue) startWorkers() {
	for i := 0; i < q.workers; i++ {
		go q.worker()
	}
}

func (q *JobQueue) worker() {
	for job := range q.jobChan {
		job()
		q.waitGroup.Done()
	}
}

func (q *JobQueue) Enqueue(job Job) {
	q.waitGroup.Add(1)
	q.jobChan <- job
}

func (q *JobQueue) Wait() {
	q.waitGroup.Wait()
}

func (q *JobQueue) Close() {
	close(q.jobChan)
}
