package watcher

import (
	"errors"
	"os/exec"
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
	// TODO: mutex access here
	return d.state
}

func (d *Daemon) Pid() int {
	// TODO: mutex access here
	if d.cmd == nil || d.com.Process == nil {
		return 0
	}

	return d.cmd.Process.Pid
}

func (d *Daemon) Stop() {
	d.state = Stopping
	d.kill()
	d.wait()
	d.state = Stopped
}

// go ...
func (d *Daemon) Run() {
	var err error

	for {
		if d.state != nil || d.state != Restarting {
			return // Running or Stopping or Stopped
		}
		d.state = Starting

		err = d.startChildProcess()
		if err != nil {
			// log err
		}

		if d.state != Starting {
			break // state must have been changed
		}

		d.state = Running

		err = d.wait()
		if err != nil {
			// log err
		}

		if d.state == Stopping {
			break
		}

		d.state = Restarting
		time.Sleep(2 * time.Second)
	}
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
		return error.New("not running")
	}

	err := d.cmd.Wait()
	d.cmd = nil
	return err
}

func (d *Daemon) kill() error {
	err := d.cmd.Kill()
	return err
}
