package platform

import (
	"io/fs"
	"os"
	"os/signal"
	"time"
)

// osPlatform is the default implementation that delegates to the actual OS
type osPlatform struct {
	signalHandler  SignalHandler
	envHandler     EnvHandler
	fileHandler    FileHandler
	processHandler ProcessHandler
	clock          Clock
}

// NewOSPlatform creates a new platform implementation that uses the actual OS
func NewOSPlatform() Platform {
	return &osPlatform{
		signalHandler:  &osSignalHandler{},
		envHandler:     &osEnvHandler{},
		fileHandler:    &osFileHandler{},
		processHandler: &osProcessHandler{},
		clock:          &osClock{},
	}
}

func (p *osPlatform) Signal() SignalHandler {
	return p.signalHandler
}

func (p *osPlatform) Env() EnvHandler {
	return p.envHandler
}

func (p *osPlatform) File() FileHandler {
	return p.fileHandler
}

func (p *osPlatform) Process() ProcessHandler {
	return p.processHandler
}

func (p *osPlatform) Clock() Clock {
	return p.clock
}

// osSignalHandler wraps the os/signal package
type osSignalHandler struct{}

func (h *osSignalHandler) Notify(c chan<- os.Signal, sig ...os.Signal) {
	signal.Notify(c, sig...)
}

func (h *osSignalHandler) Stop(c chan<- os.Signal) {
	signal.Stop(c)
}

func (h *osSignalHandler) Ignore(sig ...os.Signal) {
	signal.Ignore(sig...)
}

func (h *osSignalHandler) Reset(sig ...os.Signal) {
	signal.Reset(sig...)
}

// osEnvHandler wraps the os environment functions
type osEnvHandler struct{}

func (h *osEnvHandler) Getenv(key string) string {
	return os.Getenv(key)
}

func (h *osEnvHandler) Setenv(key, value string) error {
	return os.Setenv(key, value)
}

func (h *osEnvHandler) Unsetenv(key string) error {
	return os.Unsetenv(key)
}

func (h *osEnvHandler) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

func (h *osEnvHandler) Environ() []string {
	return os.Environ()
}

func (h *osEnvHandler) Clearenv() {
	os.Clearenv()
}

func (h *osEnvHandler) ExpandEnv(s string) string {
	return os.ExpandEnv(s)
}

// osFileHandler wraps the os file functions
type osFileHandler struct{}

func (h *osFileHandler) Open(name string) (File, error) {
	return os.Open(name)
}

func (h *osFileHandler) Create(name string) (File, error) {
	return os.Create(name)
}

func (h *osFileHandler) OpenFile(name string, flag int, perm fs.FileMode) (File, error) {
	return os.OpenFile(name, flag, perm)
}

func (h *osFileHandler) Remove(name string) error {
	return os.Remove(name)
}

func (h *osFileHandler) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (h *osFileHandler) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func (h *osFileHandler) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (h *osFileHandler) Lstat(name string) (fs.FileInfo, error) {
	return os.Lstat(name)
}

func (h *osFileHandler) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (h *osFileHandler) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (h *osFileHandler) Mkdir(name string, perm fs.FileMode) error {
	return os.Mkdir(name, perm)
}

func (h *osFileHandler) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (h *osFileHandler) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(name)
}

func (h *osFileHandler) Getwd() (string, error) {
	return os.Getwd()
}

func (h *osFileHandler) Chdir(dir string) error {
	return os.Chdir(dir)
}

func (h *osFileHandler) TempDir() string {
	return os.TempDir()
}

func (h *osFileHandler) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

// osProcessHandler wraps the os process functions
type osProcessHandler struct{}

func (h *osProcessHandler) Getpid() int {
	return os.Getpid()
}

func (h *osProcessHandler) Getppid() int {
	return os.Getppid()
}

func (h *osProcessHandler) Getuid() int {
	return os.Getuid()
}

func (h *osProcessHandler) Geteuid() int {
	return os.Geteuid()
}

func (h *osProcessHandler) Getgid() int {
	return os.Getgid()
}

func (h *osProcessHandler) Getegid() int {
	return os.Getegid()
}

func (h *osProcessHandler) Exit(code int) {
	os.Exit(code)
}

func (h *osProcessHandler) Hostname() (string, error) {
	return os.Hostname()
}

func (h *osProcessHandler) FindProcess(pid int) (Process, error) {
	p, err := os.FindProcess(pid)
	if err != nil {
		return nil, err
	}
	return &osProcess{process: p}, nil
}

func (h *osProcessHandler) StartProcess(name string, argv []string, attr *os.ProcAttr) (Process, error) {
	p, err := os.StartProcess(name, argv, attr)
	if err != nil {
		return nil, err
	}
	return &osProcess{process: p}, nil
}

// osProcess wraps os.Process to implement the Process interface
type osProcess struct {
	process *os.Process
}

func (p *osProcess) Kill() error {
	return p.process.Kill()
}

func (p *osProcess) Signal(sig os.Signal) error {
	return p.process.Signal(sig)
}

func (p *osProcess) Wait() (*os.ProcessState, error) {
	return p.process.Wait()
}

func (p *osProcess) Release() error {
	return p.process.Release()
}

func (p *osProcess) Pid() int {
	return p.process.Pid
}

// osClock wraps the time package
type osClock struct{}

func (c *osClock) Now() time.Time {
	return time.Now()
}

func (c *osClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

func (c *osClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

func (c *osClock) Tick(d time.Duration) <-chan time.Time {
	// Note: time.Tick leaks the underlying ticker.
	// This is only safe for use in endless functions, tests, and main package.
	// Consider using NewTicker() instead for production code.
	return time.Tick(d)
}

func (c *osClock) NewTimer(d time.Duration) Timer {
	return &osTimer{timer: time.NewTimer(d)}
}

func (c *osClock) NewTicker(d time.Duration) Ticker {
	return &osTicker{ticker: time.NewTicker(d)}
}

// osTimer wraps time.Timer
type osTimer struct {
	timer *time.Timer
}

func (t *osTimer) C() <-chan time.Time {
	return t.timer.C
}

func (t *osTimer) Stop() bool {
	return t.timer.Stop()
}

func (t *osTimer) Reset(d time.Duration) bool {
	return t.timer.Reset(d)
}

// osTicker wraps time.Ticker
type osTicker struct {
	ticker *time.Ticker
}

func (t *osTicker) C() <-chan time.Time {
	return t.ticker.C
}

func (t *osTicker) Stop() {
	t.ticker.Stop()
}

func (t *osTicker) Reset(d time.Duration) {
	t.ticker.Reset(d)
}
