package executor

import (
	"fmt"
	"github.com/LanceLRQ/deer-executor/v2/common/provider"
	commonStructs "github.com/LanceLRQ/deer-executor/v2/common/structs"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path"
	"syscall"
)

// Max find max between x & y
func Max(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

// Max32 find max between a & b
func Max32(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// 文件读写(有重试次数，checker专用)
func readFileWithTry(filePath string, name string, tryOnFailed int) ([]byte, string, error) {
	errCnt, errText := 0, ""
	var err error
	for errCnt < tryOnFailed {
		fp, err := os.OpenFile(filePath, os.O_RDONLY|syscall.O_NONBLOCK, 0)
		if err != nil {
			errText = err.Error()
			errCnt++
			continue
		}
		data, err := ioutil.ReadAll(fp)
		if err != nil {
			_ = fp.Close()
			errText = fmt.Sprintf("Read file(%s) i/o error: %s", name, err.Error())
			errCnt++
			continue
		}
		_ = fp.Close()
		return data, errText, nil
	}
	return nil, errText, err
}

// CheckRequireFilesExists 检查配置文件里的所有文件是否存在
func CheckRequireFilesExists(config *commonStructs.JudgeConfiguration, configDir string) error {
	var err error
	// 检查特判程序是否存在
	if config.SpecialJudge.Mode != 0 {
		_, err = os.Stat(path.Join(configDir, config.SpecialJudge.Checker))
		if os.IsNotExist(err) {
			return errors.Errorf("special judge checker file (%s) not exists", config.SpecialJudge.Checker)
		}
	}
	// 检查每个测试数据里的文件是否存在
	// 新版判题机要求无论有没有数据，都要有对应的输入输出文件。
	// 但Testlib模式例外，因为数据是由generator自动生成的。
	for i := 0; i < len(config.TestCases); i++ {
		tcase := config.TestCases[i]
		if !tcase.Enabled || tcase.UseGenerator {
			continue
		}
		_, err = os.Stat(path.Join(configDir, tcase.Input))
		if os.IsNotExist(err) {
			return errors.Errorf("test case (%s) input file (%s) not exists", tcase.Handle, tcase.Input)
		}
		_, err = os.Stat(path.Join(configDir, tcase.Output))
		if os.IsNotExist(err) {
			return errors.Errorf("test case (%s) output file (%s) not exists", tcase.Handle, tcase.Output)
		}
	}
	return nil
}

// GetOrCreateBinaryRoot 获取二进制文件的目录
func GetOrCreateBinaryRoot(config *commonStructs.JudgeConfiguration) (string, error) {
	binRoot := path.Join(config.ConfigDir, "bin")
	_, err := os.Stat(binRoot)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(binRoot, 0775)
		if err != nil {
			return "", errors.Errorf("cannot create binary work directory: %s", err.Error())
		}
	}
	return binRoot, nil
}

// CompileSpecialJudgeCodeFile 普通特殊评测的编译方法
func CompileSpecialJudgeCodeFile(source, name, binRoot, configDir, libraryDir, lang string) (string, error) {
	genCodeFile := path.Join(configDir, source)
	compileTarget := path.Join(binRoot, name)
	_, err := os.Stat(genCodeFile)
	if err != nil && os.IsNotExist(err) {
		return compileTarget, errors.Errorf("checker source code file not exists")
	}
	var ok bool
	var ceinfo string
	switch lang {
	case "c", "gcc", "gnu-c":
		compiler := provider.NewGnucppCompileProvider()
		ok, ceinfo = compiler.ManualCompile(genCodeFile, compileTarget, []string{libraryDir})
	case "go", "golang":
		compiler := provider.NewGolangCompileProvider()
		ok, ceinfo = compiler.ManualCompile(genCodeFile, compileTarget)
	case "cpp", "gcc-cpp", "gcpp", "g++", "":
		compiler := provider.NewGnucppCompileProvider()
		ok, ceinfo = compiler.ManualCompile(genCodeFile, compileTarget, []string{libraryDir})
	default:
		return compileTarget, errors.Errorf("checker must be written by c/c++/golang")
	}
	if ok {
		return compileTarget, nil
	}
	return compileTarget, errors.Errorf("compile error: %s", ceinfo)
}
