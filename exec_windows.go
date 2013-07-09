package exec

func (c *Cmd) KillAll() (err error) {
	return c.Process.Kill()
}
