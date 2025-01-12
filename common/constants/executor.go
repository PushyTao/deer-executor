package constants

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// Judge Flags
const (
	// JudgeFlagAC Accepted
	JudgeFlagAC = 0
	// JudgeFlagPE Presentation Error
	JudgeFlagPE = 1
	// JudgeFlagTLE Time Limit Exceeded
	JudgeFlagTLE = 2
	// JudgeFlagMLE Memory Limit Exceeded
	JudgeFlagMLE = 3
	// JudgeFlagWA Wrong Answer
	JudgeFlagWA = 4
	// JudgeFlagRE Runtime Error
	JudgeFlagRE = 5
	// JudgeFlagOLE Output Limit Exceeded
	JudgeFlagOLE = 6
	// JudgeFlagCE Compile Error
	JudgeFlagCE = 7
	// JudgeFlagSE System Error
	JudgeFlagSE = 8

	// 9 is for ReJudge

	// Special Judge Checker Time OUT
	JudgeFlagSpecialJudgeTimeout = 10
	// Special Judge Checker ERROR
	JudgeFlagSpecialJudgeError = 11
	// Special Judge Checker Finish, Need Standard Checkup
	JudgeFlagSpecialJudgeRequireChecker = 12
)

// Special Judge Mode
const (
	SpecialJudgeModeDisabled    = 0
	SpecialJudgeModeChecker     = 1
	SpecialJudgeModeInteractive = 2

	// unit: ms
	SpecialJudgeTimeLimit = 1 * 1000
	// unit: kb
	SpecialJudgeMemoryLimit = 256 * 1024
)

// SignalNumberMap  map unix signal to text
var SignalNumberMap = map[int][]string{
	1: {"SIGHUP", "Hangup (POSIX)."},
	2: {"SIGINT", "Interrupt (ANSI)."},
	3: {"SIGQUIT", "Quit (POSIX)."},
	4: {"SIGILL", "Illegal instruction (ANSI)."},
	5: {"SIGTRAP", "Trace trap (POSIX)."},
	6: {"SIGABRT", "Abort (ANSI)."},
	//6:  []string{"SIGIOT", "IOT trap (4.2 BSD)."},
	7:  {"SIGBUS", "BUS error (4.2 BSD)."},
	8:  {"SIGFPE", "Floating-point exception (ANSI)."},
	9:  {"SIGKILL", "Kill, unblockable (POSIX)."},
	10: {"SIGUSR1", "User-defined signal 1 (POSIX)."},
	11: {"SIGSEGV", "Segmentation violation (ANSI)."},
	12: {"SIGUSR2", "User-defined signal 2 (POSIX)."},
	13: {"SIGPIPE", "Broken pipe (POSIX)."},
	14: {"SIGALRM", "Alarm clock (POSIX)."},
	15: {"SIGTERM", "Termination (ANSI)."},
	16: {"SIGSTKFLT", "Stack fault."},
	17: {"SIGCHLD", "Child status has changed (POSIX)."},
	18: {"SIGCONT", "Continue (POSIX)."},
	19: {"SIGSTOP", "Stop, unblockable (POSIX)."},
	20: {"SIGTSTP", "Keyboard stop (POSIX)."},
	21: {"SIGTTIN", "Background read from tty (POSIX)."},
	22: {"SIGTTOU", "Background write to tty (POSIX)."},
	23: {"SIGURG", "Urgent condition on socket (4.2 BSD)."},
	24: {"SIGXCPU", "CPU limit exceeded (4.2 BSD)."},
	25: {"SIGXFSZ", "File size limit exceeded (4.2 BSD)."},
	26: {"SIGVTALRM", "Virtual alarm clock (4.2 BSD)."},
	27: {"SIGPROF", "Profiling alarm clock (4.2 BSD)."},
	28: {"SIGWINCH", "Window size change (4.3 BSD, Sun)."},
	29: {"SIGIO", "I/O now possible (4.2 BSD)."},
	30: {"SIGPWR", "Power failure restart (System V)."},
	31: {"SIGSYS", "Bad system call."},
}

// FlagMeansMap map judge flags to text
var FlagMeansMap = map[int]string{
	0:  "Accepted",
	1:  "Presentation Error",
	2:  "Time Limit Exceeded",
	3:  "Memory Limit Exceeded",
	4:  "Wrong Answer",
	5:  "Runtime Error",
	6:  "Output Limit Exceeded",
	7:  "Compile Error",
	8:  "System Error",
	9:  "Special Judge Checker Time OUT",
	10: "Special Judge Checker ERROR",
	11: "Special Judge Checker Finish, Need Standard Checkup",
}

// MemorySizeForJIT 给动态语言、带虚拟机的语言设定虚拟机自身的初始内存大小
var MemorySizeForJIT = map[string]int{
	"gcc":     0,
	"g++":     0,
	"java":    393216, // java
	"python2": 65536,  // py2
	"python3": 65536,  // py3
	"nodejs":  262144, // js
	"golang":  0,
	"php":     131072, // php
	"ruby":    65536,  // ruby
	"rust":    0,
}

// PlaceMemorySizeForJIT 替换JIT虚拟内存表
func PlaceMemorySizeForJIT(configFile string) error {
	if configFile != "" {
		_, err := os.Stat(configFile)
		// ignore
		if os.IsNotExist(err) {
			return nil
		}
		cbody, err := ioutil.ReadFile(configFile)
		if err != nil {
			return err
		}
		err = json.Unmarshal(cbody, &MemorySizeForJIT)
		if err != nil {
			return err
		}
	}
	return nil
}
