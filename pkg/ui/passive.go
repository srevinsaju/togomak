package ui

import "github.com/sirupsen/logrus"

type PassiveProgressBar struct {
	Logger  *logrus.Entry
	Message string

	pb        *ProgressWriter
	completer chan bool
}

func NewPassiveProgressBar(logger *logrus.Entry, message string) *PassiveProgressBar {
	return &PassiveProgressBar{
		Logger:  logger,
		Message: message,
	}
}

func (p *PassiveProgressBar) Init() {
	p.completer = make(chan bool)
	go p.run()
}

func (p *PassiveProgressBar) run() {
	p.pb = NewProgressWriter(p.Logger, p.Message)
	for {
		select {
		case <-p.completer:
			p.pb.Close()
			return
		default:
			p.pb.printProgress()
		}
	}
}

func (p *PassiveProgressBar) Done() {
	p.completer <- true
}
