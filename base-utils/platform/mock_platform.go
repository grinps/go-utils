package platform

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"sync"
	"time"
)

// MockPlatform provides a mock implementation for testing
type MockPlatform struct {
	signalHandler  *MockSignalHandler
	envHandler     *MockEnvHandler
	fileHandler    *MockFileHandler
	processHandler *MockProcessHandler
	clock          *MockClock
}

// NewMockPlatform creates a new mock platform for testing
func NewMockPlatform() *MockPlatform {
	return &MockPlatform{
		signalHandler:  NewMockSignalHandler(),
		envHandler:     NewMockEnvHandler(),
		fileHandler:    NewMockFileHandler(),
		processHandler: NewMockProcessHandler(),
		clock:          NewMockClock(),
	}
}

func (p *MockPlatform) Signal() SignalHandler {
	return p.signalHandler
}

func (p *MockPlatform) Env() EnvHandler {
	return p.envHandler
}

func (p *MockPlatform) File() FileHandler {
	return p.fileHandler
}

func (p *MockPlatform) Process() ProcessHandler {
	return p.processHandler
}

func (p *MockPlatform) Clock() Clock {
	return p.clock
}

// MockSignalHandler provides a mock signal handler for testing
type MockSignalHandler struct {
	mu       sync.Mutex
	channels map[chan<- os.Signal][]os.Signal
}

func NewMockSignalHandler() *MockSignalHandler {
	return &MockSignalHandler{
		channels: make(map[chan<- os.Signal][]os.Signal),
	}
}

func (h *MockSignalHandler) Notify(c chan<- os.Signal, sig ...os.Signal) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.channels[c] = sig
}

func (h *MockSignalHandler) Stop(c chan<- os.Signal) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.channels, c)
}

func (h *MockSignalHandler) Ignore(sig ...os.Signal) {
	// Mock implementation - no-op
}

func (h *MockSignalHandler) Reset(sig ...os.Signal) {
	// Mock implementation - no-op
}

// SendSignal simulates sending a signal to registered channels
func (h *MockSignalHandler) SendSignal(sig os.Signal) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch, sigs := range h.channels {
		for _, s := range sigs {
			if s == sig {
				select {
				case ch <- sig:
				default:
				}
				break
			}
		}
	}
}

// MockEnvHandler provides a mock environment handler for testing
type MockEnvHandler struct {
	mu   sync.RWMutex
	vars map[string]string
}

func NewMockEnvHandler() *MockEnvHandler {
	return &MockEnvHandler{
		vars: make(map[string]string),
	}
}

func (h *MockEnvHandler) Getenv(key string) string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.vars[key]
}

func (h *MockEnvHandler) Setenv(key, value string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.vars[key] = value
	return nil
}

func (h *MockEnvHandler) Unsetenv(key string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.vars, key)
	return nil
}

func (h *MockEnvHandler) LookupEnv(key string) (string, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	val, ok := h.vars[key]
	return val, ok
}

func (h *MockEnvHandler) Environ() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make([]string, 0, len(h.vars))
	for k, v := range h.vars {
		result = append(result, k+"="+v)
	}
	return result
}

func (h *MockEnvHandler) Clearenv() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.vars = make(map[string]string)
}

func (h *MockEnvHandler) ExpandEnv(s string) string {
	// Simple implementation - could be enhanced
	return os.Expand(s, h.Getenv)
}

// MockFileHandler provides a mock file handler for testing
type MockFileHandler struct {
	mu    sync.RWMutex
	files map[string]*MockFile
	dirs  map[string]bool
	cwd   string
}

func NewMockFileHandler() *MockFileHandler {
	return &MockFileHandler{
		files: make(map[string]*MockFile),
		dirs:  make(map[string]bool),
		cwd:   "/",
	}
}

func (h *MockFileHandler) Open(name string) (File, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	f, ok := h.files[name]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return f.Clone(), nil
}

func (h *MockFileHandler) Create(name string) (File, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	f := NewMockFile(name)
	h.files[name] = f
	return f.Clone(), nil
}

func (h *MockFileHandler) OpenFile(name string, flag int, perm fs.FileMode) (File, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if flag&os.O_CREATE != 0 {
		f := NewMockFile(name)
		h.files[name] = f
		return f.Clone(), nil
	}
	f, ok := h.files[name]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return f.Clone(), nil
}

func (h *MockFileHandler) Remove(name string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.files, name)
	delete(h.dirs, name)
	return nil
}

func (h *MockFileHandler) RemoveAll(path string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	// Simple implementation - remove exact matches
	delete(h.files, path)
	delete(h.dirs, path)
	return nil
}

func (h *MockFileHandler) Rename(oldpath, newpath string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	f, ok := h.files[oldpath]
	if !ok {
		return fs.ErrNotExist
	}
	h.files[newpath] = f
	delete(h.files, oldpath)
	return nil
}

func (h *MockFileHandler) Stat(name string) (fs.FileInfo, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if _, ok := h.files[name]; ok {
		return &mockFileInfo{name: name, isDir: false}, nil
	}
	if _, ok := h.dirs[name]; ok {
		return &mockFileInfo{name: name, isDir: true}, nil
	}
	return nil, fs.ErrNotExist
}

func (h *MockFileHandler) Lstat(name string) (fs.FileInfo, error) {
	return h.Stat(name)
}

func (h *MockFileHandler) ReadFile(name string) ([]byte, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	f, ok := h.files[name]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return f.data.Bytes(), nil
}

func (h *MockFileHandler) WriteFile(name string, data []byte, perm fs.FileMode) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	f := NewMockFile(name)
	f.data.Write(data)
	h.files[name] = f
	return nil
}

func (h *MockFileHandler) Mkdir(name string, perm fs.FileMode) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.dirs[name] = true
	return nil
}

func (h *MockFileHandler) MkdirAll(path string, perm fs.FileMode) error {
	return h.Mkdir(path, perm)
}

func (h *MockFileHandler) ReadDir(name string) ([]fs.DirEntry, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if !h.dirs[name] {
		return nil, fs.ErrNotExist
	}
	return []fs.DirEntry{}, nil
}

func (h *MockFileHandler) Getwd() (string, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.cwd, nil
}

func (h *MockFileHandler) Chdir(dir string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cwd = dir
	return nil
}

func (h *MockFileHandler) TempDir() string {
	return "/tmp"
}

func (h *MockFileHandler) UserHomeDir() (string, error) {
	return "/home/user", nil
}

// MockFile implements the File interface for testing
type MockFile struct {
	mu     sync.RWMutex
	name   string
	data   *bytes.Buffer
	offset int64
	closed bool
}

func NewMockFile(name string) *MockFile {
	return &MockFile{
		name: name,
		data: &bytes.Buffer{},
	}
}

func (f *MockFile) Clone() *MockFile {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return &MockFile{
		name:   f.name,
		data:   bytes.NewBuffer(f.data.Bytes()),
		offset: 0,
		closed: false,
	}
}

func (f *MockFile) Read(p []byte) (n int, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return 0, fs.ErrClosed
	}
	return f.data.Read(p)
}

func (f *MockFile) Write(p []byte) (n int, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return 0, fs.ErrClosed
	}
	return f.data.Write(p)
}

func (f *MockFile) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	return nil
}

func (f *MockFile) Seek(offset int64, whence int) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return 0, fs.ErrClosed
	}
	// Simple implementation
	f.offset = offset
	return offset, nil
}

func (f *MockFile) ReadAt(p []byte, off int64) (n int, err error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if f.closed {
		return 0, fs.ErrClosed
	}
	data := f.data.Bytes()
	if off >= int64(len(data)) {
		return 0, io.EOF
	}
	n = copy(p, data[off:])
	if n < len(p) {
		err = io.EOF
	}
	return n, err
}

func (f *MockFile) WriteAt(p []byte, off int64) (n int, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return 0, fs.ErrClosed
	}
	// Simple implementation - not fully accurate
	return len(p), nil
}

func (f *MockFile) Name() string {
	return f.name
}

func (f *MockFile) Stat() (fs.FileInfo, error) {
	return &mockFileInfo{name: f.name, size: int64(f.data.Len())}, nil
}

func (f *MockFile) Sync() error {
	return nil
}

func (f *MockFile) Truncate(size int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return fs.ErrClosed
	}
	f.data.Truncate(int(size))
	return nil
}

func (f *MockFile) Chmod(mode fs.FileMode) error {
	return nil
}

func (f *MockFile) Chown(uid, gid int) error {
	return nil
}

func (f *MockFile) ReadDir(n int) ([]fs.DirEntry, error) {
	return []fs.DirEntry{}, nil
}

func (f *MockFile) Readdir(n int) ([]fs.FileInfo, error) {
	return []fs.FileInfo{}, nil
}

func (f *MockFile) Readdirnames(n int) ([]string, error) {
	return []string{}, nil
}

type mockFileInfo struct {
	name  string
	size  int64
	isDir bool
}

func (i *mockFileInfo) Name() string       { return i.name }
func (i *mockFileInfo) Size() int64        { return i.size }
func (i *mockFileInfo) Mode() fs.FileMode  { return 0644 }
func (i *mockFileInfo) ModTime() time.Time { return time.Now() }
func (i *mockFileInfo) IsDir() bool        { return i.isDir }
func (i *mockFileInfo) Sys() interface{}   { return nil }

// MockProcessHandler provides a mock process handler for testing
type MockProcessHandler struct {
	mu       sync.Mutex
	pid      int
	hostname string
	exitCode int
	exited   bool
}

func NewMockProcessHandler() *MockProcessHandler {
	return &MockProcessHandler{
		pid:      12345,
		hostname: "mock-host",
	}
}

func (h *MockProcessHandler) Getpid() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.pid
}

func (h *MockProcessHandler) Getppid() int {
	return 1
}

func (h *MockProcessHandler) Getuid() int {
	return 1000
}

func (h *MockProcessHandler) Geteuid() int {
	return 1000
}

func (h *MockProcessHandler) Getgid() int {
	return 1000
}

func (h *MockProcessHandler) Getegid() int {
	return 1000
}

func (h *MockProcessHandler) Exit(code int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.exitCode = code
	h.exited = true
}

func (h *MockProcessHandler) Hostname() (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.hostname, nil
}

func (h *MockProcessHandler) FindProcess(pid int) (Process, error) {
	return &MockProcess{pid: pid}, nil
}

func (h *MockProcessHandler) StartProcess(name string, argv []string, attr *os.ProcAttr) (Process, error) {
	return &MockProcess{pid: 99999}, nil
}

// SetPid sets the mock PID for testing
func (h *MockProcessHandler) SetPid(pid int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.pid = pid
}

// SetHostname sets the mock hostname for testing
func (h *MockProcessHandler) SetHostname(hostname string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.hostname = hostname
}

// MockProcess implements the Process interface for testing
type MockProcess struct {
	mu       sync.Mutex
	pid      int
	killed   bool
	released bool
	signals  []os.Signal
}

func (p *MockProcess) Kill() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.killed = true
	return nil
}

func (p *MockProcess) Signal(sig os.Signal) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.signals = append(p.signals, sig)
	return nil
}

func (p *MockProcess) Wait() (*os.ProcessState, error) {
	return nil, errors.New("mock wait not implemented")
}

func (p *MockProcess) Release() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.released = true
	return nil
}

func (p *MockProcess) Pid() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.pid
}

// MockClock provides a mock clock for testing time-dependent code
type MockClock struct {
	mu      sync.RWMutex
	current time.Time
}

func NewMockClock() *MockClock {
	return &MockClock{
		current: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func (c *MockClock) Now() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.current
}

func (c *MockClock) Sleep(d time.Duration) {
	c.Advance(d)
}

func (c *MockClock) After(d time.Duration) <-chan time.Time {
	ch := make(chan time.Time, 1)
	go func() {
		time.Sleep(1 * time.Millisecond) // Small delay to simulate async
		ch <- c.Now().Add(d)
	}()
	return ch
}

func (c *MockClock) Tick(d time.Duration) <-chan time.Time {
	ch := make(chan time.Time)
	// Simple mock - doesn't actually tick
	return ch
}

func (c *MockClock) NewTimer(d time.Duration) Timer {
	return &MockTimer{
		ch:       make(chan time.Time, 1),
		duration: d,
		clock:    c,
	}
}

func (c *MockClock) NewTicker(d time.Duration) Ticker {
	return &MockTicker{
		ch:       make(chan time.Time),
		duration: d,
	}
}

// Advance advances the mock clock by the given duration
func (c *MockClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.current = c.current.Add(d)
}

// Set sets the mock clock to a specific time
func (c *MockClock) Set(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.current = t
}

// MockTimer implements the Timer interface for testing
type MockTimer struct {
	mu       sync.Mutex
	ch       chan time.Time
	duration time.Duration
	stopped  bool
	clock    *MockClock
}

func (t *MockTimer) C() <-chan time.Time {
	return t.ch
}

func (t *MockTimer) Stop() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.stopped {
		return false
	}
	t.stopped = true
	return true
}

func (t *MockTimer) Reset(d time.Duration) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	wasActive := !t.stopped
	t.stopped = false
	t.duration = d
	return wasActive
}

// Fire manually fires the timer for testing
func (t *MockTimer) Fire() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.stopped {
		select {
		case t.ch <- t.clock.Now():
		default:
		}
	}
}

// MockTicker implements the Ticker interface for testing
type MockTicker struct {
	mu       sync.Mutex
	ch       chan time.Time
	duration time.Duration
	stopped  bool
}

func (t *MockTicker) C() <-chan time.Time {
	return t.ch
}

func (t *MockTicker) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.stopped = true
}

func (t *MockTicker) Reset(d time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.duration = d
}

// Tick manually ticks the ticker for testing
func (t *MockTicker) Tick(now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.stopped {
		select {
		case t.ch <- now:
		default:
		}
	}
}
