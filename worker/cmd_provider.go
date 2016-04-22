package worker

import (
	"os"
	"os/exec"
	"strings"

	"github.com/anmitsu/go-shlex"
	"github.com/codeskyblue/go-sh"
)

type cmdConfig struct {
	name                        string
	upstreamURL, command        string
	workingDir, logDir, logFile string
	interval                    int
	env                         map[string]string
}

type cmdProvider struct {
	baseProvider
	cmdConfig
	command []string
	cmd     *exec.Cmd
	session *sh.Session
}

func newCmdProvider(c cmdConfig) (*cmdProvider, error) {
	// TODO: check config options
	provider := &cmdProvider{
		baseProvider: baseProvider{
			name:     c.name,
			ctx:      NewContext(),
			interval: c.interval,
		},
		cmdConfig: c,
	}

	provider.ctx.Set(_WorkingDirKey, c.workingDir)
	provider.ctx.Set(_LogDirKey, c.logDir)
	provider.ctx.Set(_LogFileKey, c.logFile)

	cmd, err := shlex.Split(c.command, true)
	if err != nil {
		return nil, err
	}
	provider.command = cmd

	return provider, nil
}

// Copied from go-sh
func newEnviron(env map[string]string, inherit bool) []string { //map[string]string {
	environ := make([]string, 0, len(env))
	if inherit {
		for _, line := range os.Environ() {
			// if os environment and env collapses,
			// omit the os one
			k := strings.Split(line, "=")[0]
			if _, ok := env[k]; ok {
				continue
			}
			environ = append(environ, line)
		}
	}
	for k, v := range env {
		environ = append(environ, k+"="+v)
	}
	return environ
}

// TODO: implement this
func (p *cmdProvider) Run() error {
	if len(p.command) == 1 {
		p.cmd = exec.Command(p.command[0])
	} else if len(p.command) > 1 {
		c := p.command[0]
		args := p.command[1:]
		p.cmd = exec.Command(c, args...)
	} else if len(p.command) == 0 {
		panic("Command length should be at least 1!")
	}
	p.cmd.Dir = p.WorkingDir()

	env := map[string]string{
		"TUNASYNC_MIRROR_NAME":  p.Name(),
		"TUNASYNC_WORKING_DIR":  p.WorkingDir(),
		"TUNASYNC_UPSTREAM_URL": p.upstreamURL,
		"TUNASYNC_LOG_FILE":     p.LogFile(),
	}
	for k, v := range p.env {
		env[k] = v
	}
	p.cmd.Env = newEnviron(env, true)

	logFile, err := os.OpenFile(p.LogFile(), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	p.cmd.Stdout = logFile
	p.cmd.Stderr = logFile

	return p.cmd.Start()
}

func (p *cmdProvider) Wait() error {
	return p.cmd.Wait()
}

// TODO: implement this
func (p *cmdProvider) Terminate() {

}

// TODO: implement this
func (p *cmdProvider) Hooks() {

}
