package ci

import (
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/internal/dg"
	"github.com/srevinsaju/togomak/v1/internal/ui"
	"github.com/srevinsaju/togomak/v1/internal/x"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Tracker struct {
	runnables   Blocks
	runnablesMu sync.Mutex
	runnablesWg sync.WaitGroup

	daemons   Blocks
	daemonsMu sync.Mutex
	daemonsWg sync.WaitGroup

	completed       Blocks
	completedMu     sync.Mutex
	completedSignal chan Block

	killSignal      chan os.Signal
	interruptSignal chan os.Signal
}

func NewTracker() *Tracker {
	return &Tracker{
		completedSignal: make(chan Block, 1),

		killSignal:      make(chan os.Signal, 1),
		interruptSignal: make(chan os.Signal, 1),
	}
}

func (t *Tracker) AppendRunnable(runnable Block) {
	t.runnablesWg.Add(1)
	t.runnablesMu.Lock()
	defer t.runnablesMu.Unlock()
	t.runnables = append(t.runnables, runnable)
}

func (t *Tracker) RunnableWait() {
	t.runnablesWg.Wait()
}

func (t *Tracker) RunnableDone() {
	t.runnablesWg.Done()
}

func (t *Tracker) AppendDaemon(daemon Block) {
	t.daemonsWg.Add(1)
	t.daemonsMu.Lock()
	defer t.daemonsMu.Unlock()
	t.daemons = append(t.daemons, daemon)
}

func (t *Tracker) DaemonWait() {
	t.daemonsWg.Wait()
}

func (t *Tracker) DaemonDone() {
	t.daemonsWg.Done()
}

func (t *Tracker) HasDaemons() bool {
	return len(t.daemons) > 0
}

func (t *Tracker) AppendCompleted(completed Block) {
	t.completedMu.Lock()
	defer t.completedMu.Unlock()
	t.completed = append(t.completed, completed)
	t.completedSignal <- completed
}

type Handler struct {
	Tracker *Tracker
	Diags   *dg.SafeDiagnostics
	Logger  logrus.Ext1FieldLogger
	Process *HandlerProcess

	diagWriter hcl.DiagnosticWriter
	ctxMu      sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

func (h *Handler) Context() context.Context {
	h.ctxMu.RLock()
	defer h.ctxMu.RUnlock()
	return h.ctx
}

type HandlerOption func(*Handler)

func WithContext(ctx context.Context) HandlerOption {
	return func(h *Handler) {
		h.ctxMu.RLock()
		ctx, cancel := context.WithCancel(ctx)
		h.ctxMu.RUnlock()
		h.ctxMu.Lock()
		h.ctx = ctx
		h.cancel = cancel
		h.ctxMu.Unlock()
	}
}

func WithLogger(logger logrus.Ext1FieldLogger) HandlerOption {
	return func(h *Handler) {
		h.Logger = logger
	}
}

func WithDiagnostics(diagnostics *dg.SafeDiagnostics) HandlerOption {
	return func(h *Handler) {
		h.Diags = diagnostics
	}
}

func WithDiagnosticWriter(diagnosticWriter hcl.DiagnosticWriter) HandlerOption {
	return func(h *Handler) {
		h.diagWriter = diagnosticWriter
	}
}

func WithTracker(tracker *Tracker) HandlerOption {
	return func(h *Handler) {
		h.Tracker = tracker
	}
}

func WithProcessBootTime(bootTime time.Time) HandlerOption {
	return func(h *Handler) {
		h.Process.BootTime = bootTime
	}
}

type HandlerProcess struct {
	BootTime time.Time
}

func NewHandlerProcess() *HandlerProcess {
	return &HandlerProcess{
		BootTime: time.Now(),
	}
}

func NewHandler(opts ...HandlerOption) *Handler {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	h := &Handler{
		Tracker: NewTracker(),
		Diags:   &dg.SafeDiagnostics{},
		Logger:  logrus.New(),
		Process: NewHandlerProcess(),

		diagWriter: hcl.NewDiagnosticTextWriter(os.Stdout, nil, 0, true),
		ctx:        ctx,
		cancel:     cancel,
	}

	h = h.Update(opts...)
	return h
}

func (h *Handler) Update(opts ...HandlerOption) *Handler {
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *Handler) Ptr() *Handler {
	return h
}

func (h *Handler) Kill() {
	signal.Notify(h.Tracker.killSignal, os.Kill)
	ctx := h.Context()
	logger := h.Logger.WithField("orchestra", "watchdog")
	select {
	case <-h.Tracker.killSignal:
		var diags hcl.Diagnostics
		logger.Warn("received kill signal, killing all subprocesses")
		logger.Warn("stopping running operations...")

		for _, runnable := range h.Tracker.runnables {
			logger.Debugf("killing runnable %s", runnable.Identifier())
			d := runnable.Kill()
			diags = diags.Extend(d)
		}

		h.cancel()
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Force quit",
			Detail:   "data loss may have occurred",
		})
		if diags.HasErrors() {
			writer := hcl.NewDiagnosticTextWriter(os.Stderr, nil, 78, true)
			_ = writer.WriteDiagnostics(diags)
		}
		os.Exit(h.Fatal())
	case <-ctx.Done():
		logger.Tracef("took %s to complete the pipeline", time.Since(h.Process.BootTime))
		return
	}
}

func (h *Handler) Daemons() {
	logger := h.Logger.WithField("orchestra", "watchdog")
	var completedRunnables Blocks

	defer h.WriteDiagnostics()
	logger.Tracef("starting watchdog")

	// execute the following function when we receive any message on the completed channel
	for {
		c := <-h.Tracker.completedSignal
		logger.Debugf("received completed runnable, %s", c.Identifier())
		completedRunnables = append(completedRunnables, c)

		daemons := h.Tracker.daemons
		for _, daemon := range daemons {
			if daemon.Terminated() {
				continue
			}
			logger.Tracef("checking daemon %s", daemon.Identifier())
			lifecycle, d := daemon.ExecutionOptions(h.Context())
			if d.HasErrors() {
				h.Diags.Extend(d)
				d := daemon.Terminate(nil, false)
				h.Diags.Extend(d)
				return
			}
			if lifecycle == nil {
				continue
			}

			allCompleted := true
			for _, block := range lifecycle.StopWhenComplete {
				logger.Tracef("checking daemon %s, requires block %s to complete", daemon.Identifier(), block.Identifier())
				completed := false
				for _, completedBlocks := range completedRunnables {
					if block.Identifier() == completedBlocks.Identifier() {
						completed = true
						break
					}
				}
				if !completed {
					allCompleted = false
					break
				}
			}
			if allCompleted {
				logger.Infof("stopping daemon %s", daemon.Identifier())
				d := daemon.Terminate(nil, true)
				if d.HasErrors() {
					h.Diags.Extend(d)
				}
			}
		}
	}
}

func (h *Handler) Interrupt() {
	signal.Notify(h.Tracker.interruptSignal, os.Interrupt)
	signal.Notify(h.Tracker.interruptSignal, syscall.SIGTERM)

	ctx := h.Context()
	logger := h.Logger.WithField("orchestra", "watchdog")
	select {
	case <-h.Tracker.interruptSignal:
		var diags hcl.Diagnostics
		logger.Warn("received interrupt signal, cancelling the pipeline")
		logger.Warn("stopping running operations...")
		logger.Warn("press CTRL+C again to force quit")

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)
		go func() {
			<-ch
			logger.Warn("Two interrupts received. Exiting immediately.")
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Force quit",
				Detail:   "data loss may have occurred",
			})
			os.Exit(h.Fatal())
			return
		}()
		for _, runnable := range h.Tracker.runnables {
			logger.Debugf("stopping runnable %s", runnable.Identifier())
			d := runnable.Terminate(nil, false)
			diags = diags.Extend(d)
		}

		if diags.HasErrors() {
			writer := hcl.NewDiagnosticTextWriter(os.Stderr, nil, 78, true)
			_ = writer.WriteDiagnostics(diags)
			os.Exit(h.Fatal())
		}
		h.cancel()
	case <-ctx.Done():
		logger.Debugf("took %s to complete the pipeline", time.Since(h.Process.BootTime))
		return
	}
}

func (h *Handler) WriteDiagnostics() {
	if h.Diags.Diagnostics() == nil {
		return
	}
	x.Must(h.diagWriter.WriteDiagnostics(h.Diags.Diagnostics()))
}

func (h *Handler) finale(logLevel logrus.Level) {
	message := ui.Grey(fmt.Sprintf("took %s", time.Since(h.Process.BootTime).Round(time.Millisecond)))
	switch logLevel {
	case logrus.ErrorLevel:
		h.Logger.Error(message)
	case logrus.InfoLevel:
		h.Logger.Info(message)
	default:
		panic("invalid log level")
	}
}

func (h *Handler) Fatal() int {
	h.finale(logrus.ErrorLevel)
	return 1
}

func (h *Handler) Ok() int {
	h.finale(logrus.InfoLevel)
	return 0
}
