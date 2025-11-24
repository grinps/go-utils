package platform

import (
	"io"
	"io/fs"
	"os"
	"time"
)

// Platform provides an abstraction layer over OS-level operations
// to enable testing and alternative implementations
type Platform interface {
	Signal() SignalHandler
	Env() EnvHandler
	File() FileHandler
	Process() ProcessHandler
	Clock() Clock
}

// SignalHandler provides signal handling operations
type SignalHandler interface {
	// Notify causes package signal to relay incoming signals to c
	Notify(c chan<- os.Signal, sig ...os.Signal)
	// Stop causes package signal to stop relaying incoming signals to c
	Stop(c chan<- os.Signal)
	// Ignore causes the provided signals to be ignored
	Ignore(sig ...os.Signal)
	// Reset undoes the effect of any prior calls to Notify for the provided signals
	Reset(sig ...os.Signal)
}

// EnvHandler provides environment variable operations
type EnvHandler interface {
	// Getenv retrieves the value of the environment variable named by the key
	Getenv(key string) string
	// Setenv sets the value of the environment variable named by the key
	Setenv(key, value string) error
	// Unsetenv unsets a single environment variable
	Unsetenv(key string) error
	// LookupEnv retrieves the value of the environment variable named by the key
	LookupEnv(key string) (string, bool)
	// Environ returns a copy of strings representing the environment
	Environ() []string
	// Clearenv deletes all environment variables
	Clearenv()
	// ExpandEnv replaces ${var} or $var in the string according to the values of the current environment variables
	ExpandEnv(s string) string
}

// FileHandler provides file system operations
type FileHandler interface {
	// Open opens the named file for reading
	Open(name string) (File, error)
	// Create creates or truncates the named file
	Create(name string) (File, error)
	// OpenFile is the generalized open call
	OpenFile(name string, flag int, perm fs.FileMode) (File, error)
	// Remove removes the named file or directory
	Remove(name string) error
	// RemoveAll removes path and any children it contains
	RemoveAll(path string) error
	// Rename renames (moves) oldpath to newpath
	Rename(oldpath, newpath string) error
	// Stat returns a FileInfo describing the named file
	Stat(name string) (fs.FileInfo, error)
	// Lstat returns a FileInfo describing the named file without following symbolic links
	Lstat(name string) (fs.FileInfo, error)
	// ReadFile reads the named file and returns the contents
	ReadFile(name string) ([]byte, error)
	// WriteFile writes data to the named file, creating it if necessary
	WriteFile(name string, data []byte, perm fs.FileMode) error
	// Mkdir creates a new directory with the specified name and permission bits
	Mkdir(name string, perm fs.FileMode) error
	// MkdirAll creates a directory named path, along with any necessary parents
	MkdirAll(path string, perm fs.FileMode) error
	// ReadDir reads the named directory and returns all its directory entries sorted by filename
	ReadDir(name string) ([]fs.DirEntry, error)
	// Getwd returns a rooted path name corresponding to the current directory
	Getwd() (dir string, err error)
	// Chdir changes the current working directory to the named directory
	Chdir(dir string) error
	// TempDir returns the default directory to use for temporary files
	TempDir() string
	// UserHomeDir returns the current user's home directory
	UserHomeDir() (string, error)
}

// File represents an open file descriptor
type File interface {
	io.Reader
	io.Writer
	io.Closer
	io.Seeker
	io.ReaderAt
	io.WriterAt

	// Name returns the name of the file
	Name() string
	// Stat returns the FileInfo structure describing file
	Stat() (fs.FileInfo, error)
	// Sync commits the current contents of the file to stable storage
	Sync() error
	// Truncate changes the size of the file
	Truncate(size int64) error
	// Chmod changes the mode of the file
	Chmod(mode fs.FileMode) error
	// Chown changes the numeric uid and gid of the named file
	Chown(uid, gid int) error
	// ReadDir reads the contents of the directory and returns a slice of up to n DirEntry values
	ReadDir(n int) ([]fs.DirEntry, error)
	// Readdir reads the contents of the directory and returns a slice of up to n FileInfo values
	Readdir(n int) ([]fs.FileInfo, error)
	// Readdirnames reads the contents of the directory and returns a slice of up to n names
	Readdirnames(n int) ([]string, error)
}

// ProcessHandler provides process-related operations
type ProcessHandler interface {
	// Getpid returns the process id of the caller
	Getpid() int
	// Getppid returns the process id of the caller's parent
	Getppid() int
	// Getuid returns the numeric user id of the caller
	Getuid() int
	// Geteuid returns the numeric effective user id of the caller
	Geteuid() int
	// Getgid returns the numeric group id of the caller
	Getgid() int
	// Getegid returns the numeric effective group id of the caller
	Getegid() int
	// Exit causes the current program to exit with the given status code
	Exit(code int)
	// Hostname returns the host name reported by the kernel
	Hostname() (name string, err error)
	// FindProcess looks for a running process by its pid
	FindProcess(pid int) (Process, error)
	// StartProcess starts a new process with the program, arguments and attributes specified
	StartProcess(name string, argv []string, attr *os.ProcAttr) (Process, error)
}

// Process represents an OS process
type Process interface {
	// Kill causes the Process to exit immediately
	Kill() error
	// Signal sends a signal to the Process
	Signal(sig os.Signal) error
	// Wait waits for the Process to exit, and then returns a ProcessState describing its status
	Wait() (*os.ProcessState, error)
	// Release releases any resources associated with the Process
	Release() error
	// Pid returns the process id
	Pid() int
}

// Clock provides time-related operations for better testability
type Clock interface {
	// Now returns the current time
	Now() time.Time
	// Sleep pauses the current goroutine for at least the duration d
	Sleep(d time.Duration)
	// After waits for the duration to elapse and then sends the current time on the returned channel
	After(d time.Duration) <-chan time.Time
	// Tick returns a channel that delivers ticks of a clock at intervals
	Tick(d time.Duration) <-chan time.Time
	// NewTimer creates a new Timer that will send the current time on its channel after at least duration d
	NewTimer(d time.Duration) Timer
	// NewTicker returns a new Ticker containing a channel that will send the current time on the channel after each tick
	NewTicker(d time.Duration) Ticker
}

// Timer represents a single event
type Timer interface {
	// C returns the channel on which the time is delivered
	C() <-chan time.Time
	// Stop prevents the Timer from firing
	Stop() bool
	// Reset changes the timer to expire after duration d
	Reset(d time.Duration) bool
}

// Ticker holds a channel that delivers ticks of a clock at intervals
type Ticker interface {
	// C returns the channel on which the ticks are delivered
	C() <-chan time.Time
	// Stop turns off a ticker
	Stop()
	// Reset stops a ticker and resets its period to the specified duration
	Reset(d time.Duration)
}
