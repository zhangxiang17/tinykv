package worker

import "sync"

type TaskStop struct{}

type Task interface{}

type Worker struct {
	name     string
	sender   chan<- Task //单向只写通道
	receiver <-chan Task //单向只读通道
	closeCh  chan struct{}
	wg       *sync.WaitGroup
}

type TaskHandler interface {
	Handle(t Task)
}

type Starter interface {
	Start()
}

// 关注一下此处 wait.group的add方法
// 这里worker 角色，有start & stop, 如何停止这个work呢，就是发送一个 stop方法，
func (w *Worker) Start(handler TaskHandler) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		if s, ok := handler.(Starter); ok {
			s.Start()
		}
		for {
			Task := <-w.receiver
			if _, ok := Task.(TaskStop); ok {
				return
			}
			handler.Handle(Task)
		}
	}()
}

func (w *Worker) Sender() chan<- Task {
	return w.sender
}

func (w *Worker) Stop() {
	w.sender <- TaskStop{}
}

const defaultWorkerCapacity = 128

func NewWorker(name string, wg *sync.WaitGroup) *Worker {
	ch := make(chan Task, defaultWorkerCapacity)
	return &Worker{
		sender:   (chan<- Task)(ch),
		receiver: (<-chan Task)(ch),
		name:     name,
		wg:       wg,
	}
}
