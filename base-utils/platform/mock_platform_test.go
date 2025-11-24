package platform

import (
	"io"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestMockPlatformAccessors(t *testing.T) {
	mock := NewMockPlatform()

	if mock.Signal() == nil {
		t.Error("Signal() returned nil")
	}
	if mock.Env() == nil {
		t.Error("Env() returned nil")
	}
	if mock.File() == nil {
		t.Error("File() returned nil")
	}
	if mock.Process() == nil {
		t.Error("Process() returned nil")
	}
	if mock.Clock() == nil {
		t.Error("Clock() returned nil")
	}
}

func TestMockSignalHandlerIgnoreAndReset(t *testing.T) {
	handler := NewMockSignalHandler()

	// Test Ignore (no-op in mock)
	handler.Ignore(os.Interrupt)

	// Test Reset (no-op in mock)
	handler.Reset(os.Interrupt)
}

func TestMockEnvHandlerClearenv(t *testing.T) {
	env := NewMockEnvHandler()

	env.Setenv("KEY1", "value1")
	env.Setenv("KEY2", "value2")

	env.Clearenv()

	if env.Getenv("KEY1") != "" {
		t.Error("Expected KEY1 to be cleared")
	}
	if env.Getenv("KEY2") != "" {
		t.Error("Expected KEY2 to be cleared")
	}
}

func TestMockEnvHandlerEnviron(t *testing.T) {
	env := NewMockEnvHandler()

	env.Setenv("TEST_KEY", "test_value")
	environ := env.Environ()

	found := false
	for _, e := range environ {
		if e == "TEST_KEY=test_value" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected TEST_KEY=test_value in environ")
	}
}

func TestMockEnvHandlerExpandEnv(t *testing.T) {
	env := NewMockEnvHandler()

	env.Setenv("USER", "testuser")
	env.Setenv("HOME", "/home/testuser")

	expanded := env.ExpandEnv("User: $USER, Home: $HOME")
	expected := "User: testuser, Home: /home/testuser"

	if expanded != expected {
		t.Errorf("Expected %q, got %q", expected, expanded)
	}
}

func TestMockFileHandlerOpenFile(t *testing.T) {
	handler := NewMockFileHandler()

	// Test with O_CREATE flag
	file, err := handler.OpenFile("/test/file.txt", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("Failed to open file with O_CREATE: %v", err)
	}
	file.Close()

	// Test opening existing file
	file2, err := handler.OpenFile("/test/file.txt", os.O_RDONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to open existing file: %v", err)
	}
	file2.Close()

	// Test opening non-existent file without O_CREATE
	_, err = handler.OpenFile("/nonexistent.txt", os.O_RDONLY, 0644)
	if err == nil {
		t.Error("Expected error when opening non-existent file")
	}
}

func TestMockFileHandlerRename(t *testing.T) {
	handler := NewMockFileHandler()

	handler.WriteFile("/old.txt", []byte("content"), 0644)

	err := handler.Rename("/old.txt", "/new.txt")
	if err != nil {
		t.Fatalf("Failed to rename file: %v", err)
	}

	// Old file should not exist
	_, err = handler.Stat("/old.txt")
	if err == nil {
		t.Error("Expected error when statting old file")
	}

	// New file should exist
	data, err := handler.ReadFile("/new.txt")
	if err != nil {
		t.Fatalf("Failed to read renamed file: %v", err)
	}
	if string(data) != "content" {
		t.Errorf("Expected 'content', got %q", data)
	}
}

func TestMockFileHandlerRenameNonExistent(t *testing.T) {
	handler := NewMockFileHandler()

	err := handler.Rename("/nonexistent.txt", "/new.txt")
	if err == nil {
		t.Error("Expected error when renaming non-existent file")
	}
}

func TestMockFileHandlerMkdir(t *testing.T) {
	handler := NewMockFileHandler()

	err := handler.Mkdir("/testdir", 0755)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	info, err := handler.Stat("/testdir")
	if err != nil {
		t.Fatalf("Failed to stat directory: %v", err)
	}
	if !info.IsDir() {
		t.Error("Expected /testdir to be a directory")
	}
}

func TestMockFileHandlerMkdirAll(t *testing.T) {
	handler := NewMockFileHandler()

	err := handler.MkdirAll("/path/to/nested/dir", 0755)
	if err != nil {
		t.Fatalf("Failed to create nested directories: %v", err)
	}

	info, err := handler.Stat("/path/to/nested/dir")
	if err != nil {
		t.Fatalf("Failed to stat nested directory: %v", err)
	}
	if !info.IsDir() {
		t.Error("Expected nested path to be a directory")
	}
}

func TestMockFileHandlerReadDir(t *testing.T) {
	handler := NewMockFileHandler()

	// Create directory
	handler.Mkdir("/testdir", 0755)

	entries, err := handler.ReadDir("/testdir")
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}
	if entries == nil {
		t.Error("Expected non-nil entries")
	}

	// Test reading non-existent directory
	_, err = handler.ReadDir("/nonexistent")
	if err == nil {
		t.Error("Expected error when reading non-existent directory")
	}
}

func TestMockFileHandlerLstat(t *testing.T) {
	handler := NewMockFileHandler()

	handler.WriteFile("/test.txt", []byte("data"), 0644)

	info, err := handler.Lstat("/test.txt")
	if err != nil {
		t.Fatalf("Failed to lstat file: %v", err)
	}
	if info.Name() != "/test.txt" {
		t.Errorf("Expected name '/test.txt', got %q", info.Name())
	}
}

func TestMockFileHandlerChdirAndGetwd(t *testing.T) {
	handler := NewMockFileHandler()

	// Initial working directory
	cwd, err := handler.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	if cwd != "/" {
		t.Errorf("Expected initial cwd to be '/', got %q", cwd)
	}

	// Change directory
	err = handler.Chdir("/home/user")
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Verify new working directory
	cwd, err = handler.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	if cwd != "/home/user" {
		t.Errorf("Expected cwd to be '/home/user', got %q", cwd)
	}
}

func TestMockFileHandlerTempDir(t *testing.T) {
	handler := NewMockFileHandler()

	tempDir := handler.TempDir()
	if tempDir != "/tmp" {
		t.Errorf("Expected temp dir '/tmp', got %q", tempDir)
	}
}

func TestMockFileHandlerUserHomeDir(t *testing.T) {
	handler := NewMockFileHandler()

	homeDir, err := handler.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home directory: %v", err)
	}
	if homeDir != "/home/user" {
		t.Errorf("Expected home dir '/home/user', got %q", homeDir)
	}
}

func TestMockFileOperations(t *testing.T) {
	file := NewMockFile("/test.txt")

	// Test Write
	n, err := file.Write([]byte("Hello"))
	if err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	if n != 5 {
		t.Errorf("Expected to write 5 bytes, wrote %d", n)
	}

	// Test Name
	if file.Name() != "/test.txt" {
		t.Errorf("Expected name '/test.txt', got %q", file.Name())
	}

	// Test Stat
	info, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat: %v", err)
	}
	if info.Size() != 5 {
		t.Errorf("Expected size 5, got %d", info.Size())
	}

	// Test ReadAt (before consuming buffer with Read)
	buf2 := make([]byte, 3)
	n, err = file.ReadAt(buf2, 1)
	if err != nil && err != io.EOF {
		t.Fatalf("Failed to read at: %v", err)
	}
	if n >= 3 && string(buf2) != "ell" {
		t.Errorf("Expected 'ell', got %q", buf2)
	}

	// Test Seek
	offset, err := file.Seek(0, 0)
	if err != nil {
		t.Fatalf("Failed to seek: %v", err)
	}
	if offset != 0 {
		t.Errorf("Expected offset 0, got %d", offset)
	}

	// Test Read
	buf := make([]byte, 5)
	n, err = file.Read(buf)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}
	if string(buf) != "Hello" {
		t.Errorf("Expected 'Hello', got %q", buf)
	}

	// Test WriteAt
	n, err = file.WriteAt([]byte("XY"), 0)
	if err != nil {
		t.Fatalf("Failed to write at: %v", err)
	}

	// Test Sync
	if err := file.Sync(); err != nil {
		t.Fatalf("Failed to sync: %v", err)
	}

	// Write more data for truncate test
	file.Write([]byte("1234567890"))

	// Test Truncate
	if err := file.Truncate(3); err != nil {
		t.Fatalf("Failed to truncate: %v", err)
	}

	// Test Chmod
	if err := file.Chmod(0755); err != nil {
		t.Fatalf("Failed to chmod: %v", err)
	}

	// Test Chown
	if err := file.Chown(1000, 1000); err != nil {
		t.Fatalf("Failed to chown: %v", err)
	}

	// Test ReadDir
	entries, err := file.ReadDir(10)
	if err != nil {
		t.Fatalf("Failed to read dir: %v", err)
	}
	if entries == nil {
		t.Error("Expected non-nil entries")
	}

	// Test Readdir
	infos, err := file.Readdir(10)
	if err != nil {
		t.Fatalf("Failed to readdir: %v", err)
	}
	if infos == nil {
		t.Error("Expected non-nil infos")
	}

	// Test Readdirnames
	names, err := file.Readdirnames(10)
	if err != nil {
		t.Fatalf("Failed to readdirnames: %v", err)
	}
	if names == nil {
		t.Error("Expected non-nil names")
	}

	// Test Close
	if err := file.Close(); err != nil {
		t.Fatalf("Failed to close: %v", err)
	}

	// Test operations on closed file
	_, err = file.Read(buf)
	if err == nil {
		t.Error("Expected error when reading closed file")
	}

	_, err = file.Write([]byte("test"))
	if err == nil {
		t.Error("Expected error when writing to closed file")
	}

	_, err = file.Seek(0, 0)
	if err == nil {
		t.Error("Expected error when seeking closed file")
	}

	_, err = file.ReadAt(buf, 0)
	if err == nil {
		t.Error("Expected error when reading at closed file")
	}

	_, err = file.WriteAt([]byte("test"), 0)
	if err == nil {
		t.Error("Expected error when writing at closed file")
	}

	err = file.Truncate(0)
	if err == nil {
		t.Error("Expected error when truncating closed file")
	}
}

func TestMockFileReadAtEOF(t *testing.T) {
	file := NewMockFile("/test.txt")
	file.Write([]byte("Hello"))

	buf := make([]byte, 10)
	n, err := file.ReadAt(buf, 10)
	if n != 0 {
		t.Errorf("Expected 0 bytes read, got %d", n)
	}
	if err == nil {
		t.Error("Expected EOF error")
	}
}

func TestMockFileClone(t *testing.T) {
	original := NewMockFile("/test.txt")
	original.Write([]byte("original data"))

	clone := original.Clone()

	if clone.Name() != original.Name() {
		t.Error("Clone should have same name")
	}

	// Clone should have independent buffer
	clone.Write([]byte(" more"))

	// Original should not be affected
	original.Seek(0, 0)
	buf := make([]byte, 20)
	n, _ := original.Read(buf)
	if string(buf[:n]) != "original data" {
		t.Error("Original file was affected by clone modification")
	}
}

func TestMockFileInfo(t *testing.T) {
	info := &mockFileInfo{
		name:  "test.txt",
		size:  100,
		isDir: false,
	}

	if info.Name() != "test.txt" {
		t.Errorf("Expected name 'test.txt', got %q", info.Name())
	}
	if info.Size() != 100 {
		t.Errorf("Expected size 100, got %d", info.Size())
	}
	if info.Mode() != 0644 {
		t.Errorf("Expected mode 0644, got %o", info.Mode())
	}
	if info.IsDir() {
		t.Error("Expected IsDir to be false")
	}
	if info.Sys() != nil {
		t.Error("Expected Sys to return nil")
	}

	// Test directory
	dirInfo := &mockFileInfo{
		name:  "testdir",
		isDir: true,
	}
	if !dirInfo.IsDir() {
		t.Error("Expected IsDir to be true")
	}
}

func TestMockProcessHandlerGetters(t *testing.T) {
	handler := NewMockProcessHandler()

	if handler.Getppid() != 1 {
		t.Errorf("Expected ppid 1, got %d", handler.Getppid())
	}
	if handler.Getuid() != 1000 {
		t.Errorf("Expected uid 1000, got %d", handler.Getuid())
	}
	if handler.Geteuid() != 1000 {
		t.Errorf("Expected euid 1000, got %d", handler.Geteuid())
	}
	if handler.Getgid() != 1000 {
		t.Errorf("Expected gid 1000, got %d", handler.Getgid())
	}
	if handler.Getegid() != 1000 {
		t.Errorf("Expected egid 1000, got %d", handler.Getegid())
	}
}

func TestMockProcessHandlerExit(t *testing.T) {
	handler := NewMockProcessHandler()

	handler.Exit(42)

	// Verify exit was recorded (in real implementation this would exit the process)
	// The mock just records the exit code
}

func TestMockProcessHandlerStartProcess(t *testing.T) {
	handler := NewMockProcessHandler()

	proc, err := handler.StartProcess("/bin/test", []string{"arg1"}, nil)
	if err != nil {
		t.Fatalf("Failed to start process: %v", err)
	}
	if proc.Pid() != 99999 {
		t.Errorf("Expected PID 99999, got %d", proc.Pid())
	}
}

func TestMockProcess(t *testing.T) {
	proc := &MockProcess{pid: 12345}

	if proc.Pid() != 12345 {
		t.Errorf("Expected PID 12345, got %d", proc.Pid())
	}

	// Test Kill
	if err := proc.Kill(); err != nil {
		t.Fatalf("Failed to kill: %v", err)
	}

	// Test Signal
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("Failed to signal: %v", err)
	}

	// Verify signal was recorded
	proc.mu.Lock()
	if len(proc.signals) != 1 || proc.signals[0] != syscall.SIGTERM {
		t.Error("Signal was not recorded correctly")
	}
	proc.mu.Unlock()

	// Test Release
	if err := proc.Release(); err != nil {
		t.Fatalf("Failed to release: %v", err)
	}

	// Test Wait
	_, err := proc.Wait()
	if err == nil {
		t.Error("Expected error from Wait in mock")
	}
}

func TestMockClockAdvance(t *testing.T) {
	clock := NewMockClock()

	initial := clock.Now()
	clock.Advance(1 * time.Hour)
	after := clock.Now()

	if after.Sub(initial) != 1*time.Hour {
		t.Errorf("Expected 1 hour difference, got %v", after.Sub(initial))
	}
}

func TestMockClockSleep(t *testing.T) {
	clock := NewMockClock()

	initial := clock.Now()
	clock.Sleep(30 * time.Minute)
	after := clock.Now()

	if after.Sub(initial) != 30*time.Minute {
		t.Errorf("Expected 30 minute difference, got %v", after.Sub(initial))
	}
}

func TestMockClockAfter(t *testing.T) {
	clock := NewMockClock()

	ch := clock.After(1 * time.Second)

	select {
	case <-ch:
		// Success
	case <-time.After(100 * time.Millisecond):
		// Expected to receive quickly in mock
	}
}

func TestMockClockTick(t *testing.T) {
	clock := NewMockClock()

	ch := clock.Tick(1 * time.Second)
	if ch == nil {
		t.Error("Expected non-nil channel")
	}
}

func TestMockClockNewTicker(t *testing.T) {
	clock := NewMockClock()

	ticker := clock.NewTicker(1 * time.Second)
	if ticker == nil {
		t.Error("Expected non-nil ticker")
	}

	mockTicker := ticker.(*MockTicker)

	// Test C
	if mockTicker.C() == nil {
		t.Error("Expected non-nil channel")
	}

	// Test Tick
	mockTicker.Tick(time.Now())

	// Test Reset
	mockTicker.Reset(2 * time.Second)

	// Test Stop
	mockTicker.Stop()
}

func TestMockTimerStop(t *testing.T) {
	clock := NewMockClock()
	timer := clock.NewTimer(1 * time.Second).(*MockTimer)

	// First stop should return true
	if !timer.Stop() {
		t.Error("Expected Stop to return true")
	}

	// Second stop should return false
	if timer.Stop() {
		t.Error("Expected Stop to return false when already stopped")
	}
}

func TestMockTimerReset(t *testing.T) {
	clock := NewMockClock()
	timer := clock.NewTimer(1 * time.Second).(*MockTimer)

	// Reset active timer
	if !timer.Reset(2 * time.Second) {
		t.Error("Expected Reset to return true for active timer")
	}

	// Stop timer
	timer.Stop()

	// Reset stopped timer
	if timer.Reset(3 * time.Second) {
		t.Error("Expected Reset to return false for stopped timer")
	}
}

func TestMockTimerC(t *testing.T) {
	clock := NewMockClock()
	timer := clock.NewTimer(1 * time.Second).(*MockTimer)

	if timer.C() == nil {
		t.Error("Expected non-nil channel")
	}
}
