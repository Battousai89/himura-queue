package worker

import (
	"sync"
	"sync/atomic"
	"time"
)

type Task func()

type Pool struct {
	taskChan    chan Task
	wg          sync.WaitGroup
	active      int32
	maxWorkers  int
	minWorkers  int
	idleTimeout time.Duration
	stopChan    chan struct{}
	stopped     int32
}

func NewPool(bufferSize, minWorkers, maxWorkers int, idleTimeout time.Duration) *Pool {
	p := &Pool{
		taskChan:    make(chan Task, bufferSize),
		maxWorkers:  maxWorkers,
		minWorkers:  minWorkers,
		idleTimeout: idleTimeout,
		stopChan:    make(chan struct{}),
	}
	for i := 0; i < minWorkers; i++ {
		p.spawnWorker()
	}
	go p.scaler()
	return p
}

func (p *Pool) spawnWorker() {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		for {
			select {
			case task, ok := <-p.taskChan:
				if !ok {
					return
				}
				atomic.AddInt32(&p.active, 1)
				task()
				atomic.AddInt32(&p.active, -1)
			case <-time.After(p.idleTimeout):
				if int(atomic.LoadInt32(&p.active)) > p.minWorkers {
					return
				}
			case <-p.stopChan:
				return
			}
		}
	}()
}

func (p *Pool) scaler() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if atomic.LoadInt32(&p.stopped) == 1 {
				return
			}
			load := float64(len(p.taskChan)) / float64(cap(p.taskChan))
			current := int(atomic.LoadInt32(&p.active))
			if load > 0.8 && current < p.maxWorkers {
				p.spawnWorker()
			}
		case <-p.stopChan:
			return
		}
	}
}

func (p *Pool) Submit(task Task) {
	if atomic.LoadInt32(&p.stopped) == 1 {
		return
	}
	select {
	case p.taskChan <- task:
	default:
		p.spawnWorker()
		p.taskChan <- task
	}
}

func (p *Pool) Shutdown() {
	if !atomic.CompareAndSwapInt32(&p.stopped, 0, 1) {
		return
	}
	close(p.stopChan)
	p.wg.Wait()
	close(p.taskChan)
}
