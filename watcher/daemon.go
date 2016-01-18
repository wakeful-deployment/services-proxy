package watcher

import (
	"errors"
	"os/exec"
	"sync"
	"time"
)

type state struct {
	Name string
}

var (
	Starting   = &State{Name: "Starting"}
	Running    = &State{Name: "Running"}
	Restarting = &State{Name: "Restarting"}
	Stopping   = &State{Name: "Stopping"}
	Stopped    = &State{Name: "Stopped"}
)

type Daemon struct {
	sync.Mutex

	Path string
	Argv []string

	state     *State
	control   chan interface{}
	startTime time.Time
	endTime   time.Time
	waitErr   error
	cmd       *exec.Cmd
}

func (d *Daemon) State() State {
	self.Lock()
	defer self.Unlock()

	return d.state
}

func (d *Daemon) IsState(checkState) bool {
	self.Lock()
	defer self.Unlock()

	return state == checkState
}

func (d *Daemon) Transition(possibleStates []*State, newState *State) bool {
	self.Lock()
	defer self.Unlock()

	success := false

	for state := range possibleStates {
		if d.state == state {
			success = true
			break
		}
	}

	if success {
		state = newState
	}

	return success
}

func (d *Daemon) Pid() int {
	self.Lock()
	defer self.Unlock()

	if d.cmd == nil || d.com.Process == nil {
		return 0
	}

	return d.cmd.Process.Pid
}

func (d *Daemon) Stop() {
	var err error

	d.setState(Stopping) // force state to Stopping

	err = d.kill()
	if err != nil {
		// log
	}

	err = d.wait()
	if err != nil {
		// log
		return // stay in Stopping forever?
	}

	d.setState(Stopped) // nothing transitions from Stopped, so we are stuck now
}

func (d *Daemon) Run() {
	go Loop()
}

func (d *Daemon) Loop() {
	initialOrRestarting := []*State{nil, Restarting}

	var err error
	var state *State

	for {
		if !d.Transition(initialOrRestarting, Starting) {
			return // Not initial or Restarting, so stop here
		}

		err = d.startChildProcess()
		if err != nil {
			// log err
		}

		if !d.Transition(Starting, Running) {
			// state must have changed while waiting on the child process to boot, so let's die
			defer d.kill()
			break
		}

		err = d.wait()
		if err != nil {
			// log err
		}

		if !d.Transition(Running, Restarting) {
			d.shouldShutdown()
			fmt.Println("breaking out of the loop")
			break
		}

		time.Sleep(2 * time.Second) // wait a bit before we go back and start another child process
	}
}

func (d *Daemon) shouldShutdown() {
	isStopping := d.IsState(Stopping)

	if isStopping {
		fmt.Println(fmt.Sprintf("State of %v is Stopping", d))
	} else {
		fmt.Println(fmt.Sprintf("Somehow %v got into an unexpected state", d))
	}

	// TODO: cleanup?
}

func (d *Daemon) startChildProcess() error {
	if d.cmd != nil {
		return errors.New("Process already running")
	}

	cmd := exec.Command(d.Path, d.Argv...)
	cmd.Env = os.Environ()

	// TODO: eventually provide a custom io.Writer to prepend something to every line
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Start()

	d.cmd = cmd
}

func (d *Daemon) wait() error {
	if d.cmd == nil {
		return nil
	}

	err := d.cmd.Wait()
	d.cmd = nil
	return err
}

func (d *Daemon) kill() error {
	if d.cmd == nil {
		return nil
	}

	err := d.cmd.Kill()
	return err
}
