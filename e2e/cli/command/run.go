package command

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	capi "github.com/hashicorp/consul/api"
	hclog "github.com/hashicorp/go-hclog"
	vapi "github.com/hashicorp/vault/api"
	"github.com/mitchellh/cli"
)

func RunCommandFactory(ui cli.Ui, logger hclog.Logger) cli.CommandFactory {
	return func() (cli.Command, error) {
		meta := Meta{
			Ui:     ui,
			logger: logger,
		}
		return &Run{Meta: meta}, nil
	}
}

type Run struct {
	Meta
}

func (c *Run) Help() string {
	helpText := `
Usage: nomad-e2e run
`
	return strings.TrimSpace(helpText)
}

func (c *Run) Synopsis() string {
	return "Runs the e2e test suite"
}

func (c *Run) Run(args []string) int {
	var envPath string
	var nomadBinary string
	var tfPath string
	var slow bool
	var run string
	cmdFlags := c.FlagSet("run")
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }
	cmdFlags.StringVar(&envPath, "env-path", "./environments/", "Path to e2e environment terraform configs")
	cmdFlags.StringVar(&nomadBinary, "nomad-binary", "", "")
	cmdFlags.StringVar(&tfPath, "tf-path", "", "")
	cmdFlags.StringVar(&run, "run", "", "Regex to target specific test suites/cases")
	cmdFlags.BoolVar(&slow, "slow", false, "Toggle slow running suites")

	if err := cmdFlags.Parse(args); err != nil {
		c.logger.Error("failed to parse flags", "error", err)
		return 1
	}
	if c.verbose {
		c.logger.SetLevel(hclog.Debug)
	}

	args = cmdFlags.Args()

	if len(args) == 0 {
		c.logger.Info("no environments specified, running test suite locally")
		var report *TestReport
		var err error
		if report, err = c.run(&runOpts{
			slow:    slow,
			verbose: c.verbose,
		}); err != nil {
			c.logger.Error("failed to run test suite", "error", err)
			return 1
		}
		if report.TotalFailedTests == 0 {
			c.Ui.Output("PASSED!")
			if c.verbose {
				c.Ui.Output(report.Summary())
			}
		} else {
			c.Ui.Output("***FAILED***")
			c.Ui.Output(report.Summary())
		}
		return 0
	}

	environments := []*environment{}
	for _, e := range args {
		if len(strings.Split(e, "/")) != 2 {
			c.logger.Error("argument should be formated as <provider>/<environment>", "args", e)
			return 1
		}
		envs, err := envsFromGlob(envPath, e, tfPath, c.logger)
		if err != nil {
			c.logger.Error("failed to build environment", "environment", e, "error", err)
			return 1
		}
		environments = append(environments, envs...)

	}
	envCount := len(environments)
	// Use go-getter to fetch the nomad binary
	nomadPath, err := fetchBinary(nomadBinary)
	defer os.RemoveAll(nomadPath)
	if err != nil {
		c.logger.Error("failed to fetch nomad binary", "error", err)
		return 1
	}

	c.logger.Debug("starting tests", "totalEnvironments", envCount)
	for i, env := range environments {
		logger := c.logger.With("name", env.name, "provider", env.provider)
		logger.Debug("provisioning environment")
		results, err := env.provision(nomadPath)
		if err != nil {
			logger.Error("failed to provision environment", "error", err)
			return 1
		}

		opts := &runOpts{
			provider:   env.provider,
			env:        env.name,
			slow:       slow,
			verbose:    c.verbose,
			nomadAddr:  results.nomadAddr,
			consulAddr: results.consulAddr,
			vaultAddr:  results.vaultAddr,
		}

		var report *TestReport
		if report, err = c.run(opts); err != nil {
			logger.Error("failed to run tests against environment", "error", err)
			return 1
		}
		if report.TotalFailedTests == 0 {

			c.Ui.Output(fmt.Sprintf("[%d/%d] %s/%s: PASSED!\n", i+1, envCount, env.provider, env.name))
			if c.verbose {
				c.Ui.Output(fmt.Sprintf("[%d/%d] %s/%s: %s", i+1, envCount, env.provider, env.name, report.Summary()))
			}
		} else {
			c.Ui.Output(fmt.Sprintf("[%d/%d] %s/%s: ***FAILED***\n", i+1, envCount, env.provider, env.name))
			c.Ui.Output(fmt.Sprintf("[%d/%d] %s/%s: %s", i+1, envCount, env.provider, env.name, report.Summary()))
		}
	}
	return 0
}

func (c *Run) run(opts *runOpts) (*TestReport, error) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(goBin, opts.goArgs()...)
	cmd.Env = opts.goEnv()
	out, err := cmd.StdoutPipe()
	defer out.Close()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	dec := NewDecoder(out)
	report, err := dec.Decode(c.logger.Named("run.gotest"))
	if err != nil {
		return nil, err
	}

	return report, nil

}

type runOpts struct {
	nomadAddr  string
	consulAddr string
	vaultAddr  string
	provider   string
	env        string
	local      bool
	slow       bool
	verbose    bool
}

func (opts *runOpts) goArgs() []string {
	a := []string{
		"test",
		"-json",
		"github.com/hashicorp/nomad/e2e",
		"-env=" + opts.env,
		"-env.provider=" + opts.provider,
	}

	if opts.slow {
		a = append(a, "-slow")
	}

	if opts.local {
		a = append(a, "-local")
	}
	return a
}

func (opts *runOpts) goEnv() []string {
	env := append(os.Environ(), "NOMAD_E2E=1")
	if opts.nomadAddr != "" {
		env = append(env, "NOMAD_ADDR="+opts.nomadAddr)
	}
	if opts.consulAddr != "" {
		env = append(env, fmt.Sprintf("%s=%s", capi.HTTPAddrEnvName, opts.consulAddr))
	}
	if opts.vaultAddr != "" {
		env = append(env, fmt.Sprintf("%s=%s", vapi.EnvVaultAddress, opts.consulAddr))
	}

	return env
}