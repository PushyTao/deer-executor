// +build linux darwin

package executor

import (
    "context"
    "github.com/LanceLRQ/deer-common/constants"
    "github.com/LanceLRQ/deer-common/sandbox/forkexec"
    "github.com/LanceLRQ/deer-common/sandbox/process"
    commonStructs "github.com/LanceLRQ/deer-common/structs"
    "github.com/pkg/errors"
    "log"
    "os"
    "os/exec"
    "path"
    "path/filepath"
    "syscall"
    "time"
)

// 额外需要被注入的环境变量
var ExtraEnviron = []string{"PYTHONIOENCODING=utf-8"}


// 运行目标程序
func (session *JudgeSession) runNormalJudge(rst *commonStructs.TestCaseResult) (*ProcessInfo, error) {
    ctx, cancel := context.WithTimeout(context.Background(), time.Duration(session.Timeout) * time.Second)
    defer cancel()
    return session.runAsync(rst, false, ctx)
}

// 运行特殊评测
func (session *JudgeSession) runSpecialJudge(rst *commonStructs.TestCaseResult) (*ProcessInfo, *ProcessInfo, error) {
    if session.JudgeConfig.SpecialJudge.Mode == constants.SpecialJudgeModeChecker {
        // checker模式，用runAsync依次运行
        ctx1, cancel1 := context.WithTimeout(context.Background(), time.Duration(session.Timeout) * time.Second)
        defer cancel1()
        answer, err := session.runAsync(rst, false, ctx1)
        if err != nil {
            return nil, nil, err
        }
        ctx2, cancel2 := context.WithTimeout(context.Background(), time.Duration(session.Timeout) * time.Second)
        defer cancel2()
        checker, err := session.runAsync(rst, true, ctx2)
        if err != nil {
            return nil, nil, err
        }
        return answer, checker, nil
    }  else if session.JudgeConfig.SpecialJudge.Mode == constants.SpecialJudgeModeInteractive {
        // 交互模式
        ctx, cancel := context.WithTimeout(context.Background(), time.Duration(session.Timeout) * time.Second)
        defer cancel()
        return session.runInteractiveAsync(rst, ctx)
    }
    return nil, nil, errors.Errorf("unkonw special judge mode")
}

// 运行目标程序
func (session *JudgeSession) runAsync(rst *commonStructs.TestCaseResult, isChecker bool, ctx context.Context) (*ProcessInfo, error) {
    runSuccess := make(chan bool, 1)
    pid := 0
    var pinfo *ProcessInfo
    var err error
    go func() {
        var pstate *process.ProcessState
        // create process
        pinfo, err = startProcess(session, rst, isChecker, false, nil)
        if err != nil {
            runSuccess <- false
            return
        }
        pid = pinfo.Pid
        log.Printf("Start process (%d)...\n", pinfo.Pid)
        // Wait for exit.
        pstate, err = pinfo.Process.Wait()
        if err != nil {
            runSuccess <- false
            return
        }
        log.Printf("Process (%d) exited.\n", pinfo.Pid)
        pinfo.Status = pstate.Sys().(syscall.WaitStatus)
        pinfo.Rusage = pstate.SysUsage().(syscall.Rusage)

        runSuccess <- true
    }()

    select {
        case ok := <- runSuccess:
            if ok {
                return pinfo, nil
            } else {
                return nil, err
            }
        case <- ctx.Done(): // 触发超时
            if pid > 0 {
                _ = syscall.Kill(pid, syscall.SIGKILL)
            }
            log.Println("Child process timeout!")
            return nil, errors.Errorf("Child process timeout!")
    }
}

// 运行交互评测
func (session *JudgeSession) runInteractiveAsync(rst *commonStructs.TestCaseResult, ctx context.Context) (*ProcessInfo, *ProcessInfo, error) {
    var answer, checker *ProcessInfo
    var answerErr, checkerErr, gErr error

    fdChecker, err := forkexec.GetPipe()
    if err != nil {
        return nil, nil, errors.Errorf("create pipe error: %s", err.Error())
    }

    fdAnswer, err := forkexec.GetPipe()
    if err != nil {
        return nil, nil, errors.Errorf("create pipe error: %s", err.Error())
    }

    answerSuccess := make(chan bool, 1)
    checkerSuccess := make(chan bool, 1)
    answerPid := 0
    checkerPid := 0
    exitCounter := 0

    go func() {
        var pstate *process.ProcessState
        // create process
        answer, answerErr = startProcess(session, rst, false, true, []uintptr{ fdAnswer[0], fdChecker[1] })
        if answerErr != nil {
            answerSuccess <- false
            return
        }
        answerPid = answer.Pid
        log.Printf("[Interactive]Start answer process (%d)...\n", answer.Pid)
        // Wait for exit.
        pstate, answerErr = answer.Process.Wait()
        if answerErr != nil {
            answerSuccess <- false
            return
        }
        log.Printf("Process (%d) exited.\n", answer.Pid)
        answer.Status = pstate.Sys().(syscall.WaitStatus)
        answer.Rusage = pstate.SysUsage().(syscall.Rusage)

        answerSuccess <- true
    }()

    go func() {
        var pstate *process.ProcessState
        // create process
        checker, checkerErr = startProcess(session, rst, true, true, []uintptr{ fdChecker[0], fdAnswer[1] })
        if checkerErr != nil {
            checkerSuccess <- false
            return
        }
        checkerPid = checker.Pid
        log.Printf("[Interactive]Start checker process (%d)...\n", checker.Pid)
        // Wait for exit.
        pstate, checkerErr = checker.Process.Wait()
        if checkerErr != nil {
            checkerSuccess <- false
            return
        }
        log.Printf("Process (%d) exited.\n", checker.Pid)
        checker.Status = pstate.Sys().(syscall.WaitStatus)
        checker.Rusage = pstate.SysUsage().(syscall.Rusage)

        checkerSuccess <- true
    }()

    select {
        case ok := <- answerSuccess:
            if ok {
                exitCounter++
                if exitCounter >= 2 {
                   goto finish
                }
            } else {
                gErr = answerErr
                goto doClean
            }
        case ok := <- checkerSuccess:
            if ok {
                exitCounter++
                if exitCounter >= 2 {
                   goto finish
                }
            } else {
                gErr = checkerErr
                goto doClean
            }
        case <- ctx.Done(): // 触发超时
            log.Println("Child process timeout!")
            gErr = errors.Errorf("Child process timeout!")
            goto doClean
    }

doClean:
    if answerPid > 0 {
        _ = syscall.Kill(answerPid, syscall.SIGKILL)
    }
    if checkerPid > 0 {
        _ = syscall.Kill(checkerPid, syscall.SIGKILL)
    }
finish:
    if gErr != nil {
        return nil, nil, gErr
    } else {
        return answer, checker, nil
    }
}


// 运行一个新的进程
func startProcess(session *JudgeSession, rst *commonStructs.TestCaseResult, isChecker, pipeMode bool, pipeFd []uintptr) (*ProcessInfo, error) {
    var err error
    // Get shell commands
    commands := session.Commands
    // 参考exec.Command，从环境变量获取编译器/VM真实的地址
    programPath := commands[0]
    if filepath.Base(programPath) == programPath {
        if programPath, err = exec.LookPath(programPath); err != nil {
            return nil, err
        }
    }
    var infile, outfile, errfile string
    var rlimit forkexec.ExecRLimit
    var args []string
    var files []interface{}
    if isChecker {
        // 如果不使用TestLib，可以开启把程序的Answer发送到Checker的Stdin，兼容以前的判题程序用。
        if !session.JudgeConfig.SpecialJudge.UseTestlib {
            if session.JudgeConfig.SpecialJudge.RedirectProgramOut {
                infile = path.Join(session.SessionDir, rst.ProgramOut)
            }
        }

        outfile = path.Join(session.SessionDir, rst.CheckerOut)
        errfile = path.Join(session.SessionDir, rst.CheckerError)
        rlimit = forkexec.ExecRLimit{
            TimeLimit: session.JudgeConfig.SpecialJudge.TimeLimit,
            MemoryLimit: session.JudgeConfig.SpecialJudge.MemoryLimit,
            StackLimit: session.JudgeConfig.SpecialJudge.MemoryLimit,
            RealTimeLimit: session.JudgeConfig.RealTimeLimit,
            FileSizeLimit: session.JudgeConfig.FileSizeLimit,
        }
        args = getSpecialJudgeArgs(session, rst)
    } else {
        infile = path.Join(session.ConfigDir, rst.Input)
        outfile = path.Join(session.SessionDir, rst.ProgramOut)
        errfile = path.Join(session.SessionDir, rst.ProgramError)
        rlimit = forkexec.ExecRLimit{
            TimeLimit: session.JudgeConfig.TimeLimit,
            MemoryLimit: session.JudgeConfig.MemoryLimit,
            StackLimit: session.JudgeConfig.MemoryLimit,
            RealTimeLimit: session.JudgeConfig.RealTimeLimit,
            FileSizeLimit: session.JudgeConfig.FileSizeLimit,
        }
        args = commands
    }

    if pipeMode {
        // Open err file
        stderr, err := os.OpenFile(errfile, os.O_RDWR|os.O_CREATE, 0644)
        if err != nil {
            return nil, err
        }
        files = []interface{}{ pipeFd[0], pipeFd[1], stderr }
    } else {
        // Open in file
        stdin, err := os.OpenFile(infile, os.O_WRONLY, 0)
        if err != nil {
            return nil, err
        }
        // Open out file
        stdout, err := os.OpenFile(outfile, os.O_RDWR|os.O_CREATE, 0644)
        if err != nil {
            return nil, err
        }
        // Open err file
        stderr, err := os.OpenFile(errfile, os.O_RDWR|os.O_CREATE, 0644)
        if err != nil {
            return nil, err
        }
        files = []interface{}{ stdin, stdout, stderr }
    }

    // Start process
    proc, err := process.StartProcess(programPath, args, &process.ProcAttr{
        Dir: session.SessionDir,
        Env: append(os.Environ(), ExtraEnviron...),
        Files: files,
        Sys: &forkexec.SysProcAttr{
            Rlimit: rlimit,
        },
    })
    if err != nil {
        return nil, err
    }
    // Collect process info
    pinfo := ProcessInfo{}
    pinfo.Process = proc
    pinfo.Pid = proc.Pid
    return &pinfo, nil
}

// 构建判题程序的命令行参数
func getSpecialJudgeArgs(session *JudgeSession, rst *commonStructs.TestCaseResult) []string {
    tci, err := filepath.Abs(path.Join(session.ConfigDir, rst.Input))
    if err == nil {
        tci = path.Join(session.ConfigDir, rst.Input)
    }
    tco, err := filepath.Abs(path.Join(session.ConfigDir, rst.Output))
    if err == nil {
        tco = path.Join(session.ConfigDir, rst.Output)
    }
    po, err := filepath.Abs(path.Join(session.SessionDir, rst.ProgramOut))
    if err == nil {
        po = path.Join(session.SessionDir, rst.ProgramOut)
    }
    jr, err := filepath.Abs(path.Join(session.SessionDir, rst.CheckerReport))
    if err == nil {
        jr = path.Join(session.SessionDir, rst.CheckerReport)
    }
    // Run Judger (Testlib compatible)
    // -appes prop will allow checker export result as xml.
    // ./checker <input-file> <output-file> <answer-file> <report-file> [-appes]
    args := []string{
        session.JudgeConfig.SpecialJudge.Checker,       // 程序
        tci,                                            // 输入文件流
        po,                                             // 选手输出流
        tco,                                            // 参考输出流
        jr,                                             // report
    }
    if session.JudgeConfig.SpecialJudge.UseTestlib {
        args = append(args, "-appes")
    }
    return args
}