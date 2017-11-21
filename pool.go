package nats_websocket

type Pool struct {
	work chan func()
	sem  chan bool
}

func NewPool(size int) *Pool {
	return &Pool{
		work: make(chan func()),
		sem:  make(chan bool, size),
	}
}

func (p *Pool) Schedule(task func()) {

	select {
	case p.work <- task:
	case p.sem <- true:
		go p.worker(task)
	}
}

func (p *Pool) worker(task func()) {
	defer func() { <-p.sem }()

	for {
		task()
		task = <-p.work
	}
}
