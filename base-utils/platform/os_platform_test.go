package platform

import (
	"os"
	"testing"
	"time"
)

func TestOSPlatformAccessors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping OS platform test in short mode")
	}

	osPlatform := NewOSPlatform()

	if osPlatform.Signal() == nil {
		t.Error("Signal() returned nil")
	}
	if osPlatform.Env() == nil {
		t.Error("Env() returned nil")
	}
	if osPlatform.File() == nil {
		t.Error("File() returned nil")
	}
	if osPlatform.Process() == nil {
		t.Error("Process() returned nil")
	}
	if osPlatform.Clock() == nil {
		t.Error("Clock() returned nil")
	}
}

func TestOSSignalHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping OS signal test in short mode")
	}

	handler := &osSignalHandler{}

	// Test Notify and Stop
	sigChan := make(chan os.Signal, 1)
	handler.Notify(sigChan, os.Interrupt)
	handler.Stop(sigChan)

	// Test Ignore and Reset
	handler.Ignore(os.Interrupt)
	handler.Reset(os.Interrupt)
}

func TestOSEnvHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping OS env test in short mode")
	}

	handler := &osEnvHandler{}

	// Test Setenv and Getenv
	testKey := "TEST_PLATFORM_KEY_12345"
	testValue := "test_value"

	if err := handler.Setenv(testKey, testValue); err != nil {
		t.Fatalf("Failed to set env: %v", err)
	}

	value := handler.Getenv(testKey)
	if value != testValue {
		t.Errorf("Expected %q, got %q", testValue, value)
	}

	// Test LookupEnv
	value, ok := handler.LookupEnv(testKey)
	if !ok {
		t.Error("Expected key to exist")
	}
	if value != testValue {
		t.Errorf("Expected %q, got %q", testValue, value)
	}

	// Test Environ
	environ := handler.Environ()
	if len(environ) == 0 {
		t.Error("Expected non-empty environ")
	}

	// Test ExpandEnv
	handler.Setenv("TEST_USER", "testuser")
	expanded := handler.ExpandEnv("User: $TEST_USER")
	if expanded != "User: testuser" {
		t.Errorf("Expected 'User: testuser', got %q", expanded)
	}

	// Test Unsetenv
	if err := handler.Unsetenv(testKey); err != nil {
		t.Fatalf("Failed to unset env: %v", err)
	}

	value = handler.Getenv(testKey)
	if value != "" {
		t.Errorf("Expected empty string after unset, got %q", value)
	}

	// Cleanup
	handler.Unsetenv("TEST_USER")
}

func TestOSFileHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping OS file test in short mode")
	}

	handler := &osFileHandler{}

	// Create temp directory for testing
	tempDir := os.TempDir() + "/platform_test_" + time.Now().Format("20060102150405")
	defer os.RemoveAll(tempDir)

	// Test MkdirAll
	if err := handler.MkdirAll(tempDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Test WriteFile
	testFile := tempDir + "/test.txt"
	testData := []byte("Hello, World!")
	if err := handler.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Test Stat
	info, err := handler.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if info.Size() != int64(len(testData)) {
		t.Errorf("Expected size %d, got %d", len(testData), info.Size())
	}

	// Test Lstat
	info, err = handler.Lstat(testFile)
	if err != nil {
		t.Fatalf("Failed to lstat file: %v", err)
	}

	// Test ReadFile
	data, err := handler.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(data) != string(testData) {
		t.Errorf("Expected %q, got %q", testData, data)
	}

	// Test Create
	newFile := tempDir + "/new.txt"
	file, err := handler.Create(newFile)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	file.Close()

	// Test Open
	file, err = handler.Open(newFile)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	file.Close()

	// Test OpenFile
	file, err = handler.OpenFile(newFile, os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	file.Close()

	// Test Rename
	renamedFile := tempDir + "/renamed.txt"
	if err := handler.Rename(testFile, renamedFile); err != nil {
		t.Fatalf("Failed to rename file: %v", err)
	}

	// Test Mkdir
	subDir := tempDir + "/subdir"
	if err := handler.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Test ReadDir
	entries, err := handler.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}
	if len(entries) == 0 {
		t.Error("Expected non-empty directory")
	}

	// Test Getwd
	cwd, err := handler.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	if cwd == "" {
		t.Error("Expected non-empty working directory")
	}

	// Test Chdir
	originalCwd := cwd
	if err := handler.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	// Restore original directory
	handler.Chdir(originalCwd)

	// Test TempDir
	tmpDir := handler.TempDir()
	if tmpDir == "" {
		t.Error("Expected non-empty temp directory")
	}

	// Test UserHomeDir
	homeDir, err := handler.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home directory: %v", err)
	}
	if homeDir == "" {
		t.Error("Expected non-empty home directory")
	}

	// Test Remove
	if err := handler.Remove(newFile); err != nil {
		t.Fatalf("Failed to remove file: %v", err)
	}

	// Test RemoveAll
	if err := handler.RemoveAll(tempDir); err != nil {
		t.Fatalf("Failed to remove all: %v", err)
	}
}

func TestOSProcessHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping OS process test in short mode")
	}

	handler := &osProcessHandler{}

	// Test Getpid
	pid := handler.Getpid()
	if pid <= 0 {
		t.Errorf("Expected positive PID, got %d", pid)
	}

	// Test Getppid
	ppid := handler.Getppid()
	if ppid <= 0 {
		t.Errorf("Expected positive PPID, got %d", ppid)
	}

	// Test Getuid
	uid := handler.Getuid()
	if uid < 0 {
		t.Errorf("Expected non-negative UID, got %d", uid)
	}

	// Test Geteuid
	euid := handler.Geteuid()
	if euid < 0 {
		t.Errorf("Expected non-negative EUID, got %d", euid)
	}

	// Test Getgid
	gid := handler.Getgid()
	if gid < 0 {
		t.Errorf("Expected non-negative GID, got %d", gid)
	}

	// Test Getegid
	egid := handler.Getegid()
	if egid < 0 {
		t.Errorf("Expected non-negative EGID, got %d", egid)
	}

	// Test Hostname
	hostname, err := handler.Hostname()
	if err != nil {
		t.Fatalf("Failed to get hostname: %v", err)
	}
	if hostname == "" {
		t.Error("Expected non-empty hostname")
	}

	// Test FindProcess
	proc, err := handler.FindProcess(pid)
	if err != nil {
		t.Fatalf("Failed to find process: %v", err)
	}
	if proc.Pid() != pid {
		t.Errorf("Expected PID %d, got %d", pid, proc.Pid())
	}

	// Test osProcess methods
	osProc := proc.(*osProcess)

	// Test Pid
	if osProc.Pid() != pid {
		t.Errorf("Expected PID %d, got %d", pid, osProc.Pid())
	}

	// Note: We don't test Kill, Signal, Wait, or Release on the current process
	// as they would have side effects
}

func TestOSClock(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping OS clock test in short mode")
	}

	clock := &osClock{}

	// Test Now
	now := clock.Now()
	if now.IsZero() {
		t.Error("Expected non-zero time")
	}

	// Test Sleep (very short duration)
	start := time.Now()
	clock.Sleep(10 * time.Millisecond)
	elapsed := time.Since(start)
	if elapsed < 10*time.Millisecond {
		t.Errorf("Sleep didn't wait long enough: %v", elapsed)
	}

	// Test After
	afterChan := clock.After(10 * time.Millisecond)
	select {
	case <-afterChan:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Error("After channel didn't fire in time")
	}

	// Test Tick
	tickChan := clock.Tick(10 * time.Millisecond)
	if tickChan == nil {
		t.Error("Expected non-nil tick channel")
	}

	// Test NewTimer
	timer := clock.NewTimer(10 * time.Millisecond)
	if timer == nil {
		t.Error("Expected non-nil timer")
	}

	osTimer := timer.(*osTimer)

	// Test timer C
	if osTimer.C() == nil {
		t.Error("Expected non-nil timer channel")
	}

	// Wait for timer to fire
	<-osTimer.C()

	// Test timer Reset (may return false if timer already fired)
	osTimer.Reset(10 * time.Millisecond)

	// Test timer Stop
	if !osTimer.Stop() {
		t.Error("Expected Stop to return true")
	}

	// Test NewTicker
	ticker := clock.NewTicker(10 * time.Millisecond)
	if ticker == nil {
		t.Error("Expected non-nil ticker")
	}

	osTicker := ticker.(*osTicker)

	// Test ticker C
	if osTicker.C() == nil {
		t.Error("Expected non-nil ticker channel")
	}

	// Wait for at least one tick
	select {
	case <-osTicker.C():
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Error("Ticker didn't tick in time")
	}

	// Test ticker Reset
	osTicker.Reset(20 * time.Millisecond)

	// Test ticker Stop
	osTicker.Stop()
}

func TestOSProcessStartProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping OS start process test in short mode")
	}

	handler := &osProcessHandler{}

	// Start a simple process (echo command)
	proc, err := handler.StartProcess("/bin/echo", []string{"echo", "test"}, &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	})
	if err != nil {
		t.Fatalf("Failed to start process: %v", err)
	}

	if proc.Pid() <= 0 {
		t.Errorf("Expected positive PID, got %d", proc.Pid())
	}

	// Wait for process to complete
	state, err := proc.Wait()
	if err != nil {
		t.Fatalf("Failed to wait for process: %v", err)
	}
	if state == nil {
		t.Error("Expected non-nil process state")
	}

	// Test Release
	if err := proc.Release(); err != nil {
		t.Fatalf("Failed to release process: %v", err)
	}
}
