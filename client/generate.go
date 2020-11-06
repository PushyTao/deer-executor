package client

import (
    "fmt"
    "github.com/LanceLRQ/deer-common/constants"
    "github.com/LanceLRQ/deer-common/provider"
    commonStructs "github.com/LanceLRQ/deer-common/structs"
    "github.com/LanceLRQ/deer-common/utils"
    "github.com/LanceLRQ/deer-executor/executor"
    "github.com/urfave/cli/v2"
    "log"
    "os"
)

var AppMakeSubCommands = cli.Commands{
    {
        Name:   "config",
        Action: MakeConfigFile,
        Flags: []cli.Flag{
            &cli.StringFlag{
                Name:    "output",
                Aliases: []string{"out"},
                Value:   "",
                Usage:   "output config file",
            },
        },
    },
    {
        Name:   "compiler",
        Action: MakeCompileConfigFile,
        Flags: []cli.Flag{
            &cli.StringFlag{
                Name:    "output",
                Aliases: []string{"out"},
                Value:   "",
                Usage:   "output config file",
            },
        },
    }, {
        Name:   "jit_memory",
        Action: MakeJITMemoryConfigFile,
        Flags: []cli.Flag{
            &cli.StringFlag{
                Name:    "output",
                Aliases: []string{"out"},
                Value:   "",
                Usage:   "output config file",
            },
        },
    },
}

// 生成评测配置文件
func MakeConfigFile(c *cli.Context) error {
    config, _ := executor.NewSession("")
    config.JudgeConfig.TestCases = []commonStructs.TestCase{
        {
            Handle: "1",
            Input:  "",
            Output: "",
        },
    }
    config.JudgeConfig.Problem.Sample = []commonStructs.ProblemIOSample{
        {
            Input:  "",
            Output: "",
        },
    }
    config.JudgeConfig.Limitation = make(map[string]commonStructs.JudgeResourceLimit)
    config.JudgeConfig.Limitation["gcc"] = commonStructs.JudgeResourceLimit{
        TimeLimit:     config.JudgeConfig.TimeLimit,
        MemoryLimit:   config.JudgeConfig.MemoryLimit,
        RealTimeLimit: config.JudgeConfig.RealTimeLimit,
        FileSizeLimit: config.JudgeConfig.FileSizeLimit,
    }
    output := c.String("output")
    if output != "" {
        s, err := os.Stat(output)
        if s != nil || os.IsExist(err) {
            log.Fatal("output file exists")
            return nil
        }
        fp, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE, 0644)
        if err != nil {
            log.Fatalf("open output file error: %s\n", err.Error())
            return nil
        }
        defer fp.Close()
        _, err = fp.WriteString(utils.ObjectToJSONStringFormatted(config.JudgeConfig))
        if err != nil {
            return err
        }
    } else {
        fmt.Println(utils.ObjectToJSONStringFormatted(config.JudgeConfig))
    }
    return nil
}

// 生成编译器配置(程序使用)
func MakeCompileConfigFile(c *cli.Context) error {
    config := provider.CompileCommands
    output := c.String("output")
    if output == "" {
        output = "./compilers.json"
    }
    s, err := os.Stat(output)
    if s != nil || os.IsExist(err) {
        log.Fatal("output file exists")
        return nil
    }
    fp, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE, 0644)
    if err != nil {
        log.Fatalf("open output file error: %s\n", err.Error())
        return nil
    }
    defer fp.Close()
    _, err = fp.WriteString(utils.ObjectToJSONStringFormatted(config))
    if err != nil {
        return err
    }
    return nil
}

// 生成JIT内存宽限配置
func MakeJITMemoryConfigFile(c *cli.Context) error {
    config := constants.MemorySizeForJIT
    output := c.String("output")
    if output == "" {
        output = "./jit_memory.json"
    }
    s, err := os.Stat(output)
    if s != nil || os.IsExist(err) {
        log.Fatal("output file exists")
        return nil
    }
    fmt.Println(output)
    fp, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE, 0644)
    if err != nil {
        log.Fatalf("open output file error: %s\n", err.Error())
        return nil
    }
    defer fp.Close()
    _, err = fp.WriteString(utils.ObjectToJSONStringFormatted(config))
    if err != nil {
        return err
    }
    return nil
}
