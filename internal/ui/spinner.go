package ui

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// Spinner represents a loading spinner
type Spinner struct {
	frames   []string
	delay    time.Duration
	message  string
	active   bool
	stopChan chan struct{}
	mutex    sync.Mutex
}

// Predefined spinner styles
var (
	SpinnerDots  = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	SpinnerLine  = []string{"|", "/", "-", "\\"}
	SpinnerFire  = []string{"🔥", "🔶", "🔸", "🔹", "🔷", "🔵"}
	SpinnerBox   = []string{"▖", "▘", "▝", "▗"}
	SpinnerArrow = []string{"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"}
)

// NewSpinner creates a new spinner with the given style and message
func NewSpinner(frames []string, message string) *Spinner {
	return &Spinner{
		frames:   frames,
		delay:    100 * time.Millisecond,
		message:  message,
		stopChan: make(chan struct{}),
	}
}

// NewDotsSpinner creates a dots spinner
func NewDotsSpinner(message string) *Spinner {
	return NewSpinner(SpinnerDots, message)
}

// NewFireSpinner creates a fire emoji spinner
func NewFireSpinner(message string) *Spinner {
	return NewSpinner(SpinnerFire, message)
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	s.mutex.Lock()
	if s.active {
		s.mutex.Unlock()
		return
	}
	s.active = true
	s.mutex.Unlock()

	go func() {
		for {
			for _, frame := range s.frames {
				select {
				case <-s.stopChan:
					return
				default:
					s.mutex.Lock()
					if !s.active {
						s.mutex.Unlock()
						return
					}
					// Clear line and print spinner
					fmt.Print("\r\033[K")
					fmt.Printf("%s %s", Primary.Sprint(frame), s.message)
					s.mutex.Unlock()
					time.Sleep(s.delay)
				}
			}
		}
	}()
}

// Stop ends the spinner animation
func (s *Spinner) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.active {
		return
	}

	s.active = false
	close(s.stopChan)

	// Clear the spinner line
	fmt.Print("\r\033[K")
}

// Update changes the spinner message
func (s *Spinner) Update(message string) {
	s.mutex.Lock()
	s.message = message
	s.mutex.Unlock()
}

// SetDelay sets the animation delay
func (s *Spinner) SetDelay(delay time.Duration) {
	s.mutex.Lock()
	s.delay = delay
	s.mutex.Unlock()
}

// Progress represents a progress bar
type Progress struct {
	total   int
	current int
	width   int
	message string
}

// NewProgress creates a new progress bar
func NewProgress(total int, message string) *Progress {
	return &Progress{
		total:   total,
		current: 0,
		width:   50,
		message: message,
	}
}

// Update updates the progress bar
func (p *Progress) Update(current int, message string) {
	p.current = current
	if message != "" {
		p.message = message
	}
	p.render()
}

// Increment increments the progress by 1
func (p *Progress) Increment(message string) {
	p.Update(p.current+1, message)
}

// Complete marks the progress as complete
func (p *Progress) Complete(message string) {
	p.Update(p.total, message)
	fmt.Println() // New line after completion
}

// render draws the progress bar
func (p *Progress) render() {
	percentage := float64(p.current) / float64(p.total)
	filled := int(percentage * float64(p.width))

	bar := ""
	for i := 0; i < p.width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	fmt.Printf("\r%s [%s] %d/%d %s",
		Primary.Sprint("Progress:"),
		Success.Sprint(bar),
		p.current,
		p.total,
		p.message)
}

// Typewriter effect for dramatic text display
func Typewriter(text string, delay time.Duration) {
	for _, char := range text {
		fmt.Print(string(char))
		time.Sleep(delay)
	}
	fmt.Println()
}

// AnimatedHeader displays an animated header
func AnimatedHeader(text string) {
	// Clear screen effect
	fmt.Print("\033[H\033[2J")

	// Animated fire emoji
	fireFrames := []string{"🔥", "🔶", "🔸", "💥", "🔥"}
	for _, frame := range fireFrames {
		fmt.Printf("\r%s %s", frame, Primary.Sprint(text))
		time.Sleep(200 * time.Millisecond)
	}
	fmt.Println()
}

// LoadingDots shows animated loading dots
func LoadingDots(message string, duration time.Duration) {
	start := time.Now()
	dots := ""

	for time.Since(start) < duration {
		fmt.Printf("\r%s%s", Primary.Sprint(message), Secondary.Sprint(dots))
		dots += "."
		if len(dots) > 3 {
			dots = ""
		}
		time.Sleep(300 * time.Millisecond)
	}
	fmt.Print("\r\033[K") // Clear line
}

// ShowLoader displays a loader with callback
func ShowLoader(message string, callback func() error) error {
	spinner := NewDotsSpinner(message)
	spinner.Start()

	err := callback()

	spinner.Stop()

	if err != nil {
		ErrorMsg(fmt.Sprintf("Failed: %v", err))
	} else {
		SuccessMsg("Complete!")
	}

	return err
}

// Interactive prompt with animation
func PromptWithSpinner(message string, options []string) string {
	// Show spinner while "thinking"
	spinner := NewDotsSpinner("Processing options...")
	spinner.Start()
	time.Sleep(500 * time.Millisecond) // Simulate processing
	spinner.Stop()

	fmt.Printf("🤔 %s\n", Primary.Sprint(message))
	for i, option := range options {
		fmt.Printf("  %s %s\n", Secondary.Sprintf("%d.", i+1), option)
	}

	var response string
	fmt.Print("\n➤ Your choice: ")
	fmt.Scanln(&response)

	return response
}

// CheckIfTerminalSupportsColor checks if terminal supports color output
func CheckIfTerminalSupportsColor() bool {
	return os.Getenv("TERM") != "dumb" &&
		(os.Getenv("COLORTERM") != "" || os.Getenv("TERM_PROGRAM") != "")
}

// Animate success completion
func AnimatedSuccess(message string) {
	// Build up checkmark animation
	checkFrames := []string{"○", "◐", "◓", "◑", "●", "✓", "✅"}

	for _, frame := range checkFrames {
		fmt.Printf("\r%s %s", Success.Sprint(frame), message)
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Println()
}

// Animate error
func AnimatedError(message string) {
	// Build up X animation
	errorFrames := []string{"○", "◐", "◓", "◑", "●", "✗", "❌"}

	for _, frame := range errorFrames {
		fmt.Printf("\r%s %s", Error.Sprint(frame), message)
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Println()
}
