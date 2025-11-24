package platform_test

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/grinps/go-utils/base-utils/platform"
)

// Example: Testing signal handling with mock platform
func TestSignalHandlingWithMock(t *testing.T) {
	mockPlatform := platform.NewMockPlatform()
	signalHandler := mockPlatform.Signal().(*platform.MockSignalHandler)

	// Create a channel to receive signals
	sigChan := make(chan os.Signal, 1)
	signalHandler.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Simulate sending a signal
	signalHandler.SendSignal(os.Interrupt)

	// Verify signal was received
	select {
	case sig := <-sigChan:
		if sig != os.Interrupt {
			t.Errorf("expected os.Interrupt, got %v", sig)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for signal")
	}
}

// Example: Testing environment variables with mock platform
func TestEnvironmentVariablesWithMock(t *testing.T) {
	mockPlatform := platform.NewMockPlatform()
	env := mockPlatform.Env()

	// Set environment variables
	if err := env.Setenv("TEST_VAR", "test_value"); err != nil {
		t.Fatalf("failed to set env var: %v", err)
	}

	// Get environment variable
	value := env.Getenv("TEST_VAR")
	if value != "test_value" {
		t.Errorf("expected 'test_value', got '%s'", value)
	}

	// Lookup environment variable
	value, ok := env.LookupEnv("TEST_VAR")
	if !ok {
		t.Error("expected TEST_VAR to exist")
	}
	if value != "test_value" {
		t.Errorf("expected 'test_value', got '%s'", value)
	}

	// Test non-existent variable
	_, ok = env.LookupEnv("NON_EXISTENT")
	if ok {
		t.Error("expected NON_EXISTENT to not exist")
	}

	// Unset environment variable
	if err := env.Unsetenv("TEST_VAR"); err != nil {
		t.Fatalf("failed to unset env var: %v", err)
	}

	value = env.Getenv("TEST_VAR")
	if value != "" {
		t.Errorf("expected empty string after unset, got '%s'", value)
	}
}

// Example: Testing file operations with mock platform
func TestFileOperationsWithMock(t *testing.T) {
	mockPlatform := platform.NewMockPlatform()
	fileHandler := mockPlatform.File()

	// Write a file
	testData := []byte("Hello, World!")
	if err := fileHandler.WriteFile("/test/file.txt", testData, 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Read the file
	data, err := fileHandler.ReadFile("/test/file.txt")
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(data) != string(testData) {
		t.Errorf("expected '%s', got '%s'", testData, data)
	}

	// Create a file
	file, err := fileHandler.Create("/test/newfile.txt")
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	defer file.Close()

	// Write to the file
	n, err := file.Write([]byte("test content"))
	if err != nil {
		t.Fatalf("failed to write to file: %v", err)
	}
	if n != 12 {
		t.Errorf("expected to write 12 bytes, wrote %d", n)
	}

	// Stat the file
	info, err := fileHandler.Stat("/test/file.txt")
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}
	if info.Name() != "/test/file.txt" {
		t.Errorf("expected name '/test/file.txt', got '%s'", info.Name())
	}

	// Remove the file
	if err := fileHandler.Remove("/test/file.txt"); err != nil {
		t.Fatalf("failed to remove file: %v", err)
	}

	// Verify file is removed
	_, err = fileHandler.Stat("/test/file.txt")
	if err == nil {
		t.Error("expected error when statting removed file")
	}
}

// Example: Testing process operations with mock platform
func TestProcessOperationsWithMock(t *testing.T) {
	mockPlatform := platform.NewMockPlatform()
	processHandler := mockPlatform.Process().(*platform.MockProcessHandler)

	// Set custom PID for testing
	processHandler.SetPid(54321)

	// Get PID
	pid := processHandler.Getpid()
	if pid != 54321 {
		t.Errorf("expected PID 54321, got %d", pid)
	}

	// Set custom hostname
	processHandler.SetHostname("test-host")

	// Get hostname
	hostname, err := processHandler.Hostname()
	if err != nil {
		t.Fatalf("failed to get hostname: %v", err)
	}
	if hostname != "test-host" {
		t.Errorf("expected hostname 'test-host', got '%s'", hostname)
	}

	// Find process
	proc, err := processHandler.FindProcess(12345)
	if err != nil {
		t.Fatalf("failed to find process: %v", err)
	}
	if proc.Pid() != 12345 {
		t.Errorf("expected PID 12345, got %d", proc.Pid())
	}
}

// Example: Testing time operations with mock clock
func TestClockOperationsWithMock(t *testing.T) {
	mockPlatform := platform.NewMockPlatform()
	clock := mockPlatform.Clock().(*platform.MockClock)

	// Set initial time
	initialTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	clock.Set(initialTime)

	// Get current time
	now := clock.Now()
	if !now.Equal(initialTime) {
		t.Errorf("expected time %v, got %v", initialTime, now)
	}

	// Advance time
	clock.Advance(1 * time.Hour)
	now = clock.Now()
	expectedTime := initialTime.Add(1 * time.Hour)
	if !now.Equal(expectedTime) {
		t.Errorf("expected time %v, got %v", expectedTime, now)
	}

	// Test timer
	timer := clock.NewTimer(5 * time.Minute).(*platform.MockTimer)

	// Manually fire the timer
	timer.Fire()

	select {
	case <-timer.C():
		// Timer fired successfully
	case <-time.After(100 * time.Millisecond):
		t.Error("timer did not fire")
	}
}

// Example: Real-world usage pattern with dependency injection
type Service struct {
	platform platform.Platform
}

func NewService(p platform.Platform) *Service {
	return &Service{platform: p}
}

func (s *Service) GetConfigValue(key string) string {
	return s.platform.Env().Getenv(key)
}

func (s *Service) SaveData(filename string, data []byte) error {
	return s.platform.File().WriteFile(filename, data, 0644)
}

func (s *Service) GetCurrentTime() time.Time {
	return s.platform.Clock().Now()
}

func TestServiceWithMockPlatform(t *testing.T) {
	mockPlatform := platform.NewMockPlatform()
	service := NewService(mockPlatform)

	// Setup test environment
	mockPlatform.Env().Setenv("CONFIG_KEY", "config_value")

	// Test getting config value
	value := service.GetConfigValue("CONFIG_KEY")
	if value != "config_value" {
		t.Errorf("expected 'config_value', got '%s'", value)
	}

	// Test saving data
	testData := []byte("test data")
	if err := service.SaveData("/test/data.txt", testData); err != nil {
		t.Fatalf("failed to save data: %v", err)
	}

	// Verify data was saved
	savedData, err := mockPlatform.File().ReadFile("/test/data.txt")
	if err != nil {
		t.Fatalf("failed to read saved data: %v", err)
	}
	if string(savedData) != string(testData) {
		t.Errorf("expected '%s', got '%s'", testData, savedData)
	}

	// Test getting current time
	mockClock := mockPlatform.Clock().(*platform.MockClock)
	testTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	mockClock.Set(testTime)

	currentTime := service.GetCurrentTime()
	if !currentTime.Equal(testTime) {
		t.Errorf("expected time %v, got %v", testTime, currentTime)
	}
}

// Example: Testing with real OS platform
func TestServiceWithRealPlatform(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping real platform test in short mode")
	}

	realPlatform := platform.NewOSPlatform()
	service := NewService(realPlatform)

	// This would use actual OS operations
	_ = service.GetCurrentTime()

	// Note: Be careful with real OS operations in tests
	// They can have side effects and may not be deterministic
}
